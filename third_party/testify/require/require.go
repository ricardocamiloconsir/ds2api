package require

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func NoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if !assert.NoError(t, err, msgAndArgs...) {
		t.FailNow()
	}
}

func True(t *testing.T, value bool, msgAndArgs ...any) {
	t.Helper()
	if !assert.True(t, value, msgAndArgs...) {
		t.FailNow()
	}
}
