package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	k8simagewarden "github.com/surik/k8s-image-warden"
	"github.com/surik/k8s-image-warden/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/util/homedir"
)

const kubeconfigPathFlag = "kubeconfig"
const kiwNamespace = "kiw-controller-namespace"
const kiwServiceNameNamespace = "kiw-controller-service-name"
const listAllImages = "all"

var rootCmd = &cobra.Command{
	Use:           "kiwctl",
	Short:         "kiwctl - image warder client to interact with controller running on Kubernetes",
	Version:       k8simagewarden.Version,
	SilenceErrors: true,
	SilenceUsage:  true,
}

var imagesCmd = &cobra.Command{
	Use:     "images",
	Aliases: []string{"image"},
	Short:   "Subcommand to manipulate images",
}

func Execute() {
	imagesCmd.AddCommand(listCmd)
	imagesCmd.AddCommand(validateCmd)
	imagesCmd.AddCommand(mutateCmd)

	rootCmd.AddCommand(imagesCmd)
	rootCmd.AddCommand(rulesCmd)

	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	rootCmd.PersistentFlags().String(kubeconfigPathFlag, kubeconfigPath, "An absolute path to the kubeconfig file")
	rootCmd.PersistentFlags().String(kiwNamespace, "test", "A kubernetes namespace where k8s-image-warden is deployed")
	rootCmd.PersistentFlags().String(kiwServiceNameNamespace, "k8s-image-warden-controller", "A name of k8s-image-warden service")

	listCmd.PersistentFlags().Bool(listAllImages, false, "Include all images known by controller")

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

type ControllerServiceClient interface {
	proto.ControllerServiceClient
	Stop()
}

type controllerServiceClient struct {
	proto.ControllerServiceClient
	conn   *grpc.ClientConn
	stopCh chan struct{}
}

func (c controllerServiceClient) Stop() {
	c.stopCh <- struct{}{}
	c.conn.Close()
}

func connect(cmd *cobra.Command) (ControllerServiceClient, error) {
	// extract flags
	kubeconfig, err := cmd.Flags().GetString(kubeconfigPathFlag)
	if err != nil {
		return nil, err
	}

	kiwNamespace, err := cmd.Flags().GetString(kiwNamespace)
	if err != nil {
		return nil, err
	}

	kiwServiceName, err := cmd.Flags().GetString(kiwServiceNameNamespace)
	if err != nil {
		return nil, err
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// port forward
	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
	defer cancel()

	stopCh := make(chan struct{}, 1)
	port, err := portForward(ctx, config, kiwNamespace, kiwServiceName, stopCh)
	if err != nil {
		return nil, err
	}

	// dial
	endpoint := fmt.Sprintf(":%d", port)
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		stopCh <- struct{}{}
		return nil, err
	}

	proto.NewControllerServiceClient(conn)

	return controllerServiceClient{
		ControllerServiceClient: proto.NewControllerServiceClient(conn),
		stopCh:                  stopCh,
		conn:                    conn,
	}, nil
}

func portForward(ctx context.Context, config *rest.Config, namespace, name string, stopCh chan struct{}) (uint16, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return 0, err
	}

	svc, err := clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}
	if svc == nil {
		return 0, fmt.Errorf("no such service: %+v", name)
	}

	labels := []string{}
	for key, val := range svc.Spec.Selector {
		labels = append(labels, key+"="+val)
	}
	label := strings.Join(labels, ",")

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: label, Limit: 1})
	if err != nil {
		return 0, err
	}
	if len(pods.Items) == 0 {
		return 0, fmt.Errorf("no such pods of the service of %v", name)
	}
	pod := pods.Items[0]

	podPort := int32(0)
	found := false
	for _, container := range pod.Spec.Containers {
		if found {
			break
		}
		if container.Name == "controller" {
			for _, port := range container.Ports {
				if port.Name == "grpc" {
					podPort = port.ContainerPort
					found = true
					break
				}
			}
		}
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, pod.Name)
	url, err := url.Parse(config.Host + path)
	if err != nil {
		return 0, err
	}

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return 0, err
	}

	readyCh := make(chan struct{})
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, url)
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", 0, podPort)}, stopCh, readyCh, nil, os.Stderr)
	if err != nil {
		return 0, err
	}

	errCh := make(chan error)
	go func() {
		errCh <- fw.ForwardPorts()
	}()
	select {
	case <-readyCh:
	case <-errCh:
	}

	p, err := fw.GetPorts()
	if err != nil {
		stopCh <- struct{}{}
		return 0, err
	}

	return p[0].Local, nil
}
