// Copyright (c) 2017 Uber Technologies, Inc.
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

package util

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/m3db/m3cluster/generated/proto/commonpb"
	"github.com/m3db/m3cluster/kv"
	"github.com/m3db/m3cluster/kv/mem"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testNow = time.Now()
)

func TestWatchAndUpdateBool(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v bool
	}{}

	valueFn := func() bool {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store        = mem.NewStore()
		defaultValue = false
	)

	watch, err := WatchAndUpdateBool(
		store, "foo", &testConfig.v, &testConfig.RWMutex, defaultValue, nil,
	)
	require.NoError(t, err)

	// Valid update.
	_, err = store.Set("foo", &commonpb.BoolProto{Value: false})
	require.NoError(t, err)
	for {
		if !valueFn() {
			break
		}
	}

	// Malformed updates should not be applied.
	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 20})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.False(t, valueFn())

	_, err = store.Set("foo", &commonpb.BoolProto{Value: true})
	require.NoError(t, err)
	for {
		if valueFn() {
			break
		}
	}

	// Nil updates should apply the default value.
	_, err = store.Delete("foo")
	require.NoError(t, err)
	for {
		if !valueFn() {
			break
		}
	}

	_, err = store.Set("foo", &commonpb.BoolProto{Value: true})
	require.NoError(t, err)
	for {
		if valueFn() {
			break
		}
	}

	// Updates should not be applied after the watch is closed and there should not
	// be any goroutines still running.
	watch.Close()
	time.Sleep(100 * time.Millisecond)
	_, err = store.Set("foo", &commonpb.BoolProto{Value: false})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.True(t, valueFn())

	leaktest.Check(t)()
}

func TestWatchAndUpdateFloat64(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v float64
	}{}

	valueFn := func() float64 {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store        = mem.NewStore()
		defaultValue = 1.35
	)

	watch, err := WatchAndUpdateFloat64(
		store, "foo", &testConfig.v, &testConfig.RWMutex, defaultValue, nil,
	)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 3.7})
	require.NoError(t, err)
	for {
		if valueFn() == 3.7 {
			break
		}
	}

	// Malformed updates should not be applied.
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: 1})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 3.7, valueFn())

	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 1.2})
	require.NoError(t, err)
	for {
		if valueFn() == 1.2 {
			break
		}
	}

	// Nil updates should apply the default value.
	_, err = store.Delete("foo")
	require.NoError(t, err)
	for {
		if valueFn() == defaultValue {
			break
		}
	}

	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 6.2})
	require.NoError(t, err)
	for {
		if valueFn() == 6.2 {
			break
		}
	}

	// Updates should not be applied after the watch is closed and there should not
	// be any goroutines still running.
	watch.Close()
	time.Sleep(100 * time.Millisecond)
	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 7.2})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 6.2, valueFn())

	leaktest.Check(t)
}

func TestWatchAndUpdateInt64(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v int64
	}{}

	valueFn := func() int64 {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store              = mem.NewStore()
		defaultValue int64 = 3
	)

	watch, err := WatchAndUpdateInt64(
		store, "foo", &testConfig.v, &testConfig.RWMutex, defaultValue, nil,
	)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.Int64Proto{Value: 1})
	require.NoError(t, err)
	for {
		if valueFn() == 1 {
			break
		}
	}

	// Malformed updates should not be applied.
	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 100})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, int64(1), valueFn())

	_, err = store.Set("foo", &commonpb.Int64Proto{Value: 7})
	require.NoError(t, err)
	for {
		if valueFn() == 7 {
			break
		}
	}

	// Nil updates should apply the default value.
	_, err = store.Delete("foo")
	require.NoError(t, err)
	for {
		if valueFn() == defaultValue {
			break
		}
	}

	_, err = store.Set("foo", &commonpb.Int64Proto{Value: 21})
	require.NoError(t, err)
	for {
		if valueFn() == 21 {
			break
		}
	}

	// Updates should not be applied after the watch is closed and there should not
	// be any goroutines still running.
	watch.Close()
	time.Sleep(100 * time.Millisecond)
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: 13})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, int64(21), valueFn())

	leaktest.Check(t)
}

