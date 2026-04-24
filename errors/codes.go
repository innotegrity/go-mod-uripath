package errors

const (
	// InvalidParameterErrorCode is returned when an invalid parameter is passed to a function.
	InvalidParameterErrorCode = 2

	// SchemeNotFoundErrorCode is returned when a scheme is not registered.
	SchemeNotFoundErrorCode = 10

	// SchemeExistsErrorCode is returned when a scheme being registered already exists.
	SchemeExistsErrorCode = 11

	// InvalidSchemeErrorCode is returned when a scheme is invalid.
	InvalidSchemeErrorCode = 12

	// BackendDeleteErrorCode is returned when a deletion of a resource fails.
	BackendDeleteErrorCode = 20

	// BackendExistsErrorCode is returned when an error occurs while checking for the existing of a resource.
	BackendExistsErrorCode = 21

	// BackendGetErrorCode is returned when retrieving the contents of a resource fails.
	BackendGetErrorCode = 22

	// BackendListErrorCode is returned when listing a resource fails.
	BackendListErrorCode = 23

	// BackendPutErrorCode is returned when writing the contents of a resource fails.
	BackendPutErrorCode = 24

	// BackendInitErrorCode is returned when initializing a backend fails.
	BackendInitErrorCode = 25
)
