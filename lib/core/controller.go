// @Author Roger Liang
//
// General purose controller using
// - Rate limiting queue
// - Shared index informer
// - Custom index key which serialised to json string
//   containers old key and new key to facilitate update handler
//
// The container conform to cache.Controller interface
package core

import (
	"fmt"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// Work queue item max retry when errors
	retry = 5
)

// Controller
type Controller struct {
	// Work queue
	queue workqueue.RateLimitingInterface

	// Shared informer
	informer cache.SharedIndexInformer

	// Cache indexer
	indexer cache.Indexer

	// Handler for event
	handler Handler
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

// Run worker (can be run multiple concurrently)
func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

// Processing items from the queue
func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.processItem(key)
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)

	return true
}

// Processing items
// Can handler normal string key or IndexKey
func (c *Controller) processItem(key interface{}) error {
	var newKey, oldKey string
	var new interface{}

	if v, ok := key.(IndexKey); ok {
		newKey = v.New
		oldKey = v.Old
	} else {
		newKey = key.(string)
	}

	new, exists, err := c.indexer.GetByKey(newKey)
	if err != nil {
		log.Printf("Fetching object with key %s from store failed with %v", newKey, err)
		return err
	}

	// If namespace has been deleted
	if !exists {
		return c.handler.Delete(new)
	} else {
		// If it's an update
		if oldKey != "" {
			old, oldExists, err := c.indexer.GetByKey(oldKey)
			if err != nil {
				log.Printf("Fetching object with key %s from store failed with %v", oldKey, err)
				return err
			}

			// Note: No error given, just warning, the handler for update has to deal with it
			if !oldExists {
				log.Printf("Old object with key %s from store doesn't exist anymore", oldKey)
			}

			// Make sure you check the old object as it might have
			// delete before use
			return c.handler.Update(old, new)
		} else {
			// If it's creation
			return c.handler.Create(new)
		}
	}

	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < retry {
		log.Printf("Error syncing object %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Printf("Dropping object work for index key %q out of the queue: %v", key, err)
}

// Run controller
func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()

	log.Println("Starting controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	// Spawn workers
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	// Stop when we receive stop signal
	<-stopCh

	log.Println("Stopping controller")
}

// Create a new controller
func NewController(queue workqueue.RateLimitingInterface, informer cache.SharedIndexInformer, indexer cache.Indexer, handler Handler) *Controller {
	return &Controller{
		queue:    queue,
		informer: informer,
		indexer:  indexer,
		handler:  handler,
		//logger:   logger,
	}
}
