package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/surik/k8s-image-warden/pkg/proto"
)

var mutateCmd = &cobra.Command{
	Use:   "mutate",
	Short: "Mutate image on the controller",
	Run:   mutate,
}

func mutate(cmd *cobra.Command, args []string) {
	controllerClient, err := connect(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer controllerClient.Stop()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "image reference is required")
		os.Exit(1)
	}

	image := args[0]

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
	defer cancel()

	resp, err := controllerClient.Mutate(ctx, &proto.MutateRequest{Image: image})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(resp.Rules) > 0 {
		fmt.Printf("'%s' is mutated to '%s' after applying the rules: %s\n",
			image, resp.Image, strings.Join(resp.Rules, ","))
	} else {
		fmt.Printf("No mutation rules for '%s'\n", image)
	}
}
