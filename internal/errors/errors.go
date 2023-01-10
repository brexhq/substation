package errors

// Error is an exported constant error string
// use of constant errors is based on this blog post: https://dave.cheney.net/2016/04/07/constant-errors
// constant errors make it easier to reference errors across the application
type Error string

func (e Error) Error() string { return string(e) }

// ErrInvalidFactoryInput is returned when an unsupported input is referenced in any factory function.
const ErrInvalidFactoryInput = Error("invalid factory input")

// ErrMissingRequiredOptions is returned when a component does not have the required options to properly run.
const ErrMissingRequiredOptions = Error("missing required options")

// ErrInvalidType is returned when a component is configured with an invalid type.
const ErrInvalidType = Error("invalid type")
