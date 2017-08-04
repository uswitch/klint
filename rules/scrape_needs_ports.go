package rules

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	extv1b1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uswitch/klint/alerts"
)

func validScrapeAndPorts(d *extv1b1.Deployment) bool {
	_, hasScrapeAnnotation := d.Spec.Template.Annotations["prometheus.io.scrape"]

	hasPorts := false
	for _, c := range d.Spec.Template.Spec.Containers {
		if len(c.Ports) > 0 {
			hasPorts = true
			break
		}
	}

	return !hasScrapeAnnotation || hasPorts
}

var ScrapeNeedsPortsRule = NewRule(
	func(old runtime.Object, new runtime.Object, out chan *alerts.Alert) {
		deployment := new.(*extv1b1.Deployment)

		if old == nil || validScrapeAndPorts(old.(*extv1b1.Deployment)) != validScrapeAndPorts(deployment) {
			podName := podNameForDeployment(deployment)

			if validScrapeAndPorts(deployment) { // everything is good
				if old != nil {
					out <- &alerts.Alert{new, fmt.Sprintf("Thanks for sorting the ports for scraping on %s.%s", deployment.ObjectMeta.Namespace, podName)}
				}
			} else { // stuff has gone bad
				out <- &alerts.Alert{new, fmt.Sprintf("%s.%s wants to be scraped so it needs to expose some ports", deployment.ObjectMeta.Namespace, podName)}
			}
		} else {
			log.Debugf("ScrapeNeedsPortsRule: %s.%s hadn't changed", deployment.ObjectMeta.Namespace, podNameForDeployment(deployment))
		}
	},
	WantDeployments,
)
