package engine

import (
	"context"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	metatypes "k8s.io/apimachinery/pkg/types"
)

var testRule = NewRule(
	func(_ runtime.Object, _ runtime.Object, _ *RuleHandlerContext) {},
)

func createResource(id string) runtime.Object {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: metatypes.UID(id),
		},
	}
}

func TestThing(t *testing.T) {
	in := make(chan *Alert, 3)

	in <- &Alert{testRule, createResource("123"), "Foobles"}
	in <- &Alert{testRule, createResource("123"), "Foobles"}
	in <- &Alert{testRule, createResource("123"), "Barbles"}

	filterContext, cancelFilter := context.WithCancel(context.Background())
	outCh := filterAlerts(filterContext, in)

	outArr := []*Alert{}
	for alert := range outCh {
		outArr = append(outArr, alert)

		if len(outArr) >= 2 {
			cancelFilter()
			break
		}
	}

	if outArr[1].Message == "Foobles" {
		t.Fatal("Didn't filter out 2nd message")
	}

	cancelFilter()
}
