package engine

import (
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/meta"
)

func filterAlerts(context context.Context, in <-chan *Alert) chan *Alert {
	out := make(chan *Alert)
	accessor := meta.NewAccessor()
	lastSeenByIdent := map[string]string{}

	go func() {
		for {
			select {
			case <-context.Done():
				return
			case alert := <-in:
				uid, _ := accessor.UID(alert.Resource)
				ident := fmt.Sprintf("%s:%s", alert.Rule.Id, string(uid))

				if message, ok := lastSeenByIdent[ident]; !ok || (ok && alert.Message != message) {
					out <- alert
				} else {
					log.Debug("Alert filtered")
				}

				lastSeenByIdent[ident] = alert.Message
			}
		}
	}()

	return out
}