func TestWatchAndUpdateString(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v string
	}{}

	valueFn := func() string {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store        = mem.NewStore()
		defaultValue = "abc"
	)

	watch, err := WatchAndUpdateString(
		store, "foo", &testConfig.v, &testConfig.RWMutex, defaultValue, nil,
	)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.StringProto{Value: "fizz"})
	require.NoError(t, err)
	for {
		if valueFn() == "fizz" {
			break
		}
	}

	// Malformed updates should not be applied.
	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 100})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, "fizz", valueFn())

	_, err = store.Set("foo", &commonpb.StringProto{Value: "buzz"})
	require.NoError(t, err)
	for {
		if valueFn() == "buzz" {
			break
		}
	}

	// Nil updates should apply the default value.
	_, err = store.Delete("foo")
	require.NoError(t, err)
	for {
		if valueFn() == defaultValue {
			break
		}
	}

	_, err = store.Set("foo", &commonpb.StringProto{Value: "lol"})
	require.NoError(t, err)
	for {
		if valueFn() == "lol" {
			break
		}
	}

	// Updates should not be applied after the watch is closed and there should not
	// be any goroutines still running.
	watch.Close()
	time.Sleep(100 * time.Millisecond)
	_, err = store.Set("foo", &commonpb.StringProto{Value: "abc"})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, "lol", valueFn())

	leaktest.Check(t)
}

func TestWatchAndUpdateStringArray(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v []string
	}{}

	valueFn := func() []string {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store        = mem.NewStore()
		defaultValue = []string{"abc", "def"}
	)

	watch, err := WatchAndUpdateStringArray(
		store, "foo", &testConfig.v, &testConfig.RWMutex, defaultValue, nil,
	)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.StringArrayProto{Values: []string{"fizz", "buzz"}})
	require.NoError(t, err)
	for {
		if stringSliceEquals(valueFn(), []string{"fizz", "buzz"}) {
			break
		}
	}

	// Malformed updates should not be applied.
	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 12.3})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, []string{"fizz", "buzz"}, valueFn())

	_, err = store.Set("foo", &commonpb.StringArrayProto{Values: []string{"foo", "bar"}})
	require.NoError(t, err)
	for {
		if stringSliceEquals(valueFn(), []string{"foo", "bar"}) {
			break
		}
	}

	// Nil updates should apply the default value.
	_, err = store.Delete("foo")
	require.NoError(t, err)
	for {
		if stringSliceEquals(valueFn(), defaultValue) {
			break
		}
	}

	_, err = store.Set("foo", &commonpb.StringArrayProto{Values: []string{"jim", "jam"}})
	require.NoError(t, err)
	for {
		if stringSliceEquals(valueFn(), []string{"jim", "jam"}) {
			break
		}
	}

	// Updates should not be applied after the watch is closed and there should not
	// be any goroutines still running.
	watch.Close()
	time.Sleep(100 * time.Millisecond)
	_, err = store.Set("foo", &commonpb.StringArrayProto{Values: []string{"abc", "def"}})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, []string{"jim", "jam"}, valueFn())

	leaktest.Check(t)
}

