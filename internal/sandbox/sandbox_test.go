package sandbox

import "testing"

func TestApply(t *testing.T) {
	if err := Apply("70000"); err == nil {
		t.Fatal("expected invalid port to fail")
	}

	if err := Apply("8080"); err != nil {
		t.Fatalf("apply: %v", err)
	}
}
