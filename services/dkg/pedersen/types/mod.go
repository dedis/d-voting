package types

type Status uint16

const (
	// Initialized is when the service has been initialized
	Initialized Status = 0
	// Setup is when the service was set up
	Setup Status = 1
	// Failed is when the service failed to set up
	Failed Status = 2
)
