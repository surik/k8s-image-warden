package runtime

import (
	"context"

	"github.com/surik/k8s-image-warden/pkg/proto"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type ImageService struct {
	client cri.ImageServiceClient
}

func NewImageService(runtimeEndpoint string) (*ImageService, error) {
	conn, err := newConnection(runtimeEndpoint)
	if err != nil {
		return nil, err
	}

	client := cri.NewImageServiceClient(conn)

	return &ImageService{client: client}, nil
}

func (svc ImageService) GetFsInfo(ctx context.Context) ([]*proto.FilesystemUsage, error) {
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	resp, err := svc.client.ImageFsInfo(ctx, &cri.ImageFsInfoRequest{})
	if err != nil {
		return nil, err
	}

	info := make([]*proto.FilesystemUsage, 0, len(resp.ImageFilesystems))
	for _, fs := range resp.ImageFilesystems {
		info = append(info, &proto.FilesystemUsage{
			// NOTE: cri-dockerd has a bug https://github.com/Mirantis/cri-dockerd/pull/218
			Timestamp:  fs.Timestamp,
			FsId:       &proto.FilesystemIdentifier{Mountpoint: fs.FsId.Mountpoint},
			UsedBytes:  &proto.UInt64Value{Value: fs.UsedBytes.Value},
			InodesUsed: &proto.UInt64Value{Value: fs.InodesUsed.Value},
		})
	}

	return info, nil
}

func (svc ImageService) ListImages(ctx context.Context) ([]*proto.Image, error) {
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	resp, err := svc.client.ListImages(ctx, &cri.ListImagesRequest{})
	if err != nil {
		return nil, err
	}

	images := make([]*proto.Image, 0, len(resp.Images))
	for _, image := range resp.Images {
		uid := int64(0)
		if image.Uid != nil {
			uid = image.Uid.Value
		}

		img := &proto.Image{
			Id:          image.Id,
			RepoTags:    image.RepoTags,
			RepoDigests: image.RepoDigests,
			Size:        image.Size_,
			Uid:         &proto.Int64Value{Value: uid},
			Username:    image.Username,
			Pinned:      image.Pinned,
		}

		if image.Spec != nil {
			spec := proto.ImageSpec{
				Image:       image.Spec.Image,
				Annotations: image.Spec.Annotations,
			}
			img.Spec = &spec
		}

		images = append(images, img)
	}

	return images, nil
}
