package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/surik/k8s-image-warden/pkg/proto"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate image on the controller",
	Run:   validate,
}

func validate(cmd *cobra.Command, args []string) {
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

	resp, err := controllerClient.Validate(ctx, &proto.ValidateRequest{Image: image})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if resp.Valid {
		fmt.Printf("'%s' is valid\n", image)
	} else {
		fmt.Printf("'%s' rejected by rule '%s'\n", image, resp.Rule)
	}
}
