package assert

import (
	"reflect"
	"strings"
	"testing"
)

func Equal(t *testing.T, expected, actual any, msgAndArgs ...any) bool {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("not equal: expected=%v actual=%v", expected, actual)
		return false
	}
	return true
}

func True(t *testing.T, value bool, msgAndArgs ...any) bool {
	t.Helper()
	if !value {
		t.Errorf("expected true")
		return false
	}
	return true
}

func False(t *testing.T, value bool, msgAndArgs ...any) bool {
	t.Helper()
	if value {
		t.Errorf("expected false")
		return false
	}
	return true
}

func NoError(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		return false
	}
	return true
}

func Error(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	if err == nil {
		t.Errorf("expected error")
		return false
	}
	return true
}

func NotEmpty(t *testing.T, v any, msgAndArgs ...any) bool {
	t.Helper()
	rv := reflect.ValueOf(v)
	ok := true
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		ok = rv.Len() > 0
	default:
		ok = !rv.IsZero()
	}
	if !ok {
		t.Errorf("expected not empty")
		return false
	}
	return true
}

func NotZero(t *testing.T, v any, msgAndArgs ...any) bool {
	t.Helper()
	if reflect.ValueOf(v).IsZero() {
		t.Errorf("expected non-zero")
		return false
	}
	return true
}

func Contains(t *testing.T, s any, contains any, msgAndArgs ...any) bool {
	t.Helper()
	sv := reflect.ValueOf(s)
	switch sv.Kind() {
	case reflect.String:
		if !strings.Contains(s.(string), contains.(string)) {
			t.Errorf("expected %v to contain %v", s, contains)
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
	t.Errorf("expected %v to contain %v", s, contains)
	return false
}

func NotContains(t *testing.T, s any, contains any, msgAndArgs ...any) bool {
	t.Helper()
	sv := reflect.ValueOf(s)
	switch sv.Kind() {
	case reflect.String:
		if strings.Contains(s.(string), contains.(string)) {
			t.Errorf("expected %v not to contain %v", s, contains)
			return false
		}
		return true
	case reflect.Slice, reflect.Array:
		for i := 0; i < sv.Len(); i++ {
			if reflect.DeepEqual(sv.Index(i).Interface(), contains) {
				t.Errorf("expected %v not to contain %v", s, contains)
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
		t.Errorf("expected %v >= %v", actual, expected)
		return false
	}
	return true
}

func NotNil(t *testing.T, v any, msgAndArgs ...any) bool {
	t.Helper()
	if v == nil {
		t.Errorf("expected non-nil")
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		if rv.IsNil() {
			t.Errorf("expected non-nil")
			return false
		}
	}
	return true
}

func NotEqual(t *testing.T, expected, actual any, msgAndArgs ...any) bool {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("expected values to be different: %v", actual)
		return false
	}
	return true
}
