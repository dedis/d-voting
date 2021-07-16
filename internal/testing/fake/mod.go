// Package fake provides fake implementations for interfaces commonly used in
// the repository.
// The implementations offer configuration to return errors when it is needed by
// the unit test and it is also possible to record the call of functions of an
// object in some cases.
package fake

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func init() {
	// A random value is injected every time, so that the error is never the
	// same and prevent hardcoded values in the tests.
	random := make([]byte, 4)
	rand.Read(random)

	fakeErr = xerrors.Errorf("fake error (%x)", random)
}

// fakeErr is initialized with a random value so that the test suite cannot rely
// on a fixed value.
var fakeErr error

// GetError returns the fake error.
func GetError() error {
	return fakeErr
}

// Err returns the expected format of an error returned by a fake component.
func Err(msg string) string {
	return fmt.Sprintf("%s: %v", msg, fakeErr)
}

// Call is a tool to keep track of a function calls.
type Call struct {
	sync.Mutex
	calls [][]interface{}
}

// NewCall returns a new empty call monitor.
func NewCall() *Call {
	return &Call{}
}

// Get returns the nth call ith parameter.
func (c *Call) Get(n, i int) interface{} {
	if c == nil {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	return c.calls[n][i]
}

// Len returns the number of calls.
func (c *Call) Len() int {
	if c == nil {
		return 0
	}

	c.Lock()
	defer c.Unlock()

	return len(c.calls)
}

// Add adds a call to the list.
func (c *Call) Add(args ...interface{}) {
	if c == nil {
		return
	}

	c.Lock()
	defer c.Unlock()

	c.calls = append(c.calls, args)
}

// Clear clears the array of calls.
func (c *Call) Clear() {
	if c != nil {
		c.Lock()
		c.calls = nil
		c.Unlock()
	}
}

// Counter is a helper to delay errors or actions. It can be nil without panics.
type Counter struct {
	Value int
}

// NewCounter returns a new counter set to the given value.
func NewCounter(value int) *Counter {
	return &Counter{
		Value: value,
	}
}

// Done returns true when the counter reached zero.
func (c *Counter) Done() bool {
	return c == nil || c.Value <= 0
}

// Decrease decrements the counter.
func (c *Counter) Decrease() {
	if c == nil {
		return
	}
	c.Value--
}

// WaitLog is a helper to wait for a log to be printed. It executes the callback
// when it detects it.
func WaitLog(msg string, t time.Duration) (zerolog.Logger, func(t *testing.T)) {
	reader, writer := io.Pipe()
	done := make(chan struct{})
	found := false

	buffer := new(bytes.Buffer)
	tee := io.TeeReader(reader, buffer)

	go func() {
		select {
		case <-done:
		case <-time.After(t):
			writer.Close()
		}
	}()

	go func() {
		defer close(done)

		data := make([]byte, 1024)

		for {
			n, err := tee.Read(data)
			if err != nil {
				return
			}

			if strings.Contains(string(data[:n]), fmt.Sprintf(`"%s"`, msg)) {
				found = true
				return
			}
		}
	}()

	wait := func(t *testing.T) {
		<-done
		if !found {
			t.Fatalf("log not found in %s", buffer.String())
		}
	}

	return zerolog.New(writer), wait
}

// CheckLog returns a logger and a check function. When called, the function
// will verify if the logger has seen the message printed.
func CheckLog(msg string) (zerolog.Logger, func(t *testing.T)) {
	buffer := new(bytes.Buffer)

	check := func(t *testing.T) {
		require.Contains(t, buffer.String(), fmt.Sprintf(`"%s"`, msg))
	}

	return zerolog.New(buffer), check
}
