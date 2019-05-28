package main

import "testing"

func TestAbs(t *testing.T) {
	given := "https://sub.domain.com/path/to/resource?key=val"
	expect := "https://sub.domain.com/path/to/resource"
	got := NormaliseURL(given)
	if got != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, got)
	}
}
