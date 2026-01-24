package ihfs

// Operation represents a file system operation.
type Operation interface {
	// Subject returns the subject of the operation, typically a file or directory path.
	Subject() string
}
