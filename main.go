package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	"gopkg.in/alecthomas/kingpin.v2"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/uswitch/klint/alerts"
	"github.com/uswitch/klint/engine"
	"github.com/uswitch/klint/rules"
)

type options struct {
	kubeconfig string
	namespace  string
	debug      bool
	slackToken string
	awsRegion  string
	ageLimit   int
	jsonFormat bool
}

func createClientConfig(opts *options) (*rest.Config, error) {
	if opts.kubeconfig == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", opts.kubeconfig)
}

func createClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func main() {
	opts := &options{}
	kingpin.Flag("kubeconfig", "Path to kubeconfig.").StringVar(&opts.kubeconfig)
	kingpin.Flag("namespace", "Namespace to monitor").Default("").StringVar(&opts.namespace)
	kingpin.Flag("age-limit", "Will discard updates for resources old than n minutes. 0 disables").Default("5").IntVar(&opts.ageLimit)
	kingpin.Flag("debug", "Debug mode").BoolVar(&opts.debug)
	kingpin.Flag("slack-token", "").Envar("SLACK_TOKEN").StringVar(&opts.slackToken)
	kingpin.Flag("aws-region", "").Envar("AWS_REGION").Default("eu-west-1").StringVar(&opts.awsRegion)
	kingpin.Flag("json", "Output log data in JSON format").Default("false").BoolVar(&opts.jsonFormat)

	kingpin.Parse()

	if opts.debug {
		log.SetLevel(log.DebugLevel)
		log.Debugln("Debug logging enabled")
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if opts.jsonFormat {
		log.SetFormatter(&log.JSONFormatter{})
	}

	config, err := createClientConfig(opts)
	if err != nil {
		log.Fatalf("error creating client config: %s", err)
	}

	clientSet, err := createClientSet(config)
	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}

	executionContext, stop := context.WithCancel(context.Background())
	defer stop()

	engine := engine.NewEngine(clientSet)

	engine.AddRule(rules.UnsuccessfulExitRule)
	engine.AddRule(rules.ResourceAnnotationRule)
	engine.AddRule(rules.ScrapeNeedsPortsRule)
	engine.AddRule(rules.ValidIAMRoleRule)
	engine.AddRule(rules.RequireCronJobHistoryLimits)
	engine.AddRule(rules.IngressNeedsAnnotation)

	if len(opts.slackToken) > 0 {
		engine.AddOutput(alerts.NewSlackOutput(opts.slackToken))
	}
	engine.AddOutput(alerts.NewSNSOutput(opts.awsRegion))

	go engine.Run(executionContext, opts.namespace, opts.ageLimit)
	select {}
}
