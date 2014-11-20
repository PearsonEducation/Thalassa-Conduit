package main

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Assert provides functions for performing assertions when testing.
type Assert struct{}

var assert = Assert{}

// gets the location in the test code of an assertion failure
func (as Assert) getLoc() string {
	file := ""
	line := 0

	var ok bool
	for i := 0; ; i++ {
		_, file, line, ok = runtime.Caller(i)
		if !ok {
			return ""
		}
		parts := strings.Split(file, "/")
		file = parts[len(parts)-1]
		if len(parts) > 1 {
			if dir := parts[len(parts)-2]; dir != "assert" {
				break
			}
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// uses T.log to output failure messages in an easy to read format
func (as Assert) log(t *testing.T, failureMsg, userMsg string) {
	format := "\r\tLocation:  %s\n\r\tError:     %s\n\n"
	formatMsg := "\r\tLocation:  %s\n\r\tError:     %s\n\r\tMessage:   %s\n\n"

	msg := strings.Replace(failureMsg, "\n", "\n\r\t           ", 10)
	if len(userMsg) > 0 {
		t.Logf(formatMsg, as.getLoc(), msg, userMsg)
	} else {
		t.Logf(format, as.getLoc(), msg)
	}
}

// Equal asserts that the given actual and expected instances are equal. If they aren't, it
// logs an error but continues executing the test.
func (as Assert) Equal(t *testing.T, act, exp interface{}, format string, a ...interface{}) bool {
	return as.assertEqual(t, act, exp, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureEqual asserts that the given actual and expected instances are equal. If they aren't, it
// logs an error and stops test execution.
func (as Assert) EnsureEqual(t *testing.T, act, exp interface{}, format string, a ...interface{}) bool {
	return as.assertEqual(t, act, exp, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertEqual(t *testing.T, act, exp interface{}, msg string, fail func()) bool {
	eq := as.areEqual(act, exp)
	if !eq {
		str := fmt.Sprintf("Not Equal:\n   actual %#v\n expected %#v", act, exp)
		as.log(t, str, msg)
		fail()
	}
	return eq
}

// NotEqual asserts that the given actual and expected instances are not equal. If they are, it
// logs an error but continues executing the test.
func (as Assert) NotEqual(t *testing.T, act, exp interface{}, format string, a ...interface{}) bool {
	return as.assertNotEqual(t, act, exp, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureNotEqual asserts that the given actual and expected instances are not equal. If they are, it
// logs an error and stops test execution.
func (as Assert) EnsureNotEqual(t *testing.T, act, exp interface{}, format string, a ...interface{}) bool {
	return as.assertNotEqual(t, act, exp, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertNotEqual(t *testing.T, act, exp interface{}, msg string, fail func()) bool {
	eq := as.areEqual(act, exp)
	if eq {
		as.log(t, "Should not be equal", msg)
		fail()
	}
	return !eq
}

func (as Assert) areEqual(act, exp interface{}) bool {
	if act == nil || exp == nil {
		return act == exp
	}

	if reflect.DeepEqual(act, exp) {
		return true
	}

	actValue := reflect.ValueOf(act)
	expValue := reflect.ValueOf(exp)
	if actValue == expValue {
		return true
	}

	t := expValue.Type()
	if actValue.Type().ConvertibleTo(t) && actValue.Convert(t) == expValue {
		return true
	}

	return false
}

// Same asserts that the given actual and expected instances are equal and of the same type.
// If they aren't, it logs an error but continues executing the test.
func (as Assert) Same(t *testing.T, act, exp interface{}, format string, a ...interface{}) bool {
	return as.assertSame(t, act, exp, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureSame asserts that the given actual and expected instances are equal and of the same type.
// If they aren't, it logs an error and stops test execution.
func (as Assert) EnsureSame(t *testing.T, act, exp interface{}, format string, a ...interface{}) bool {
	return as.assertSame(t, act, exp, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertSame(t *testing.T, act, exp interface{}, msg string, fail func()) bool {
	same := as.areSame(act, exp)
	if !same {
		str := fmt.Sprintf("Not Same:\n   actual %#v\n expected %#v", act, exp)
		as.log(t, str, msg)
		fail()
	}

	return same
}

// NotSame asserts that the given actual and expected instances are not equal or not of the same type.
// If they are, it logs an error but continues executing the test.
func (as Assert) NotSame(t *testing.T, act, exp interface{}, format string, a ...interface{}) bool {
	return as.assertNotSame(t, act, exp, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureNotSame asserts that the given actual and expected instances are not equal or not of the same type.
// If they are, it logs an error and stops test execution.
func (as Assert) EnsureNotSame(t *testing.T, act, exp interface{}, format string, a ...interface{}) bool {
	return as.assertNotSame(t, act, exp, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertNotSame(t *testing.T, act, exp interface{}, msg string, fail func()) bool {
	same := as.areSame(act, exp)
	if same {
		as.log(t, "Should be the same (type, value)", msg)
		fail()
	}
	return !same
}

func (as Assert) areSame(act, exp interface{}) bool {
	actType := reflect.TypeOf(act)
	expType := reflect.TypeOf(exp)
	if actType != expType {
		return false
	}
	return as.areEqual(act, exp)
}

// Nil asserts that the given object is nil. If it isn't, it logs an error but continues executing the test.
func (as Assert) Nil(t *testing.T, obj interface{}, format string, a ...interface{}) bool {
	return as.assertNil(t, obj, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureNil asserts that the given object is nil. If it isn't, it logs an error and stops executing the test.
func (as Assert) EnsureNil(t *testing.T, obj interface{}, format string, a ...interface{}) bool {
	return as.assertNil(t, obj, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertNil(t *testing.T, obj interface{}, msg string, fail func()) bool {
	n := as.isNil(obj)
	if !n {
		as.log(t, fmt.Sprintf("Not Nil: actual '%#v'", obj), msg)
		fail()
	}
	return n
}

// NotNil asserts that the given object is not nil. If it is, it logs an error but continues executing the test.
func (as Assert) NotNil(t *testing.T, obj interface{}, format string, a ...interface{}) bool {
	return as.assertNotNil(t, obj, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureNotNil asserts that the given object is nil. If it isn't, it logs an error and stops executing the test.
func (as Assert) EnsureNotNil(t *testing.T, obj interface{}, format string, a ...interface{}) bool {
	return as.assertNotNil(t, obj, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertNotNil(t *testing.T, obj interface{}, msg string, fail func()) bool {
	n := as.isNil(obj)
	if n {
		as.log(t, "Nil, but expected a value", msg)
		fail()
	}
	return !n
}

func (as Assert) isNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	value := reflect.ValueOf(obj)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

// Empty asserts that the given object is empty. If it isn't, it logs an error but continues executing the test.
// An empty object is one that is nil, empty (strings, collections), false (booleans), or 0 (numeric types).
func (as Assert) Empty(t *testing.T, obj interface{}, format string, a ...interface{}) bool {
	return as.assertEmpty(t, obj, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureEmpty asserts that the given object is empty. If it isn't, it logs an error and stops executing the test.
// An empty object is one that is nil, empty (strings, collections), false (booleans), or 0 (numeric types).
func (as Assert) EnsureEmpty(t *testing.T, obj interface{}, format string, a ...interface{}) bool {
	return as.assertEmpty(t, obj, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertEmpty(t *testing.T, obj interface{}, msg string, fail func()) bool {
	n := as.isEmpty(obj)
	if !n {
		as.log(t, fmt.Sprintf("Expected empty, got '%#v'", obj), msg)
		fail()
	}
	return n
}

// NotEmpty asserts that the given object is not empty. If it is, it logs an error but continues executing the test.
// An empty object is one that is nil, empty (strings, collections), false (booleans), or 0 (numeric types).
func (as Assert) NotEmpty(t *testing.T, obj interface{}, format string, a ...interface{}) bool {
	return as.assertNotEmpty(t, obj, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureNotEmpty asserts that the given object is not empty. If it is, it logs an error and stops executing the test.
// An empty object is one that is nil, empty (strings, collections), false (booleans), or 0 (numeric types).
func (as Assert) EnsureNotEmpty(t *testing.T, obj interface{}, format string, a ...interface{}) bool {
	return as.assertNotEmpty(t, obj, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertNotEmpty(t *testing.T, obj interface{}, msg string, fail func()) bool {
	n := as.isEmpty(obj)
	if n {
		as.log(t, "Empty, but expected a value", msg)
		fail()
	}
	return !n
}

func (as Assert) isEmpty(obj interface{}) bool {
	if obj == nil {
		return true
	}
	if obj == "" {
		return true
	}
	if obj == false {
		return true
	}

	if f, err := as.getFloat(obj); err == nil {
		if f == float64(0) {
			return true
		}
	}

	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Chan:
		return v.Len() == 0
	case reflect.Ptr:
		switch obj.(type) {
		case *time.Time:
			return obj.(*time.Time).IsZero()
		default:
			return false
		}
	}
	return false
}

func (as Assert) getFloat(obj interface{}) (float64, error) {
	v := reflect.ValueOf(obj)
	v = reflect.Indirect(v)
	floatType := reflect.TypeOf(float64(0))
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert to float64")
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

// StringContains asserts that the given string contains the given substring. If it doesn't, it
// logs an error but continues executing the test.
func (as Assert) StringContains(t *testing.T, str, substr, format string, a ...interface{}) bool {
	return as.assertStringContains(t, str, substr, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureStringContains asserts that the given string contains the given substring. If it doesn't, it
// logs an error and stops executing the test.
func (as Assert) EnsureStringContains(t *testing.T, str, substr, format string, a ...interface{}) bool {
	return as.assertStringContains(t, str, substr, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertStringContains(t *testing.T, str, substr, msg string, fail func()) bool {
	c := strings.Contains(str, substr)
	if !c {
		s := fmt.Sprintf("'%s' does not contain '%s'", str, substr)
		as.log(t, s, msg)
		fail()
	}
	return c
}

// NotStringContains asserts that the given string does not contain the given substring. If it does, it
// logs an error but continues executing the test.
func (as Assert) NotStringContains(t *testing.T, str, substr, format string, a ...interface{}) bool {
	return as.assertStringNotContains(t, str, substr, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureNotStringContains asserts that the given string does not contain the given substring. If it does, it
// logs an error and stops executing the test.
func (as Assert) EnsureNotStringContains(t *testing.T, str, substr, format string, a ...interface{}) bool {
	return as.assertStringNotContains(t, str, substr, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertStringNotContains(t *testing.T, str, substr, msg string, fail func()) bool {
	c := strings.Contains(str, substr)
	if c {
		s := fmt.Sprintf("'%s' should not contain '%s'", str, substr)
		as.log(t, s, msg)
		fail()
	}
	return !c
}

// True asserts that the given boolean value is true. If it isn't, it logs an error
// but continues executing the test.
func (as Assert) True(t *testing.T, value bool, format string, a ...interface{}) bool {
	return as.assertTrue(t, value, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureTrue asserts that the given boolean value is true. If it isn't, it logs an error
// and stops executing the test.
func (as Assert) EnsureTrue(t *testing.T, value bool, format string, a ...interface{}) bool {
	return as.assertTrue(t, value, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertTrue(t *testing.T, value bool, msg string, fail func()) bool {
	if !value {
		as.log(t, "Should be true", msg)
		fail()
	}
	return value
}

// False asserts that the given boolean value is false. If it isn't, it logs an error
// but continues executing the test.
func (as Assert) False(t *testing.T, value bool, format string, a ...interface{}) bool {
	return as.assertFalse(t, value, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureFalse asserts that the given boolean value is false. If it isn't, it logs an error
// and stops executing the test.
func (as Assert) EnsureFalse(t *testing.T, value bool, format string, a ...interface{}) bool {
	return as.assertFalse(t, value, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertFalse(t *testing.T, value bool, msg string, fail func()) bool {
	if value {
		as.log(t, "Should be false", msg)
		fail()
	}
	return !value
}

// Panic asserts that the given func (as Assert)tion triggers a panic. If it doesn't, it logs an error
// but continues executing the test.
func (as Assert) Panic(t *testing.T, f func(), format string, a ...interface{}) bool {
	return as.assertPanic(t, f, fmt.Sprintf(format, a...), t.Fail)
}

// EnsurePanic asserts that the given func (as Assert)tion triggers a panic. If it doesn't, it logs an error
// and stops executing the test.
func (as Assert) EnsurePanic(t *testing.T, f func(), format string, a ...interface{}) bool {
	return as.assertPanic(t, f, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertPanic(t *testing.T, f func(), msg string, fail func()) bool {
	var panicObj interface{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicObj = r
			}
		}()
		f()
	}()

	if panicObj == nil {
		as.log(t, "Expected panic was not triggered", msg)
		fail()
		return false
	}
	return true
}

// NotPanic asserts that the given func (as Assert)tion does not trigger a panic. If it does, it logs an error
// but continues executing the test.
func (as Assert) NotPanic(t *testing.T, f func(), format string, a ...interface{}) bool {
	return as.assertNotPanic(t, f, fmt.Sprintf(format, a...), t.Fail)
}

// EnsureNotPanic asserts that the given func (as Assert)tion does not trigger a panic. If it does, it logs an error
// and stops executing the test.
func (as Assert) EnsureNotPanic(t *testing.T, f func(), format string, a ...interface{}) bool {
	return as.assertNotPanic(t, f, fmt.Sprintf(format, a...), t.FailNow)
}

func (as Assert) assertNotPanic(t *testing.T, f func(), msg string, fail func()) bool {
	var panicObj interface{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicObj = r
			}
		}()
		f()
	}()

	if panicObj != nil {
		str := fmt.Sprintf("Unexpected panic was triggered: %#v", panicObj)
		as.log(t, str, msg)
		fail()
		return false
	}
	return true
}
