package repo

import "time"

func (r *Repo) SetReportInterval(interval uint16) {
	r.opts.ReportInterval = time.Duration(interval) * time.Second
}

func (r *Repo) SetRetention(retentionInSecs uint16) {
	r.opts.Retention = time.Duration(retentionInSecs) * time.Second
}

func (r *Repo) CleanStaleRecords(d time.Duration) error {
	return r.cleanStaleRecords(d)
}
