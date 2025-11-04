// package testutils contains assertions with struct [expect]
//
// Adapted from https://www.alexedwards.net/blog/the-9-go-test-assertions-i-use
package testutils

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

type expect struct{}

// Expect is the exported instance users call.
var Expect = expect{}

// Equal asserts that got and want are equal (using reflect.DeepEqual where necessary).
func (e expect) Equal(t *testing.T, got, want any, args ...any) {
	t.Helper()
	if !isDeepEqual(got, want) {
		t.Errorf("Equal assertion failed: got %#v; want %#v", got, want)
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// NotEqual asserts that got and want are *not* equal.
func (e expect) NotEqual(t *testing.T, got, want any, args ...any) {
	t.Helper()
	if isDeepEqual(got, want) {
		t.Errorf("NotEqual assertion failed: got %#v; expected different value", got)
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// True asserts that the boolean got is true.
func (e expect) True(t *testing.T, got bool, args ...any) {
	t.Helper()
	if !got {
		t.Errorf("True assertion failed: got false; want true")
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// False asserts that the boolean got is false.
func (e expect) False(t *testing.T, got bool, args ...any) {
	t.Helper()
	if got {
		t.Errorf("False assertion failed: got true; want false")
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// Nil asserts that got is nil.
func (e expect) Nil(t *testing.T, got any, args ...any) {
	t.Helper()
	if !isNil(got) {
		t.Errorf("Nil assertion failed: got non-nil value %#v", got)
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// NotNil asserts that got is *not* nil.
func (e expect) NotNil(t *testing.T, got any, args ...any) {
	t.Helper()
	if isNil(got) {
		t.Errorf("NotNil assertion failed: got nil; want non-nil")
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// ErrorIs asserts that err wraps or is target.
func (e expect) ErrorIs(t *testing.T, err, target error, args ...any) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("ErrorIs assertion failed: got error %#v; want error matching %#v", err, target)
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// ErrorAs asserts that err can be assigned to target via [errors.As].
func (e expect) ErrorAs(t *testing.T, err error, target any, args ...any) {
	t.Helper()
	if err == nil {
		t.Errorf("ErrorAs assertion failed: got nil; want assignable to %T", target)
		return
	}
	if !errors.As(err, &target) {
		t.Errorf("ErrorAs assertion failed: got error %#v; want assignable to %T", err, target)
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// MatchesRegexp asserts that the string got matches the given regex pattern.
func (e expect) MatchesRegexp(t *testing.T, got, pattern string, args ...any) {
	t.Helper()
	matched, err := regexp.MatchString(pattern, got)
	if err != nil {
		t.Fatalf("MatchesRegexp assertion: invalid pattern %q: %v", pattern, err)
		return
	}
	if !matched {
		t.Errorf("MatchesRegexp assertion failed: got %#v; want to match pattern %q", got, pattern)
		if len(args) > 0 {
			t.Logf("Message: %v", fmt.Sprint(args...))
		}
	}
}

// isDeepEqual handles deep equality including nil checks.
func isDeepEqual(a, b any) bool {
	if isNil(a) && isNil(b) {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// isNil tests nil for interface, pointer, slice, map, chan, func types.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	default:
		return false
	}
}
