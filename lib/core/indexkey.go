// Index key that holds new and old index key
package core

// Index key cater for passing
// old key and new key so we can
// handling update event
type IndexKey struct {
	// Olde object key, for update event
	Old string `json:"old,omitempty"`

	// New object key
	New string `json:"new"`
}