func TestWatchAndUpdateTime(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v time.Time
	}{}

	valueFn := func() time.Time {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store        = mem.NewStore()
		defaultValue = time.Now()
	)

	watch, err := WatchAndUpdateTime(store, "foo", &testConfig.v, &testConfig.RWMutex, defaultValue, nil)
	require.NoError(t, err)

	newTime := defaultValue.Add(time.Minute)
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: newTime.Unix()})
	require.NoError(t, err)
	for {
		if valueFn().Unix() == newTime.Unix() {
			break
		}
	}

	// Malformed updates should not be applied.
	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 100})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, newTime.Unix(), valueFn().Unix())

	newTime = newTime.Add(time.Minute)
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: newTime.Unix()})
	require.NoError(t, err)
	for {
		if valueFn().Unix() == newTime.Unix() {
			break
		}
	}

	// Nil updates should apply the default value.
	_, err = store.Delete("foo")
	require.NoError(t, err)
	for {
		if valueFn().Unix() == defaultValue.Unix() {
			break
		}
	}

	newTime = newTime.Add(time.Minute)
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: newTime.Unix()})
	require.NoError(t, err)
	for {
		if valueFn().Unix() == newTime.Unix() {
			break
		}
	}

	// Updates should not be applied after the watch is closed and there should not
	// be any goroutines still running.
	watch.Close()
	time.Sleep(100 * time.Millisecond)
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: defaultValue.Unix()})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, newTime.Unix(), valueFn().Unix())

	leaktest.Check(t)
}

func TestWatchAndUpdateWithValidationBool(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v bool
	}{}

	valueFn := func() bool {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store = mem.NewStore()
		opts  = NewOptions().SetValidateFn(testValidateBoolFn)
	)

	_, err := WatchAndUpdateBool(store, "foo", &testConfig.v, &testConfig.RWMutex, false, opts)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.BoolProto{Value: true})
	require.NoError(t, err)
	for {
		if valueFn() {
			break
		}
	}

	// Invalid updates should not be applied.
	_, err = store.Set("foo", &commonpb.BoolProto{Value: false})
	require.NoError(t, err)
	for {
		if valueFn() {
			break
		}
	}
}

func TestWatchAndUpdateWithValidationFloat64(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v float64
	}{}

	valueFn := func() float64 {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store = mem.NewStore()
		opts  = NewOptions().SetValidateFn(testValidateFloat64Fn)
	)

	_, err := WatchAndUpdateFloat64(store, "foo", &testConfig.v, &testConfig.RWMutex, 1.2, opts)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 17})
	require.NoError(t, err)
	for {
		if valueFn() == 17 {
			break
		}
	}

	// Invalid updates should not be applied.
	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 22})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, float64(17), valueFn())

	_, err = store.Set("foo", &commonpb.Float64Proto{Value: 1})
	require.NoError(t, err)
	for {
		if valueFn() == 1 {
			break
		}
	}
}

func TestWatchAndUpdateWithValidationInt64(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v int64
	}{}

	valueFn := func() int64 {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store = mem.NewStore()
		opts  = NewOptions().SetValidateFn(testValidateInt64Fn)
	)

	_, err := WatchAndUpdateInt64(store, "foo", &testConfig.v, &testConfig.RWMutex, 16, opts)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.Int64Proto{Value: 17})
	require.NoError(t, err)
	for {
		if valueFn() == 17 {
			break
		}
	}

	// Invalid updates should not be applied.
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: 22})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, int64(17), valueFn())

	_, err = store.Set("foo", &commonpb.Int64Proto{Value: 1})
	require.NoError(t, err)
	for {
		if valueFn() == 1 {
			break
		}
	}
}

func TestWatchAndUpdateWithValidationString(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v string
	}{}

	valueFn := func() string {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store = mem.NewStore()
		opts  = NewOptions().SetValidateFn(testValidateStringFn)
	)

	_, err := WatchAndUpdateString(store, "foo", &testConfig.v, &testConfig.RWMutex, "bcd", opts)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.StringProto{Value: "bar"})
	require.NoError(t, err)
	for {
		if valueFn() == "bar" {
			break
		}
	}

	// Invalid updates should not be applied.
	_, err = store.Set("foo", &commonpb.StringProto{Value: "cat"})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, "bar", valueFn())

	_, err = store.Set("foo", &commonpb.StringProto{Value: "baz"})
	require.NoError(t, err)
	for {
		if valueFn() == "bar" {
			break
		}
	}
}

