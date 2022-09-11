package webmod

// Tools is the type used to instantiate this module.
// Any variable of this type will have access to all
// the methods with the receiver *Tools.
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}
