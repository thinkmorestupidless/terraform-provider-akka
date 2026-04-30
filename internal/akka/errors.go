package akka

import "fmt"

// NotFoundError is returned when an Akka resource does not exist.
type NotFoundError struct {
	ResourceType string
	Name         string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %q not found", e.ResourceType, e.Name)
}

// IsNotFound returns true if err is a *NotFoundError.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}
