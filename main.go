package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var (
	var1             bool
	var2             bool
	hostnameOverride string
)

func init() {
	flag.BoolVar(&var1, "var1", false, "If set, do something")
	flag.BoolVar(&var2, "var2", true, "If set, do something else (default true)")
	flag.StringVar(&hostnameOverride, "hostname-override", "", "If non-empty, will be used as the name of the Node that kube-network-policies is running on. If unset, the node name is assumed to be the same as the node's hostname.")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: kube-network-thing [options]\n\n")
		flag.PrintDefaults()
	}
}

func main() {
	// enable logging
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	flag.VisitAll(func(flag *flag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
	// create a Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	config.UserAgent = "kube-network-thing"
	// use protobuf for better performance at scale
	// https://kubernetes.io/docs/reference/using-api/api-concepts/#alternate-representations-of-resources
	config.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
	config.ContentType = "application/vnd.kubernetes.protobuf"

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// trap Ctrl+C and call cancel on the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Enable signal handler
	signalCh := make(chan os.Signal, 2)
	defer func() {
		close(signalCh)
		cancel()
	}()
	signal.Notify(signalCh, os.Interrupt, unix.SIGINT)

	go func() {
		select {
		case <-signalCh:
			klog.Infof("Exiting: received signal")
			cancel()
		case <-ctx.Done():
		}
	}()

}
