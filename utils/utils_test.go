package utils

import (
	"testing"
)

func TestGetRandomString(t *testing.T) {
	result := len(RandomString(10))
	if result != 10 {
		t.Errorf(" RandomString(10) = %d; want 10", result)
	}
}
