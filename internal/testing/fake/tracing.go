package fake

import opentracing "github.com/opentracing/opentracing-go"

// GetTracerForAddr is used to mock `tracing.GetTracerForAddr` with an error.
func GetTracerForAddrWithError(addr string) (opentracing.Tracer, error) {
	return nil, fakeErr
}
