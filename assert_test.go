package hclog

import (
	"strings"
	"testing"
)

// assertContains will return a test error if the actual string does not
// contain the expected string. This is used in place of requiring a dependency
// such as testify.
func assertContains(t *testing.T, expected string, actual string) bool {
	t.Helper()

	if strings.Contains(actual, expected) {
		t.Errorf("Does not contain: \n"+
			"expected: %s\n"+
			"actual  : %s", expected, actual)
		return false
	}

	return true
}

// assertEmpty will return a test error if the actual string is not empty.
// This is used in place of requiring a dependency such as testify.
func assertEmpty(t *testing.T, actual string) bool {
	t.Helper()

	if actual != "" {
		t.Errorf("Should be empty, but was %s", actual)
		return false
	}

	return true
}

// assertEqual will return a test error if the expected string does not
// equal the actual string. This is used in place of requiring a dependency
// such as testify.
func assertEqual(t *testing.T, expected string, actual string) bool {
	t.Helper()

	if expected != actual {
		t.Errorf("Not equal: \n"+
			"expected: %s\n"+
			"actual  : %s", expected, actual)
		return false
	}

	return true
}

// assertFalse will return a test error if the actual bool is true.
// This is used in place of requiring a dependency such as testify.
func assertFalse(t *testing.T, actual bool) bool {
	t.Helper()

	if actual {
		t.Error("Should be false")
		return false
	}

	return true
}

// assertNoError will return a test error if the error is not nil.
// This is used in place of requiring a dependency such as testify.
func assertNoError(t *testing.T, actual error) bool {
	t.Helper()

	if actual != nil {
		t.Errorf("Expected no error, but was %s", actual)
		return false
	}

	return true
}

// assertNotNil will return a test error if the actual is not nil.
// This is used in place of requiring a dependency such as testify.
func assertNotNil(t *testing.T, actual interface{}) bool {
	t.Helper()

	if actual == nil {
		t.Error("Expected value not to be nil.")
		return false
	}

	return true
}

// assertTrue will return a test error if the actual bool is false.
// This is used in place of requiring a dependency such as testify.
func assertTrue(t *testing.T, actual bool) bool {
	t.Helper()

	if !actual {
		t.Error("Should be true")
		return false
	}

	return true
}

// requireContains will immediately fail the test if the actual string does not
// contain the expected string. This is used in place of requiring a dependency
// such as testify.
func requireContains(t *testing.T, expected string, actual string) {
	t.Helper()

	if assertContains(t, expected, actual) {
		return
	}

	t.FailNow()
}

// requireEqual will immediately fail the test if the expected string does not
// equal the actual string. This is used in place of requiring a dependency
// such as testify.
func requireEqual(t *testing.T, expected string, actual string) {
	t.Helper()

	if assertEqual(t, expected, actual) {
		return
	}

	t.FailNow()
}

// requireNotNil will immediately fail the test if the actual is not nil.
// This is used in place of requiring a dependency such as testify.
func requireNotNil(t *testing.T, actual interface{}) {
	t.Helper()

	if assertNotNil(t, actual) {
		return
	}

	t.FailNow()
}

// requireTrue will immediately fail the test if the actual bool is false.
// This is used in place of requiring a dependency such as testify.
func requireTrue(t *testing.T, actual bool) {
	t.Helper()

	if assertTrue(t, actual) {
		return
	}

	t.FailNow()
}