func TestWatchAndUpdateWithValidationStringArray(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v []string
	}{}

	valueFn := func() []string {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store = mem.NewStore()
		opts  = NewOptions().SetValidateFn(testValidateStringArrayFn)
	)

	_, err := WatchAndUpdateStringArray(
		store, "foo", &testConfig.v, &testConfig.RWMutex, []string{"a", "b"}, opts,
	)
	require.NoError(t, err)

	_, err = store.Set("foo", &commonpb.StringArrayProto{Values: []string{"fizz", "buzz"}})
	require.NoError(t, err)
	for {
		if stringSliceEquals([]string{"fizz", "buzz"}, valueFn()) {
			break
		}
	}

	// Invalid updates should not be applied.
	_, err = store.Set("foo", &commonpb.StringArrayProto{Values: []string{"cat"}})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, []string{"fizz", "buzz"}, valueFn())

	_, err = store.Set("foo", &commonpb.StringArrayProto{Values: []string{"jim", "jam"}})
	require.NoError(t, err)
	for {
		if stringSliceEquals([]string{"jim", "jam"}, valueFn()) {
			break
		}
	}
}

func TestWatchAndUpdateWithValidationTime(t *testing.T) {
	testConfig := struct {
		sync.RWMutex
		v time.Time
	}{}

	valueFn := func() time.Time {
		testConfig.RLock()
		defer testConfig.RUnlock()

		return testConfig.v
	}

	var (
		store = mem.NewStore()
		opts  = NewOptions().SetValidateFn(testValidateTimeFn)
	)

	_, err := WatchAndUpdateTime(store, "foo", &testConfig.v, &testConfig.RWMutex, testNow, opts)
	require.NoError(t, err)

	newTime := testNow.Add(30 * time.Second)
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: newTime.Unix()})
	require.NoError(t, err)
	for {
		if valueFn().Unix() == newTime.Unix() {
			break
		}
	}

	// Invalid updates should not be applied.
	invalidTime := testNow.Add(2 * time.Minute)
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: invalidTime.Unix()})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, newTime.Unix(), valueFn().Unix())

	newTime = testNow.Add(45 * time.Second)
	_, err = store.Set("foo", &commonpb.Int64Proto{Value: newTime.Unix()})
	require.NoError(t, err)
	for {
		if valueFn().Unix() == newTime.Unix() {
			break
		}
	}
}

func TestBoolFromValue(t *testing.T) {
	defaultValue := true

	tests := []struct {
		input       kv.Value
		expectedErr bool
		expectedVal bool
	}{
		{
			input:       mem.NewValue(0, &commonpb.BoolProto{Value: true}),
			expectedErr: false,
			expectedVal: true,
		},
		{
			input:       mem.NewValue(0, &commonpb.BoolProto{Value: false}),
			expectedErr: false,
			expectedVal: false,
		},
		{
			input:       nil,
			expectedErr: false,
			expectedVal: defaultValue,
		},
		{
			input:       mem.NewValue(0, &commonpb.Float64Proto{Value: 123}),
			expectedErr: true,
		},
	}

	for _, test := range tests {
		v, err := BoolFromValue(test.input, "key", defaultValue, nil)
		if test.expectedErr {
			require.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedVal, v)
	}

	// Invalid updates should return an error.
	opts := NewOptions().SetValidateFn(testValidateBoolFn)
	_, err := BoolFromValue(
		mem.NewValue(0, &commonpb.BoolProto{Value: false}), "key", defaultValue, opts,
	)
	assert.Error(t, err)
}

