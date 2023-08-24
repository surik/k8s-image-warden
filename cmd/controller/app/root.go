package app

import (
	"context"
	"log"
	"path"
	"syscall"

	"github.com/spf13/cobra"
	k8simagewarden "github.com/surik/k8s-image-warden"
	"github.com/surik/k8s-image-warden/pkg/controller"
	"github.com/surik/k8s-image-warden/pkg/engine"
	"github.com/surik/k8s-image-warden/pkg/repo"
	"github.com/surik/k8s-image-warden/pkg/signal"
	"github.com/surik/k8s-image-warden/pkg/webhook"
)

const grpcListeningEndpointFlag = "grpc-listening-endpoint"
const webhookListeningEndpointFlag = "webhook-listening-endpoint"
const webhookCertFileFlag = "webhook-cert-file"
const webhookKeyFileFlag = "webhook-key-file"
const rulesFileFlag = "rules-file"
const storeFileFlag = "store-file"
const reportIntervalFlag = "agent-report-interval"
const retentionFlag = "retention"

var rootCmd = &cobra.Command{
	Use:     "k8s-image-warder-controller",
	Short:   "Extended image policies controller",
	Version: k8simagewarden.Version,
	Run: func(cmd *cobra.Command, args []string) {
		grpcListeningEndpoint, err := cmd.Flags().GetString(grpcListeningEndpointFlag)
		if err != nil {
			log.Fatal(err)
		}

		webhookListeningEndpoint, err := cmd.Flags().GetString(webhookListeningEndpointFlag)
		if err != nil {
			log.Fatal(err)
		}

		rulesFile, err := cmd.Flags().GetString(rulesFileFlag)
		if err != nil {
			log.Fatal(err)
		}

		storeFile, err := cmd.Flags().GetString(storeFileFlag)
		if err != nil {
			log.Fatal(err)
		}

		reportInterval, err := cmd.Flags().GetUint16(reportIntervalFlag)
		if err != nil {
			log.Fatal(err)
		}

		retentionDays, err := cmd.Flags().GetUint16(retentionFlag)
		if err != nil {
			log.Fatal(err)
		}

		repo, err := repo.NewRepo(storeFile, reportInterval, retentionDays)
		if err != nil {
			log.Fatal(err)
		}

		repo.RunStaleRecordsCleaner()

		engine, err := engine.NewEngineFromFile(repo, engine.NewImageInspector(), rulesFile)
		if err != nil {
			log.Fatal(err)
		}

		controller := controller.NewController(grpcListeningEndpoint, repo, engine)

		go func() {
			_ = controller.Run()
		}()

		certFile, err := cmd.Flags().GetString(webhookCertFileFlag)
		if err != nil {
			log.Fatal(err)
		}

		keyFile, err := cmd.Flags().GetString(webhookKeyFileFlag)
		if err != nil {
			log.Fatal(err)
		}

		webhookServer, err := webhook.NewWebhookServer(webhookListeningEndpoint, certFile, keyFile)
		if err != nil {
			log.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		go webhookServer.Run(ctx, engine)

		signal.WaitForSignals(func() {
			controller.Stop()
			repo.StopStaleRecordsCleaner()
			webhookServer.Stop()
			defer cancel()
		}, signal.DefaultWaitTimeout, syscall.SIGINT, syscall.SIGTERM)
	},
}

func Execute() {
	flags := rootCmd.PersistentFlags()
	flags.String(grpcListeningEndpointFlag, ":5000", "The GRPC listening endpoint of image-warden controller")
	flags.String(webhookListeningEndpointFlag, ":8443", "The admission HTTPS webhook listening endpoint of image-warden controller")
	flags.String(webhookCertFileFlag, path.Join("certs", "tls.crt"), "The path to TLS certificate for webhook")
	flags.String(webhookKeyFileFlag, path.Join("certs", "tls.key"), "The path to TLS key file for webhook")
	flags.String(rulesFileFlag, path.Join("config", "rules.yaml"), "The path to YAML file that contains engine rules")
	flags.String(storeFileFlag, path.Join("store.db"), "The path to SQLite storage file")
	flags.Uint16(reportIntervalFlag, k8simagewarden.DefaultFetchInterval,
		"What is agent reporting interval, in seconds. Keep it the same as agent fetch-interval")
	flags.Uint16(retentionFlag, k8simagewarden.DefaultRetention,
		"For how long controller should keep reports in days, 0 means forever")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Whoops. There was an error while executing your CLI '%s'", err)
	}
}
