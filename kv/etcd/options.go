// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package etcd

import (
	"errors"
	"fmt"
	"time"

	"github.com/m3db/m3x/instrument"
	"github.com/m3db/m3x/retry"
)

var (
	defaultRequestTimeout         = 10 * time.Second
	defaultWatchChanCheckInterval = 10 * time.Second
	defaultWatchChanResetInterval = 10 * time.Second
	defaultWatchChanInitTimeout   = 10 * time.Second
	defaultRetryOptions           = retry.NewOptions().SetMaxRetries(5)
	defaultCacheFileFn            = func(string) string { return "" }
)

// CacheFileFn is a function to generate cache file path
type CacheFileFn func(namespace string) string

// Options are options for the client of the kv store
type Options interface {
	// RequestTimeout is the timeout for etcd requests
	RequestTimeout() time.Duration
	// SetRequestTimeout sets the RequestTimeout
	SetRequestTimeout(t time.Duration) Options

	// InstrumentsOptions is the instrument options
	InstrumentsOptions() instrument.Options
	// SetInstrumentsOptions sets the InstrumentsOptions
	SetInstrumentsOptions(iopts instrument.Options) Options

	// RetryOptions is the retry options
	RetryOptions() retry.Options
	// SetRetryOptions sets the RetryOptions
	SetRetryOptions(ropts retry.Options) Options

	// WatchChanCheckInterval will be used to periodically check if a watch chan
	// is no longer being subscribed and should be closed
	WatchChanCheckInterval() time.Duration
	// SetWatchChanCheckInterval sets the WatchChanCheckInterval
	SetWatchChanCheckInterval(t time.Duration) Options

	// WatchChanResetInterval is the delay before resetting the etcd watch chan
	WatchChanResetInterval() time.Duration
	// SetWatchChanResetInterval sets the WatchChanResetInterval
	SetWatchChanResetInterval(t time.Duration) Options

	// WatchChanInitTimeout is the timeout for a watchChan initialization
	WatchChanInitTimeout() time.Duration
	// SetWatchChanInitTimeout sets the WatchChanInitTimeout
	SetWatchChanInitTimeout(t time.Duration) Options

	// Prefix is the prefix for each key
	Prefix() string
	// SetPrefix sets the prefix
	SetPrefix(s string) Options
	// ApplyPrefix applies the prefix to the key
	ApplyPrefix(key string) string

	// CacheFileDir is the dir for cache.
	CacheFileFn() CacheFileFn
	// SetCacheFileDir sets the CacheFileDir
	SetCacheFileFn(fn CacheFileFn) Options

	// Validate validates the Options
	Validate() error
}

type options struct {
	requestTimeout         time.Duration
	prefix                 string
	iopts                  instrument.Options
	ropts                  retry.Options
	watchChanCheckInterval time.Duration
	watchChanResetInterval time.Duration
	watchChanInitTimeout   time.Duration
	cacheFileFn            CacheFileFn
}

// NewOptions creates a sane default Option
func NewOptions() Options {
	o := options{}
	return o.SetRequestTimeout(defaultRequestTimeout).
		SetInstrumentsOptions(instrument.NewOptions()).
		SetRetryOptions(defaultRetryOptions).
		SetWatchChanCheckInterval(defaultWatchChanCheckInterval).
		SetWatchChanResetInterval(defaultWatchChanResetInterval).
		SetWatchChanInitTimeout(defaultWatchChanInitTimeout).
		SetCacheFileFn(defaultCacheFileFn)
}

func (o options) Validate() error {
	if o.iopts == nil {
		return errors.New("no instrument options")
	}

	if o.ropts == nil {
		return errors.New("no retry options")
	}

	if o.watchChanCheckInterval <= 0 {
		return errors.New("invalid watch channel check interval")
	}

	return nil
}

func (o options) RequestTimeout() time.Duration {
	return o.requestTimeout
}

func (o options) SetRequestTimeout(t time.Duration) Options {
	o.requestTimeout = t
	return o
}

func (o options) InstrumentsOptions() instrument.Options {
	return o.iopts
}

func (o options) SetInstrumentsOptions(iopts instrument.Options) Options {
	o.iopts = iopts
	return o
}

func (o options) RetryOptions() retry.Options {
	return o.ropts
}

func (o options) SetRetryOptions(ropts retry.Options) Options {
	o.ropts = ropts
	return o
}

func (o options) WatchChanCheckInterval() time.Duration {
	return o.watchChanCheckInterval
}

func (o options) SetWatchChanCheckInterval(t time.Duration) Options {
	o.watchChanCheckInterval = t
	return o
}

func (o options) WatchChanResetInterval() time.Duration {
	return o.watchChanResetInterval
}

func (o options) SetWatchChanResetInterval(t time.Duration) Options {
	o.watchChanResetInterval = t
	return o
}

func (o options) WatchChanInitTimeout() time.Duration {
	return o.watchChanInitTimeout
}

func (o options) SetWatchChanInitTimeout(t time.Duration) Options {
	o.watchChanInitTimeout = t
	return o
}

func (o options) CacheFileFn() CacheFileFn {
	return o.cacheFileFn
}

func (o options) SetCacheFileFn(fn CacheFileFn) Options {
	o.cacheFileFn = fn
	return o
}

func (o options) Prefix() string {
	return o.prefix
}

func (o options) SetPrefix(prefix string) Options {
	o.prefix = prefix
	return o
}

func (o options) ApplyPrefix(key string) string {
	if o.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s/%s", o.prefix, key)
}
