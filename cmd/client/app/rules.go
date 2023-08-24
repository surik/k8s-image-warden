package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/surik/k8s-image-warden/pkg/engine"
	"github.com/surik/k8s-image-warden/pkg/proto"
	"gopkg.in/yaml.v3"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List rules known by controller",
	Run:   rules,
}

func rules(cmd *cobra.Command, args []string) {
	controllerClient, err := connect(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer controllerClient.Stop()

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
	defer cancel()

	resp, err := controllerClient.GetRules(ctx, &proto.GetRulesRequest{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// just to validate that response is correct
	var rules []engine.Rule
	err = yaml.Unmarshal(resp.RawRules, &rules)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(string(resp.RawRules))
}
