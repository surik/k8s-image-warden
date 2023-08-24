package engine

import (
	"context"
	"time"

	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

type ImageInspector interface {
	GetDigest(context.Context, string) (string, error)
}

type imageInspector struct {
	sys *types.SystemContext
}

func NewImageInspector() *imageInspector {
	return &imageInspector{
		sys: &types.SystemContext{},
	}
}

func (i *imageInspector) GetDigest(ctx context.Context, name string) (string, error) {
	ref, err := alltransports.ParseImageName(name)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	src, err := ref.NewImageSource(ctx, i.sys)
	if err != nil {
		return "", err
	}
	defer src.Close()

	raw, _, err := src.GetManifest(ctx, nil)
	if err != nil {
		return "", err
	}

	digest, err := manifest.Digest(raw)
	if err != nil {
		return "", err
	}

	return digest.String(), nil
}
