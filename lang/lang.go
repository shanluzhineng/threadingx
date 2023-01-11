//this file is copied from https://github.com/zeromicro/go-zero/blob/master/core/lang/lang.go

package lang

// Placeholder is a placeholder object that can be used globally.
var Placeholder PlaceholderType

type (
	// AnyType can be used to hold any type.
	AnyType = interface{}
	// PlaceholderType represents a placeholder type.
	PlaceholderType = struct{}
)
