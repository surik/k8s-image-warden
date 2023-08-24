package repo

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Node struct {
	ID                uint   `gorm:"primarykey"`
	Podname           string `gorm:"index"`
	Nodename          string `gorm:"index"`
	LastSeen          time.Time
	AgentVersion      string
	KubeAPIVersion    string
	RuntimeName       string
	RuntimeVersion    string
	RuntimeAPIVersion string
	Images            []ImageReport
	ImageFilesystems  []ImageFilesystemReport
}

type ImageReport struct {
	ID                 string `gorm:"primaryKey"`
	RepoTag            string `gorm:"primaryKey"`
	RepoDigest         string `gorm:"primaryKey"`
	NodeID             uint   `gorm:"primaryKey"`
	Size               uint64
	Username           string
	Image              string
	Annotations        map[string]string `gorm:"serializer:json"`
	UserSpecifiedImage string
	Pinned             bool
	UID                int64
	ReportedAt         time.Time
	SeenInLastReport   bool
}

type ImageFilesystemReport struct {
	Timestamp        time.Time
	Mountpoint       string `gorm:"primaryKey"`
	NodeID           uint   `gorm:"primaryKey"`
	UsedBytes        uint64
	InodesUsed       uint64
	ReportedAt       time.Time
	SeenInLastReport bool
}

type RepoOpts struct {
	ReportInterval  time.Duration
	Retention       time.Duration
	NodeRetention   time.Duration
	CleanerInterval time.Duration
}

type Repo struct {
	db     *gorm.DB
	doneCh chan bool
	opts   RepoOpts
}

func NewRepo(file string, reportIntervalSecs, retentionDays uint16) (*Repo, error) {
	config := gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(sqlite.Open(file), &config)
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Node{}, &ImageReport{}, &ImageFilesystemReport{})
	if err != nil {
		return nil, err
	}

	return &Repo{
		db:     db,
		doneCh: make(chan bool),
		opts: RepoOpts{
			ReportInterval:  time.Duration(reportIntervalSecs) * time.Second,
			Retention:       time.Duration(retentionDays) * time.Hour,
			NodeRetention:   time.Duration(24) * time.Hour,
			CleanerInterval: time.Minute,
		},
	}, nil
}

func (r Repo) RunStaleRecordsCleaner() {
	if r.opts.Retention == 0 {
		return
	}

	go func() {
		for {
			select {
			case <-r.doneCh:
				return
			case <-time.After(r.opts.CleanerInterval):
				err := r.cleanStaleRecords(-r.opts.NodeRetention)
				if err != nil {
					log.Printf("error when deleting stale records %s", err)
				}
			}
		}
	}()
}

func (r Repo) StopStaleRecordsCleaner() {
	if r.opts.Retention == 0 {
		return
	}

	r.doneCh <- true
	log.Println("Repo Stale Records Cleaner was shutdown")
}

func (r Repo) cleanStaleRecords(nodesRetention time.Duration) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// We delete nodes if non seen for given nodesRetention.
		// Image and Filesystem reports are being deleted based on Retention configuration.
		err := r.db.Delete(&Node{}, "last_seen <= ?", time.Now().Add(nodesRetention).UTC()).Error
		if err != nil {
			return err
		}

		reportedAt := time.Now().Add(-r.opts.Retention).UTC()

		err = r.db.Delete(&ImageReport{}, "reported_at <= ?", reportedAt).Error
		if err != nil {
			return err
		}

		err = r.db.Delete(&ImageFilesystemReport{}, "reported_at <= ?", reportedAt).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (r Repo) GetReportForNode(node string, all bool) ([]Node, error) {
	var report []Node

	db := r.db.Model(&Node{})
	if all {
		db = db.Preload("Images").Preload("ImageFilesystems")
	} else {
		db = db.
			Preload("Images", "seen_in_last_report = ?", !all).
			Preload("ImageFilesystems", "seen_in_last_report = ?", !all).
			Where("last_seen >= ?", time.Now().Add(-3*r.opts.ReportInterval).UTC())
	}
	if node != "" {
		db = db.Where("nodename = ?", node)
	}

	if result := db.Find(&report); result.Error != nil {
		return nil, result.Error
	}

	return report, nil
}

func (r Repo) StoreReport(node *Node, fsUsage []ImageFilesystemReport, images []ImageReport) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if result := r.db.Where(node).FirstOrCreate(node); result.Error != nil {
			return result.Error
		}

		result := r.db.Model(&Node{}).Where("id = ?", node.ID).Update("last_seen", time.Now().UTC())
		if result.Error != nil {
			return result.Error
		}

		result = r.db.Model(&ImageFilesystemReport{}).Where("node_id = ?", node.ID).Update("seen_in_last_report", false)
		if result.Error != nil {
			return result.Error
		}

		result = r.db.Model(&ImageReport{}).Where("node_id = ?", node.ID).Update("seen_in_last_report", false)
		if result.Error != nil {
			return result.Error
		}

		fsRows := make([]ImageFilesystemReport, len(fsUsage))
		for i := range fsUsage {
			fsRows[i] = fsUsage[i]
			fsRows[i].NodeID = node.ID
			fsRows[i].SeenInLastReport = true
			fsRows[i].ReportedAt = time.Now().UTC()
		}

		if len(fsRows) > 0 {
			result = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&fsRows)
			if result.Error != nil {
				return result.Error
			}
		}

		imageRows := make([]ImageReport, len(images))
		for i := range images {
			imageRows[i] = images[i]
			imageRows[i].NodeID = node.ID
			imageRows[i].SeenInLastReport = true
			imageRows[i].ReportedAt = time.Now().UTC()
		}

		if len(imageRows) > 0 {
			result = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&imageRows)
			if result.Error != nil {
				return result.Error
			}
		}

		return nil
	})
}

func (r Repo) GetDigestsByNameAndAfter(name string, after time.Time) ([]string, error) {
	var ids []string
	result := r.db.Table("image_reports").Distinct("repo_digest").
		Where("repo_tag = ? and reported_at > ?", name, after).Scan(&ids)

	return ids, result.Error
}

func (r Repo) GetIDsByNameAndAfter(name string, after time.Time) ([]string, error) {
	var ids []string
	result := r.db.Table("image_reports").Distinct("id").
		Where("repo_tag = ? and reported_at > ?", name, after).Scan(&ids)

	return ids, result.Error
}
