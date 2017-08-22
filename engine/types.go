package engine

import (
	"fmt"

	batchv2 "k8s.io/api/batch/v2alpha1"
	"k8s.io/api/core/v1"
	extv1b1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/satori/go.uuid"
)

type Output interface {
	Key() string
	Send(string, string) error
}

type Alert struct {
	Rule     *Rule
	Resource runtime.Object
	Message  string
}

func NewAlert(resource runtime.Object, message string) *Alert {
	return &Alert{
		Resource: resource,
		Message:  message,
	}
}

type Want struct {
	Name       string
	Object     runtime.Object
	RESTClient func(*kubernetes.Clientset) rest.Interface
}

var (
	WantPods = Want{
		"pods", &v1.Pod{},
		func(cs *kubernetes.Clientset) rest.Interface {
			return cs.CoreV1().RESTClient()
		},
	}
	WantDeployments = Want{
		"deployments", &extv1b1.Deployment{},
		func(cs *kubernetes.Clientset) rest.Interface {
			return cs.ExtensionsV1beta1().RESTClient()
		},
	}
	WantCronJobs = Want{
		"cronjobs", &batchv2.CronJob{},
		func(cs *kubernetes.Clientset) rest.Interface {
			return cs.BatchV2alpha1().RESTClient()
		},
	}
)

type RuleHandlerContext struct {
	alerts    chan *Alert
	clientset *kubernetes.Clientset
}

func (ctx *RuleHandlerContext) Alert(obj runtime.Object, message string) {
	ctx.alerts <- NewAlert(obj, message)
}

func (ctx *RuleHandlerContext) Alertf(obj runtime.Object, format string, objs ...interface{}) {
	ctx.Alert(obj, fmt.Sprintf(format, objs...))
}

type RuleHandler func(runtime.Object, runtime.Object, *RuleHandlerContext)

type Rule struct {
	Id      string
	Wants   []Want
	Handler RuleHandler
}

func NewRule(handler RuleHandler, wants ...Want) *Rule {
	rule := &Rule{
		Id:      uuid.NewV4().String(),
		Wants:   wants,
		Handler: handler,
	}

	return rule
}

func UniqueWants(rules []*Rule) []Want {
	haveWantFor := map[string]bool{}
	wants := []Want{}

	for _, rule := range rules {
		for _, want := range rule.Wants {
			if _, ok := haveWantFor[want.Name]; !ok {
				haveWantFor[want.Name] = true
				wants = append(wants, want)
			}
		}
	}

	return wants
}
