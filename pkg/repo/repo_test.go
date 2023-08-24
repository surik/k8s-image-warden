package repo_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/surik/k8s-image-warden/pkg/controller"
	helpers "github.com/surik/k8s-image-warden/pkg/repo/testing"
)

func TestRepo(t *testing.T) {
	repo := helpers.NewTestRepo(t)

	node, fsUsage, images := controller.ConvertReportToRepo(helpers.Report1())
	err := repo.StoreReport(node, fsUsage, images)
	require.NoError(t, err)

	node, fsUsage, images = controller.ConvertReportToRepo(helpers.Report2())
	err = repo.StoreReport(node, fsUsage, images)
	require.NoError(t, err)

	// get report for all two nodes
	report, err := repo.GetReportForNode("", true)
	require.NoError(t, err)
	require.Len(t, report, 2)
	require.Equal(t, helpers.Node1, report[0].Nodename)
	require.Equal(t, helpers.Node2, report[1].Nodename)
	require.Len(t, report[0].Images, 1)
	require.Equal(t, helpers.ImageSha1, report[0].Images[0].ID)
	require.Len(t, report[1].Images, 1)
	require.Equal(t, helpers.ImageSha2, report[1].Images[0].ID)

	// get report for Node1
	report, err = repo.GetReportForNode(helpers.Node2, true)
	require.NoError(t, err)
	require.Len(t, report, 1)
	require.Equal(t, helpers.Node2, report[0].Nodename)
	require.Len(t, report[0].Images, 1)
	require.Equal(t, helpers.ImageSha2, report[0].Images[0].ID)
	require.Len(t, report[0].ImageFilesystems, 1)
	require.Equal(t, helpers.FsMountpoint2, report[0].ImageFilesystems[0].Mountpoint)

	// get report for Node1 pretending that the node disappeared
	repo.SetReportInterval(0)
	report, err = repo.GetReportForNode(helpers.Node2, false)
	require.NoError(t, err)
	require.Len(t, report, 0)

	// Test Clean Stale Records job
	repo.SetRetention(1)
	time.Sleep(1 * time.Second)
	err = repo.CleanStaleRecords(0)
	require.NoError(t, err)
	report, err = repo.GetReportForNode("", true)
	require.NoError(t, err)
	require.Len(t, report, 0)
}
