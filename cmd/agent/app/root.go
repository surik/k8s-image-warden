package app

import (
	"context"
	"log"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	k8simagewarden "github.com/surik/k8s-image-warden"
	"github.com/surik/k8s-image-warden/pkg/agent"
	"github.com/surik/k8s-image-warden/pkg/signal"
)

const criEndpointFlag = "container-runtime-endpoint"
const controllerEndpointFlag = "controller-endpoint"
const fetchIntervalFlag = "cri-fetch-interval"

var rootCmd = &cobra.Command{
	Use:     "k8s-image-warder-agent",
	Short:   "image warder agent to interact with CRI on node",
	Version: k8simagewarden.Version,
	Run: func(cmd *cobra.Command, args []string) {
		criEndpoint, err := cmd.Flags().GetString(criEndpointFlag)
		if err != nil {
			log.Fatal(err)
		}

		ctrlEndpoint, err := cmd.Flags().GetString(controllerEndpointFlag)
		if err != nil {
			log.Fatal(err)
		}

		fetchInterval, err := cmd.Flags().GetUint16(fetchIntervalFlag)
		if err != nil {
			log.Fatal(err)
		}

		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		}

		nodeName := os.Getenv("NODE_NAME")
		if nodeName == "" {
			log.Fatal("NODE_NAME should be set and not empty")
		}

		ctx, cancel := context.WithCancel(context.Background())

		agent, err := agent.NewAgent(ctx, criEndpoint, ctrlEndpoint, hostname, nodeName, fetchInterval)
		if err != nil {
			log.Fatal(err)
		}

		go agent.Run(ctx)

		signal.WaitForSignals(func() { cancel() }, signal.DefaultWaitTimeout, syscall.SIGINT, syscall.SIGTERM)
	},
}

func Execute() {
	rootCmd.PersistentFlags().String(criEndpointFlag, "/var/run/cri-dockerd.sock", "The endpoint of container runtime service")
	rootCmd.PersistentFlags().String(controllerEndpointFlag, "k8s-image-warden-controller:5000", "The endpoint of image-warden controller")
	rootCmd.PersistentFlags().Uint16(fetchIntervalFlag, k8simagewarden.DefaultFetchInterval,
		"How frequently to fetch info from this CRI, in seconds")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Whoops. There was an error while executing your CLI '%s'", err)
	}
}
