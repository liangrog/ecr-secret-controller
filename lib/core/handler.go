// Event handler interface
package core

// Event handler interface being used for controller
// This is where all the custom business logic lives
type Handler interface {
	// Object created event
	Create(obj interface{}) error

	// Object updated event
	Update(old, new interface{}) error

	// Object deleted event
	Delete(obj interface{}) error
}
