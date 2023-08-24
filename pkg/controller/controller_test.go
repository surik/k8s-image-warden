package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/surik/k8s-image-warden/pkg/controller"
	"github.com/surik/k8s-image-warden/pkg/engine"
	"github.com/surik/k8s-image-warden/pkg/proto"
	helpers "github.com/surik/k8s-image-warden/pkg/repo/testing"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

func TestController(t *testing.T) {
	repo := helpers.NewTestRepo(t)

	node, fsUsage, images := controller.ConvertReportToRepo(helpers.Report1())
	err := repo.StoreReport(node, fsUsage, images)
	require.NoError(t, err)

	node, fsUsage, images = controller.ConvertReportToRepo(helpers.Report2())
	err = repo.StoreReport(node, fsUsage, images)
	require.NoError(t, err)

	eng, err := engine.NewEngine(repo, nil, []engine.Rule{
		{
			Name: "default rule",
			ValidationRule: engine.ValidationRule{
				Type:  engine.ValidateTypeLatest,
				Allow: true,
			},
		},
	})
	require.NotNil(t, eng)
	require.NoError(t, err)

	controller := controller.NewController(":0", repo, eng)
	require.NotNil(t, controller)

	responseRules, err := controller.GetRules(context.Background(), &proto.GetRulesRequest{})
	require.NoError(t, err)

	// controller has engine with two rules
	var rules []engine.Rule
	err = yaml.Unmarshal(responseRules.RawRules, &rules)
	require.NoError(t, err)
	require.Len(t, rules, 1)

	response, err := controller.GetReport(context.Background(), &proto.GetReportRequest{})
	require.NoError(t, err)

	// report contains two nodes
	require.Contains(t, maps.Keys(response.Image), helpers.Node1)
	require.Contains(t, maps.Keys(response.Image), helpers.Node2)

	// report contains images on each node
	require.Greater(t, len(response.Image[helpers.Node1].Images), 0)
	require.Greater(t, len(response.Image[helpers.Node2].Images), 0)

	// simulate report from new node
	image := &proto.Image{
		Id:       helpers.ImageSha2,
		RepoTags: []string{"docker.io/nginx:1.25.2"},
	}
	fs := &proto.FilesystemUsage{
		Timestamp: time.Now().UnixNano(),
		FsId: &proto.FilesystemIdentifier{
			Mountpoint: "/var/lib/docker",
		},
	}
	_, err = controller.Report(context.Background(), &proto.ReportRequest{
		RuntimeInfo: &proto.RuntimeInfo{
			Nodename:       helpers.Node3,
			RuntimeVersion: response.Runtime[helpers.Node1].RuntimeVersion,
		},
		ImageList: &proto.ImageList{
			Images: []*proto.Image{image},
		},
		FilesystemUsageList: &proto.FilesystemUsageList{
			ImageFilesystems: []*proto.FilesystemUsage{fs},
		},
	})
	require.NoError(t, err)

	response, err = controller.GetReport(context.Background(), &proto.GetReportRequest{})
	require.NoError(t, err)

	// report contains new node
	require.Contains(t, maps.Keys(response.Runtime), helpers.Node3)

	// report contains images on new node
	require.Greater(t, len(response.Image[helpers.Node3].Images), 0)

	// report contains FS info
	require.Equal(t, "/var/lib/docker", response.FilesystemUsage[helpers.Node3].ImageFilesystems[0].GetFsId().GetMountpoint())
}
