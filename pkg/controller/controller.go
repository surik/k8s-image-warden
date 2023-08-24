package controller

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/surik/k8s-image-warden/pkg/engine"
	"github.com/surik/k8s-image-warden/pkg/proto"
	"github.com/surik/k8s-image-warden/pkg/repo"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

type Controller struct {
	proto.ControllerServiceServer
	grpcServer *grpc.Server
	listener   net.Listener
	engine     *engine.Engine
	repo       *repo.Repo
}

func NewController(endpoint string, repo *repo.Repo, engine *engine.Engine) *Controller {
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	log.Println("Listening on :" + listener.Addr().String())

	ctrl := Controller{
		grpcServer: grpc.NewServer(),
		listener:   listener,
		engine:     engine,
		repo:       repo,
	}
	proto.RegisterControllerServiceServer(ctrl.grpcServer, ctrl)

	return &ctrl
}

func ConvertReportToRepo(report *proto.ReportRequest) (*repo.Node, []repo.ImageFilesystemReport, []repo.ImageReport) {
	node := &repo.Node{
		Podname:           report.RuntimeInfo.Podname,
		Nodename:          report.RuntimeInfo.Nodename,
		AgentVersion:      report.RuntimeInfo.AgentVersion,
		KubeAPIVersion:    report.RuntimeInfo.RuntimeVersion.Version,
		RuntimeName:       report.RuntimeInfo.RuntimeVersion.RuntimeName,
		RuntimeVersion:    report.RuntimeInfo.RuntimeVersion.RuntimeVersion,
		RuntimeAPIVersion: report.RuntimeInfo.RuntimeVersion.RuntimeApiVersion,
	}

	fsUsage := make([]repo.ImageFilesystemReport, len(report.GetFilesystemUsageList().GetImageFilesystems()))
	for i, fs := range report.GetFilesystemUsageList().GetImageFilesystems() {
		fsUsage[i] = repo.ImageFilesystemReport{
			Timestamp:  time.Unix(0, fs.Timestamp),
			Mountpoint: fs.GetFsId().GetMountpoint(),
			UsedBytes:  fs.GetUsedBytes().GetValue(),
			InodesUsed: fs.GetInodesUsed().GetValue(),
		}
	}

	var images []repo.ImageReport
	for _, image := range report.GetImageList().GetImages() {
		for _, tag := range image.RepoTags {
			row := repo.ImageReport{
				ID:                 image.Id,
				RepoTag:            tag,
				Size:               image.Size,
				Username:           image.Username,
				Pinned:             image.Pinned,
				Image:              image.GetSpec().GetImage(),
				UserSpecifiedImage: image.GetSpec().GetUserSpecifiedImage(),
				UID:                image.GetUid().GetValue(),
			}

			if len(image.RepoDigests) > 0 {
				for _, digest := range image.RepoDigests {
					row := row
					row.RepoDigest = digest
					images = append(images, row)
				}
			} else {
				images = append(images, row)
			}
		}
	}

	return node, fsUsage, images
}

func (ctrl Controller) Validate(ctx context.Context, req *proto.ValidateRequest) (*proto.ValidateResponse, error) {
	result, rule := ctrl.engine.Validate(ctx, req.Image)
	return &proto.ValidateResponse{Valid: result, Rule: rule}, nil
}

func (ctrl Controller) Mutate(ctx context.Context, req *proto.MutateRequest) (*proto.MutateResponse, error) {
	newImage, rules := ctrl.engine.Mutate(ctx, req.Image)
	return &proto.MutateResponse{Image: newImage, Rules: rules}, nil
}

func (ctrl Controller) Report(ctx context.Context, report *proto.ReportRequest) (*proto.ReportResponse, error) {
	node, fsUsage, images := ConvertReportToRepo(report)
	err := ctrl.repo.StoreReport(node, fsUsage, images)

	return &proto.ReportResponse{}, err
}

func (ctrl Controller) GetRules(ctx context.Context, req *proto.GetRulesRequest) (*proto.GetRulesResponse, error) {
	rules := ctrl.engine.GetRules()
	bytes, err := yaml.Marshal(rules)
	if err != nil {
		return nil, err
	}

	return &proto.GetRulesResponse{RawRules: bytes}, nil
}

func (ctrl Controller) GetReport(ctx context.Context, req *proto.GetReportRequest) (*proto.GetReportResponse, error) {
	log.Printf("GetReport request for node: '%s'", req.Nodename)

	reports, err := ctrl.repo.GetReportForNode(req.Nodename, req.All)
	if err != nil {
		return nil, err
	}

	runtime := make(map[string]*proto.RuntimeInfo, len(reports))
	images := make(map[string]*proto.ImageList)
	fsUsage := make(map[string]*proto.FilesystemUsageList)

	for _, report := range reports {
		runtime[report.Nodename] = &proto.RuntimeInfo{
			Podname:      report.Podname,
			Nodename:     report.Nodename,
			AgentVersion: report.AgentVersion,
			RuntimeVersion: &proto.Version{
				Version:           report.KubeAPIVersion,
				RuntimeName:       report.RuntimeName,
				RuntimeVersion:    report.RuntimeVersion,
				RuntimeApiVersion: report.RuntimeAPIVersion,
			},
		}

		imageList := make([]*proto.Image, len(report.Images))
		for i, image := range report.Images {
			imageList[i] = &proto.Image{
				Id:          image.ID,
				RepoTags:    []string{image.RepoTag},
				RepoDigests: []string{image.RepoDigest},
				Size:        image.Size,
			}
		}
		images[report.Nodename] = &proto.ImageList{
			Images: imageList,
		}

		fsList := make([]*proto.FilesystemUsage, len(report.ImageFilesystems))
		for i, fs := range report.ImageFilesystems {
			fsList[i] = &proto.FilesystemUsage{
				Timestamp: fs.Timestamp.UnixNano(),
				FsId: &proto.FilesystemIdentifier{
					Mountpoint: fs.Mountpoint,
				},
				UsedBytes: &proto.UInt64Value{
					Value: fs.UsedBytes,
				},
				InodesUsed: &proto.UInt64Value{
					Value: fs.InodesUsed,
				},
			}
		}
		fsUsage[report.Nodename] = &proto.FilesystemUsageList{
			ImageFilesystems: fsList,
		}
	}

	return &proto.GetReportResponse{
		Runtime:         runtime,
		FilesystemUsage: fsUsage,
		Image:           images,
	}, nil
}

func (ctrl Controller) Run() error {
	if err := ctrl.grpcServer.Serve(ctrl.listener); err != nil {
		return err
	}
	return nil
}

func (ctrl Controller) Stop() {
	ctrl.grpcServer.Stop()
}
