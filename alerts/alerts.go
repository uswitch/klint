package alerts

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type Output interface {
	Key() string
	Send(string, string) error
}

type Alert struct {
	Resource runtime.Object
	Message  string
}
