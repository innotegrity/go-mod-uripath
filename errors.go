package uripath

const (
	// InvalidParameter is returned when an invalid parameter is passed to a function.
	InvalidParameter = 2

	// SchemeNotFound is returned when a scheme is not registered.
	SchemeNotFound = 3

	// SchemeExists is returned when a scheme being registered already exists.
	SchemeExists = 4

	// BackendDeleteError is returned when a deletion of a resource fails.
	BackendDeleteError = 20

	// BackendExistsError is returned when an error occurs while checking for the existing of a resource.
	BackendExistsError = 21

	// BackendGetError is returned when retrieving the contents of a resource fails.
	BackendGetError = 22

	// BackendListError is returned when listing a resource fails.
	BackendListError = 23

	// BackendPutError is returned when writing the contents of a resource fails.
	BackendPutError = 24

	// BackendInitError is returned when initializing a backend fails.
	BackendInitError = 25
)
