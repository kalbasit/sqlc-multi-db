package generator

import (
	"errors"
	"fmt"
)

var (
	errInvalidDictCall       = errors.New("invalid dict call")
	errDictKeysMustBeStrings = errors.New("dict keys must be strings")

	errSliceDomainStructNotSupported = errors.New(
		"slices of domain structs are not supported as direct parameters, as they require a conversion loop" +
			" to be generated. The auto-looping for bulk inserts handles this by operating on a struct" +
			" parameter containing a slice",
	)
)

func errUnsupportedSliceDomainStruct(t string) error {
	return fmt.Errorf("unsupported parameter type: slice of domain struct %s: %w", t, errSliceDomainStructNotSupported)
}
