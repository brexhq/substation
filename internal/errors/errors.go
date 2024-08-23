package errors

import "fmt"

// ErrInvalidFactoryInput is returned when an unsupported input is referenced in any factory function.
var ErrInvalidFactoryInput = fmt.Errorf("invalid factory input")

// ErrMissingRequiredOption is returned when a component does not have the required options to properly run.
var ErrMissingRequiredOption = fmt.Errorf("missing required option")

// ErrInvalidOption is returned when an invalid option is received in a constructor.
var ErrInvalidOption = fmt.Errorf("invalid option")