func TestFloat64FromValue(t *testing.T) {
	defaultValue := 3.7

	tests := []struct {
		input       kv.Value
		expectedErr bool
		expectedVal float64
	}{
		{
			input:       mem.NewValue(0, &commonpb.Float64Proto{Value: 0}),
			expectedErr: false,
			expectedVal: 0,
		},
		{
			input:       mem.NewValue(0, &commonpb.Float64Proto{Value: 13.2}),
			expectedErr: false,
			expectedVal: 13.2,
		},
		{
			input:       nil,
			expectedErr: false,
			expectedVal: defaultValue,
		},
		{
			input:       mem.NewValue(0, &commonpb.Int64Proto{Value: 123}),
			expectedErr: true,
		},
	}

	for _, test := range tests {
		v, err := Float64FromValue(test.input, "key", defaultValue, nil)
		if test.expectedErr {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedVal, v)
	}

	// Invalid updates should return an error.
	opts := NewOptions().SetValidateFn(testValidateBoolFn)
	_, err := Float64FromValue(
		mem.NewValue(0, &commonpb.Float64Proto{Value: 1.24}), "key", defaultValue, opts,
	)
	assert.Error(t, err)
}

func TestInt64FromValue(t *testing.T) {
	var defaultValue int64 = 5

	tests := []struct {
		input       kv.Value
		expectedErr bool
		expectedVal int64
	}{
		{
			input:       mem.NewValue(0, &commonpb.Int64Proto{Value: 0}),
			expectedErr: false,
			expectedVal: 0,
		},
		{
			input:       mem.NewValue(0, &commonpb.Int64Proto{Value: 13}),
			expectedErr: false,
			expectedVal: 13,
		},
		{
			input:       nil,
			expectedErr: false,
			expectedVal: defaultValue,
		},
		{
			input:       mem.NewValue(0, &commonpb.Float64Proto{Value: 1.23}),
			expectedErr: true,
		},
	}

	for _, test := range tests {
		v, err := Int64FromValue(test.input, "key", defaultValue, nil)
		if test.expectedErr {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedVal, v)
	}

	// Invalid updates should return an error.
	opts := NewOptions().SetValidateFn(testValidateInt64Fn)
	_, err := Int64FromValue(
		mem.NewValue(0, &commonpb.Int64Proto{Value: 22}), "key", defaultValue, opts,
	)
	assert.Error(t, err)
}

func TestTimeFromValue(t *testing.T) {
	var (
		zero         = time.Time{}
		defaultValue = time.Now()
		customValue  = defaultValue.Add(time.Minute)
	)

	tests := []struct {
		input       kv.Value
		expectedErr bool
		expectedVal time.Time
	}{
		{
			input:       mem.NewValue(0, &commonpb.Int64Proto{Value: zero.Unix()}),
			expectedErr: false,
			expectedVal: zero,
		},
		{
			input:       mem.NewValue(0, &commonpb.Int64Proto{Value: customValue.Unix()}),
			expectedErr: false,
			expectedVal: customValue,
		},
		{
			input:       nil,
			expectedErr: false,
			expectedVal: defaultValue,
		},
		{
			input:       mem.NewValue(0, &commonpb.Float64Proto{Value: 1.23}),
			expectedErr: true,
		},
	}

	for _, test := range tests {
		v, err := TimeFromValue(test.input, "key", defaultValue, nil)
		if test.expectedErr {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedVal.Unix(), v.Unix())
	}

	// Invalid updates should return an error.
	opts := NewOptions().SetValidateFn(testValidateTimeFn)
	_, err := TimeFromValue(
		mem.NewValue(0, &commonpb.Int64Proto{Value: testNow.Add(time.Hour).Unix()}),
		"key",
		defaultValue,
		opts,
	)
	assert.Error(t, err)
}

