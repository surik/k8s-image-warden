package agent

import (
	"context"
	"log"
	"time"

	k8simagewarden "github.com/surik/k8s-image-warden"
	"github.com/surik/k8s-image-warden/pkg/proto"
	"github.com/surik/k8s-image-warden/pkg/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Agent struct {
	RuntimeService     runtime.RemoteService
	ImageService       runtime.ImageService
	ControllerService  proto.ControllerServiceClient
	criEndpoint        string
	controllerEndpoint string
	podname            string
	nodename           string
	fetchInterval      uint16
}

func NewAgent(ctx context.Context, criEndpoint, controllerEndpoint, podname, node string, fetchInterval uint16) (*Agent, error) {
	runtimeService, err := runtime.NewRuntimeService(criEndpoint)
	if err != nil {
		return nil, err
	}

	imageService, err := runtime.NewImageService(criEndpoint)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.DialContext(ctx, controllerEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	controllerClient := proto.NewControllerServiceClient(conn)

	return &Agent{
		RuntimeService:     *runtimeService,
		ImageService:       *imageService,
		ControllerService:  controllerClient,
		criEndpoint:        criEndpoint,
		controllerEndpoint: controllerEndpoint,
		podname:            podname,
		nodename:           node,
		fetchInterval:      fetchInterval,
	}, nil
}

func (agent Agent) Run(ctx context.Context) {
	log.Println("Agent is up and running.")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(agent.fetchInterval) * time.Second):
			if err := agent.report(ctx); err != nil {
				log.Println(err)
			}
		}
	}
}

func (agent Agent) report(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	version, err := agent.RuntimeService.Version(ctx)
	if err != nil {
		return err
	}

	images, err := agent.ImageService.ListImages(ctx)
	if err != nil {
		return err
	}

	fsInfo, err := agent.ImageService.GetFsInfo(ctx)
	if err != nil {
		return err
	}

	report := proto.ReportRequest{
		RuntimeInfo: &proto.RuntimeInfo{
			Podname:        agent.podname,
			Nodename:       agent.nodename,
			AgentVersion:   k8simagewarden.Version,
			RuntimeVersion: version,
		},

		FilesystemUsageList: &proto.FilesystemUsageList{
			ImageFilesystems: fsInfo,
		},
		ImageList: &proto.ImageList{
			Images: images,
		},
	}

	_, err = agent.ControllerService.Report(ctx, &report)
	if err != nil {
		return err
	}

	return nil
}
