package runtime

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/surik/k8s-image-warden/pkg/proto"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type RemoteService struct {
	client cri.RuntimeServiceClient
}

func NewRuntimeService(runtimeEndpoint string) (*RemoteService, error) {
	conn, err := newConnection(runtimeEndpoint)
	if err != nil {
		return nil, err
	}

	client := cri.NewRuntimeServiceClient(conn)

	return &RemoteService{client: client}, nil
}

func (svc RemoteService) PrintEvents(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	stream, err := svc.client.GetContainerEvents(ctx, &cri.GetEventsRequest{})

	if err != nil {
		return err
	}

	for {
		event, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return err
		}

		if err != nil {
			return err
		}

		log.Printf("event: %+v\n", event)
	}
}

func (svc RemoteService) Version(ctx context.Context) (*proto.Version, error) {
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	resp, err := svc.client.Version(ctx, &cri.VersionRequest{})
	if err != nil {
		return nil, err
	}

	return &proto.Version{
		Version:           resp.Version,
		RuntimeName:       resp.RuntimeName,
		RuntimeVersion:    resp.RuntimeVersion,
		RuntimeApiVersion: resp.RuntimeApiVersion,
	}, nil
}