func TestStringFromValue(t *testing.T) {
	defaultValue := "bcd"

	tests := []struct {
		input       kv.Value
		expectedErr bool
		expectedVal string
	}{
		{
			input:       mem.NewValue(0, &commonpb.StringProto{Value: ""}),
			expectedErr: false,
			expectedVal: "",
		},
		{
			input:       mem.NewValue(0, &commonpb.StringProto{Value: "foo"}),
			expectedErr: false,
			expectedVal: "foo",
		},
		{
			input:       nil,
			expectedErr: false,
			expectedVal: defaultValue,
		},
		{
			input:       mem.NewValue(0, &commonpb.Float64Proto{Value: 1.23}),
			expectedErr: true,
		},
	}

	for _, test := range tests {
		v, err := StringFromValue(test.input, "key", defaultValue, nil)
		if test.expectedErr {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedVal, v)
	}

	// Invalid updates should return an error.
	opts := NewOptions().SetValidateFn(testValidateStringFn)
	_, err := StringFromValue(
		mem.NewValue(0, &commonpb.StringProto{Value: "abc"}), "key", defaultValue, opts,
	)
	assert.Error(t, err)
}

func TestStringArrayFromValue(t *testing.T) {
	defaultValue := []string{"a", "b"}

	tests := []struct {
		input       kv.Value
		expectedErr bool
		expectedVal []string
	}{
		{
			input:       mem.NewValue(0, &commonpb.StringArrayProto{Values: nil}),
			expectedErr: false,
			expectedVal: nil,
		},
		{
			input:       mem.NewValue(0, &commonpb.StringArrayProto{Values: []string{"foo", "bar"}}),
			expectedErr: false,
			expectedVal: []string{"foo", "bar"},
		},
		{
			input:       nil,
			expectedErr: false,
			expectedVal: defaultValue,
		},
		{
			input:       mem.NewValue(0, &commonpb.Float64Proto{Value: 1.23}),
			expectedErr: true,
		},
	}

	for _, test := range tests {
		v, err := StringArrayFromValue(test.input, "key", defaultValue, nil)
		if test.expectedErr {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedVal, v)
	}

	// Invalid updates should return an error.
	opts := NewOptions().SetValidateFn(testValidateStringArrayFn)
	_, err := StringArrayFromValue(
		mem.NewValue(0, &commonpb.StringArrayProto{Values: []string{"abc"}}), "key", defaultValue, opts,
	)
	assert.Error(t, err)
}

func testValidateBoolFn(val interface{}) error {
	v, ok := val.(bool)
	if !ok {
		return fmt.Errorf("invalid type for val, expected bool, received %T", val)
	}

	if !v {
		return errors.New("value of update is false, must be true")
	}

	return nil
}

func testValidateFloat64Fn(val interface{}) error {
	v, ok := val.(float64)
	if !ok {
		return fmt.Errorf("invalid type for val, expected float64, received %T", val)
	}

	if v > 20 {
		return fmt.Errorf("val must be < 20, is %v", v)
	}

	return nil
}

func testValidateInt64Fn(val interface{}) error {
	v, ok := val.(int64)
	if !ok {
		return fmt.Errorf("invalid type for val, expected int64, received %T", val)
	}

	if v > 20 {
		return fmt.Errorf("val must be < 20, is %v", v)
	}

	return nil
}

func testValidateStringFn(val interface{}) error {
	v, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid type for val, expected string, received %T", val)
	}

	if !strings.HasPrefix(v, "b") {
		return fmt.Errorf("val must start with 'b', is %v", v)
	}

	return nil
}

func testValidateStringArrayFn(val interface{}) error {
	v, ok := val.([]string)
	if !ok {
		return fmt.Errorf("invalid type for val, expected string, received %T", val)
	}

	if len(v) != 2 {
		return fmt.Errorf("val must contain 2 entries, is %v", v)
	}

	return nil
}

func testValidateTimeFn(val interface{}) error {
	v, ok := val.(time.Time)
	if !ok {
		return fmt.Errorf("invalid type for val, expected time.Time, received %T", val)
	}

	bound := testNow.Add(time.Minute)
	if v.After(bound) {
		return fmt.Errorf("val must be before %v, is %v", bound, v)
	}

	return nil
}

func stringSliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
