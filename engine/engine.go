package engine

import (
	"context"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const ANNOTATION_PREFIX = "com.uswitch.alert"

type Engine struct {
	namespaceIndexer cache.Indexer
	clientSet        *kubernetes.Clientset
	informers        map[string]cache.SharedInformer
	rules            []*Rule
	outputs          map[string]Output
}

func NewEngine(clientSet *kubernetes.Clientset) *Engine {
	return &Engine{
		clientSet: clientSet,
		informers: map[string]cache.SharedInformer{},
		rules:     []*Rule{},
		outputs:   map[string]Output{},
	}
}

func (e *Engine) AddRule(rule *Rule) {
	e.rules = append(e.rules, rule)
}

func (e *Engine) AddOutput(output Output) {
	e.outputs[output.Key()] = output
}

func (e *Engine) watchNamespaces(context context.Context) {
	listWatcher := cache.NewListWatchFromClient(e.clientSet.CoreV1().RESTClient(), "namespaces", "", fields.Everything())
	indexer, informer := cache.NewIndexerInformer(listWatcher, &v1.Namespace{}, 0, cache.ResourceEventHandlerFuncs{}, cache.Indexers{})

	go informer.Run(context.Done())

	if !cache.WaitForCacheSync(context.Done(), informer.HasSynced) {
		log.Errorf("Timed out waiting for caches to sync")
		return
	}

	e.namespaceIndexer = indexer
}

func resourceAge(resource interface{}) (time.Duration, error) {
	metaObj, err := meta.Accessor(resource)

	if err != nil {
		return 0, err
	}

	return time.Now().Sub(metaObj.GetCreationTimestamp().Time), nil
}

func bind(rule *Rule, informer cache.SharedInformer, ageLimit int, ctx *RuleHandlerContext) {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// we should make sure the resource is actually new, not just newly seen
			age, err := resourceAge(obj)

			if err == nil && ageLimit > 0 && age > time.Minute*time.Duration(ageLimit) {
				metaObj, _ := meta.Accessor(obj)

				log.Debugf("%s.%s was too old when added", metaObj.GetNamespace(), metaObj.GetName())
			} else {
				rule.Handler(nil, obj.(runtime.Object), ctx)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			rule.Handler(old.(runtime.Object), new.(runtime.Object), ctx)
		},
	})
}

func (e *Engine) attachRules(context context.Context, namespace string, ageLimit int) <-chan *Alert {
	for _, want := range UniqueWants(e.rules) {
		log.Debugf("Adding a shared informer for %s", want.Name)
		listWatcher := cache.NewListWatchFromClient(want.RESTClient(e.clientSet), want.Name, namespace, fields.Everything())
		informer := cache.NewSharedInformer(listWatcher, want.Object, 0)

		go informer.Run(context.Done())

		e.informers[want.Name] = informer
	}

	alerts := make(chan *Alert)
	ctx := &RuleHandlerContext{
		alerts:    alerts,
		clientset: e.clientSet,
	}

	for _, rule := range e.rules {
		for _, want := range rule.Wants {
			bind(rule, e.informers[want.Name], ageLimit, ctx)
		}
	}

	return alerts

}

func extractOutputAnnotations(annotations map[string]string, out map[string]string) {
	for k, v := range annotations {
		if strings.HasPrefix(k, ANNOTATION_PREFIX) {
			if parts := strings.Split(k, "/"); len(parts) > 1 {
				out[parts[1]] = v
			}
		}
	}
}

func (e *Engine) Run(context context.Context, namespace string, ageLimit int) {
	accessor := meta.NewAccessor()

	e.watchNamespaces(context)
	alerts := e.attachRules(context, namespace, ageLimit)
	filteredAlerts := filterAlerts(context, alerts)

	for {
		select {
		case alert := <-filteredAlerts:
			log.Debugf("ALERT: %s", alert.Message)

			outputAnnotations := map[string]string{}

			if e.namespaceIndexer != nil {
				ns, _ := accessor.Namespace(alert.Resource)
				nsResource, _, _ := e.namespaceIndexer.GetByKey(ns)
				nsAnnotations, _ := accessor.Annotations(nsResource.(runtime.Object))
				extractOutputAnnotations(nsAnnotations, outputAnnotations) // kind of nasty state mutation of outputAnnotations
			}

			annotations, _ := accessor.Annotations(alert.Resource)
			extractOutputAnnotations(annotations, outputAnnotations) // kind of nasty state mutation of outputAnnotations

			if len(outputAnnotations) == 0 {
				resourceName, _ := accessor.Name(alert.Resource)
				log.Debugf("There where no output annotations found on resource %s", resourceName)
			}

			resourceVersion, _ := accessor.ResourceVersion(alert.Resource)
			log.Debugf("ResourceVersion: %s", resourceVersion)

			for outputKey, outputVal := range outputAnnotations {
				if output, ok := e.outputs[outputKey]; ok {
					output.Send(outputVal, alert.Message)
				} else {
					log.Warnf("There is no output '%s'", outputKey)
				}
			}
		}
	}
}
