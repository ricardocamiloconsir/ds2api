package assert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func withMessage(base string, msgAndArgs ...any) string {
	if len(msgAndArgs) == 0 {
		return base
	}
	if format, ok := msgAndArgs[0].(string); ok {
		if len(msgAndArgs) > 1 {
			return base + ": " + fmt.Sprintf(format, msgAndArgs[1:]...)
		}
		return base + ": " + format
	}
	return base + ": " + fmt.Sprint(msgAndArgs...)
}

func Equal(t *testing.T, expected, actual any, msgAndArgs ...any) bool {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(withMessage(fmt.Sprintf("not equal: expected=%v actual=%v", expected, actual), msgAndArgs...))
		return false
	}
	return true
}

func True(t *testing.T, value bool, msgAndArgs ...any) bool {
	t.Helper()
	if !value {
		t.Errorf(withMessage("expected true", msgAndArgs...))
		return false
	}
	return true
}

func False(t *testing.T, value bool, msgAndArgs ...any) bool {
	t.Helper()
	if value {
		t.Errorf(withMessage("expected false", msgAndArgs...))
		return false
	}
	return true
}

func NoError(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	if err != nil {
		t.Errorf(withMessage(fmt.Sprintf("expected no error, got %v", err), msgAndArgs...))
		return false
	}
	return true
}

func Error(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	if err == nil {
		t.Errorf(withMessage("expected error", msgAndArgs...))
		return false
	}
	return true
}

func NotEmpty(t *testing.T, v any, msgAndArgs ...any) bool {
	t.Helper()
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		t.Errorf(withMessage("expected not empty", msgAndArgs...))
		return false
	}
	ok := true
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		ok = rv.Len() > 0
	default:
		ok = !rv.IsZero()
	}
	if !ok {
		t.Errorf(withMessage("expected not empty", msgAndArgs...))
		return false
	}
	return true
}

func NotZero(t *testing.T, v any, msgAndArgs ...any) bool {
	t.Helper()
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.IsZero() {
		t.Errorf(withMessage("expected non-zero", msgAndArgs...))
		return false
	}
	return true
}

func Contains(t *testing.T, s any, contains any, msgAndArgs ...any) bool {
	t.Helper()
	sv := reflect.ValueOf(s)
	if !sv.IsValid() {
		t.Errorf(withMessage(fmt.Sprintf("expected %v to contain %v", s, contains), msgAndArgs...))
		return false
	}
	switch sv.Kind() {
	case reflect.String:
		containsStr, ok := contains.(string)
		if !ok {
			t.Errorf(withMessage(fmt.Sprintf("Contains: expected string contains argument for string subject, got %T", contains), msgAndArgs...))
			return false
		}
		if !strings.Contains(s.(string), containsStr) {
			t.Errorf(withMessage(fmt.Sprintf("expected %v to contain %v", s, contains), msgAndArgs...))
			return false
		}
		return true
	case reflect.Slice, reflect.Array:
		for i := 0; i < sv.Len(); i++ {
			if reflect.DeepEqual(sv.Index(i).Interface(), contains) {
				return true
			}
		}
	}
	t.Errorf(withMessage(fmt.Sprintf("expected %v to contain %v", s, contains), msgAndArgs...))
	return false
}

func NotContains(t *testing.T, s any, contains any, msgAndArgs ...any) bool {
	t.Helper()
	sv := reflect.ValueOf(s)
	switch sv.Kind() {
	case reflect.String:
		if strings.Contains(s.(string), contains.(string)) {
			t.Errorf(withMessage(fmt.Sprintf("expected %v not to contain %v", s, contains), msgAndArgs...))
			return false
		}
		return true
	case reflect.Slice, reflect.Array:
		for i := 0; i < sv.Len(); i++ {
			if reflect.DeepEqual(sv.Index(i).Interface(), contains) {
				t.Errorf(withMessage(fmt.Sprintf("expected %v not to contain %v", s, contains), msgAndArgs...))
				return false
			}
		}
		return true
	}
	return true
}

func GreaterOrEqual[T ~int](t *testing.T, actual, expected T, msgAndArgs ...any) bool {
	t.Helper()
	if actual < expected {
		t.Errorf(withMessage(fmt.Sprintf("expected %v >= %v", actual, expected), msgAndArgs...))
		return false
	}
	return true
}

func NotNil(t *testing.T, v any, msgAndArgs ...any) bool {
	t.Helper()
	if v == nil {
		t.Errorf(withMessage("expected non-nil", msgAndArgs...))
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		if rv.IsNil() {
			t.Errorf(withMessage("expected non-nil", msgAndArgs...))
			return false
		}
	}
	return true
}

func NotEqual(t *testing.T, expected, actual any, msgAndArgs ...any) bool {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Errorf(withMessage(fmt.Sprintf("expected values to be different: %v", actual), msgAndArgs...))
		return false
	}
	return true
}
