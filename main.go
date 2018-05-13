// Kubernetes controller for generating docker auth
// secret so to provide access to private container repo
//
// ENVs:
//
// AWS_ACCESS_KEY_ID: ECR access key id
// AWS_SECRET_ACCESS_KEY: ECR access secret key
// AWS_DEFAULT_REGION: ECR default region
// ECR_RESYNC_PERIOD: ECR auth token refresh period in hours, default 2  (ECR token expiry is 12 hours)
// ECR_SECRET_NAME: Name of the ECR screte to create, default 'ecr'
//
// KUBE_CONFIG_TYPE: Kubernetes config file type, default 'in'. If specified 'out', you have to provide 'KUBE_CONFIG_PATH'
// KUBE_CONFIG_PATH: Kubernetes config file path. Mandatory when setting KUBE_CONFIG_TYPE to 'out'
// WORKER_NUMBER: Number of concurrent workers
// EXCLUDE_NAMESPACES: Namespaces not to observe. Default all namespaces are oberved

package main

import (
	"log"
	"os"
	"strconv"
	"time"

	lc "github.com/liangrog/kctlr-docker-auth/lib/core"
	lh "github.com/liangrog/kctlr-docker-auth/lib/handlers"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// We watch namespaces
	watchObj = "namespaces"

	// Global default sync period
	defaultSyncPeriod = 1 * time.Hour
)

func main() {
	workerNumber, _ := strconv.Atoi(lc.GetEnvWithDefault("WORKER_NUMBER", "2"))

	// Create kubernetes client
	configType := lc.GetEnvWithDefault("KUBE_CONFIG_TYPE", "in")
	clientset := lc.GetClient(configType)

	// create the list watcher
	listWatcher := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		watchObj,
		v1.NamespaceAll,
		fields.Everything(),
	)

	// Using shared informer for better
	// performance when we have multiple
	// controllers down the track
	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&v1.Namespace{},
		defaultSyncPeriod,
		cache.Indexers{},
	)

	// ECR
	// Now let's start the controller
	ecrCtlrStop := make(chan struct{})
	defer close(ecrCtlrStop)
	go ecrController(clientset, informer).Run(workerNumber, ecrCtlrStop)

	// Wait forever
	select {}
}

// Get ECR controller
func ecrController(client kubernetes.Interface, informer cache.SharedIndexInformer) *lc.Controller {
	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	handler := lh.NewEcrHandler(
		client,
		lc.CommaStrToSlice(os.Getenv("EXCLUDE_NAMESPACES")),
		os.Getenv("ECR_SECRET_NAME"),
	)

	// Get resync period
	var resyncPeriod time.Duration
	if rp, err := strconv.ParseInt(lc.GetEnvWithDefault("ECR_RESYNC_PERIOD", "5"), 10, 64); err != nil {
		log.Fatal("Failed to parse 'RESYNC_PERIOD' environment variable")
	} else {
		resyncPeriod = time.Duration(rp) * time.Hour
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					log.Printf("Processing event of creating namespace %s", key)
					queue.Add(lc.IndexKey{New: key})
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldKey, err := cache.MetaNamespaceKeyFunc(old)
				newKey, err := cache.MetaNamespaceKeyFunc(new)
				if err == nil {
					log.Printf("Processing event of updating namespace %s", newKey)
					queue.Add(lc.IndexKey{New: newKey, Old: oldKey})
				}
			},
			DeleteFunc: func(obj interface{}) {
				// IndexerInformer uses a delta queue, therefore for deletes we have to use this
				// key function.
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err == nil {
					log.Printf("Processing event of deleting namespace %s", key)
					queue.Add(lc.IndexKey{New: key})
				}
			},
		},
		resyncPeriod,
	)

	return lc.NewController(queue, informer, informer.GetIndexer(), handler)
}
