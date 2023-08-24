package testing

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	k8simagewarden "github.com/surik/k8s-image-warden"
	"github.com/surik/k8s-image-warden/pkg/controller"
	"github.com/surik/k8s-image-warden/pkg/proto"
	"github.com/surik/k8s-image-warden/pkg/repo"
)

const (
	Node1         = "docker-desktop-1"
	Node2         = "docker-desktop-2"
	Node3         = "docker-desktop-3"
	ImageSha1     = "sha256:c61f3549af26394e701f70e05f3bb7ec7df3815d40b8dce0ea3df7d7f35d0f9a"
	ImageSha2     = "sha256:d1aabb73d2339c5ebaa3681de2e9d9c18d57485045a4e311d9f8004bec208d12"
	ImageSha3     = "sha256:32c4087dcb4b17020cd7882c7b4d8011a1dfc39621aa7334bb2fb1dffb976295"
	FsMountpoint1 = "/var/lib/test/1"
	FsMountpoint2 = "/var/lib/test/2"
	Digest1       = "sha256:9f76a008888da28c6490bedf7bdaa919bac9b2be827afd58d6eb1b916eaa5911"
	Digest2       = "sha256:9f76a008888da28c6490bedf7bdaa919bac9b2be827afd58d6eb1b916eaa5922"
	Digest3       = "sha256:9f76a008888da28c6490bedf7bdaa919bac9b2be827afd58d6eb1b916eaa5933"

	storeFile = "store.db"
)

func NewTestRepo(t *testing.T) *repo.Repo {
	dir := t.TempDir()
	file := path.Join(dir, storeFile)
	repo, err := repo.NewRepo(file, k8simagewarden.DefaultFetchInterval, k8simagewarden.DefaultRetention)
	require.NoError(t, err)
	return repo
}

func Report1() *proto.ReportRequest {
	return &proto.ReportRequest{
		RuntimeInfo: &proto.RuntimeInfo{
			Podname:      "pod-agent-1",
			Nodename:     Node1,
			AgentVersion: k8simagewarden.Version,
			RuntimeVersion: &proto.Version{
				Version:           "1.0.0",
				RuntimeName:       "test",
				RuntimeVersion:    "1.0.0",
				RuntimeApiVersion: "test",
			},
		},
		FilesystemUsageList: &proto.FilesystemUsageList{
			ImageFilesystems: []*proto.FilesystemUsage{
				{
					Timestamp: time.Now().UnixNano(),
					FsId: &proto.FilesystemIdentifier{
						Mountpoint: FsMountpoint1,
					},
					InodesUsed: &proto.UInt64Value{Value: 1},
					UsedBytes:  &proto.UInt64Value{Value: 1000},
				},
			},
		},
		ImageList: &proto.ImageList{
			Images: []*proto.Image{
				{
					Id:          ImageSha1,
					RepoTags:    []string{"debian:latest"},
					RepoDigests: []string{"debian@sha256:9f76a008888da28c6490bedf7bdaa919bac9b2be827afd58d6eb1b916e1e5918"},
					Size:        1000,
				},
			},
		},
	}
}

func Report2() *proto.ReportRequest {
	return &proto.ReportRequest{
		RuntimeInfo: &proto.RuntimeInfo{
			Podname:      "pod-agent-2",
			Nodename:     Node2,
			AgentVersion: k8simagewarden.Version,
			RuntimeVersion: &proto.Version{
				Version:           "1.0.0",
				RuntimeName:       "test",
				RuntimeVersion:    "1.0.0",
				RuntimeApiVersion: "test",
			},
		},
		FilesystemUsageList: &proto.FilesystemUsageList{
			ImageFilesystems: []*proto.FilesystemUsage{
				{
					FsId: &proto.FilesystemIdentifier{
						Mountpoint: FsMountpoint2,
					},
				},
			},
		},
		ImageList: &proto.ImageList{
			Images: []*proto.Image{
				{
					Id:          ImageSha2,
					RepoTags:    []string{"k8s-image-warden-agent:latest"},
					RepoDigests: []string{},
					Size:        2000,
				},
			},
		},
	}
}

// with the rolling tag
func Report3() *proto.ReportRequest {
	return &proto.ReportRequest{
		RuntimeInfo: &proto.RuntimeInfo{
			Podname:      "pod-agent-2",
			Nodename:     Node2,
			AgentVersion: k8simagewarden.Version,
			RuntimeVersion: &proto.Version{
				Version:           "1.0.0",
				RuntimeName:       "test",
				RuntimeVersion:    "1.0.0",
				RuntimeApiVersion: "test",
			},
		},
		FilesystemUsageList: &proto.FilesystemUsageList{
			ImageFilesystems: []*proto.FilesystemUsage{},
		},
		ImageList: &proto.ImageList{
			Images: []*proto.Image{
				{
					Id:          ImageSha1,
					RepoTags:    []string{"k8s-image-warden-controller:latest"},
					RepoDigests: []string{Digest1},
					Size:        1000,
				},
				{
					Id:          ImageSha2,
					RepoTags:    []string{"k8s-image-warden-agent:latest"},
					RepoDigests: []string{Digest2},
					Size:        2000,
				},
				{
					Id:          ImageSha3,
					RepoTags:    []string{"k8s-image-warden-agent:latest"},
					RepoDigests: []string{Digest3},
					Size:        2020,
				},
			},
		},
	}
}

func PrepareRollingTags(r *repo.Repo) error {
	node, fsUsage, images := controller.ConvertReportToRepo(Report3())
	return r.StoreReport(node, fsUsage, images)
}
