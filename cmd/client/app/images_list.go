package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/surik/k8s-image-warden/pkg/proto"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List images known by controller",
	Run:   list,
}

func list(cmd *cobra.Command, args []string) {
	controllerClient, err := connect(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer controllerClient.Stop()

	nodename := ""
	if len(args) > 0 {
		nodename = args[0]
	}

	all, err := cmd.Flags().GetBool(listAllImages)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
	defer cancel()

	resp, err := controllerClient.GetReport(ctx, &proto.GetReportRequest{Nodename: nodename, All: all})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for nodename, runtime := range resp.Runtime {
		images := resp.Image[nodename]
		filesystems := resp.FilesystemUsage[nodename]

		if len(images.Images) == 0 && len(filesystems.ImageFilesystems) == 0 {
			// nothing to show
			continue
		}

		fmt.Printf("Node '%s' with agent version %s and runtime: %s:%s\n",
			runtime.Nodename, runtime.AgentVersion, runtime.RuntimeVersion.RuntimeName, runtime.RuntimeVersion.RuntimeVersion)

		if len(images.Images) > 0 {
			fmt.Println("Images:")
			for _, image := range images.Images {
				fmt.Printf("%s %s %s %d MB\n", image.Id, image.RepoTags, image.RepoDigests, image.Size/1024/1014)
			}
		}

		if len(filesystems.ImageFilesystems) > 0 {
			fmt.Println("File systems:")

			for _, fs := range filesystems.ImageFilesystems {
				fmt.Printf("%s usage %d MB (%d inodes)\n",
					fs.GetFsId().GetMountpoint(), fs.GetUsedBytes().GetValue()/1024/1014, fs.GetInodesUsed().GetValue())
			}
		}

		fmt.Println("")
	}
}
