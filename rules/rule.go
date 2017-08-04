package rules

import (
	"k8s.io/api/core/v1"
	extv1b1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/uswitch/klint/alerts"
)

type Want struct {
	Name string
	Object runtime.Object
	RESTClient func(*kubernetes.Clientset)rest.Interface
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
)

type RuleHandler func(runtime.Object, runtime.Object, chan *alerts.Alert)

type Rule struct {
	Wants []Want
	Handler RuleHandler
}

func NewRule(handler RuleHandler, wants ...Want) *Rule {
	return &Rule{
		Wants: wants,
		Handler: handler,
	}
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
