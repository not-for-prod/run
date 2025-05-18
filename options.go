package run

import "time"

// DefaultTimeout is the default duration used for both starting and stopping
// an application. It can be customized using the WithStartTimeout and
// WithStopTimeout options.
const DefaultTimeout = 15 * time.Second

// options holds configurable parameters for the Group's behavior.
type options struct {
	startTimeout time.Duration // maximum allowed time for start functions to complete
	stopTimeout  time.Duration // maximum allowed time for stop functions to complete
}

// defaultOptions provides the default timeout values used by NewGroup.
var defaultOptions = options{
	startTimeout: DefaultTimeout,
	stopTimeout:  DefaultTimeout,
}

// Option is a functional option that modifies Group's internal options.
type Option interface {
	apply(*options)
}

// optionFunc is a helper type to implement the Option interface with functions.
type optionFunc func(*options)

// apply executes the function to modify the options.
func (f optionFunc) apply(o *options) {
	f(o)
}

// WithStartTimeout returns an Option that sets the timeout duration for
// starting components. This timeout controls how long Wait() will wait for
// all Start functions to complete before timing out.
//
// Default is DefaultTimeout (15 seconds).
func WithStartTimeout(v time.Duration) Option {
	return optionFunc(func(o *options) {
		o.startTimeout = v
	})
}

// WithStopTimeout returns an Option that sets the timeout duration for
// stopping components. This timeout controls how long stop functions have
// to complete before an early exit.
//
// Default is DefaultTimeout (15 seconds).
func WithStopTimeout(v time.Duration) Option {
	return optionFunc(func(o *options) {
		o.stopTimeout = v
	})
}
