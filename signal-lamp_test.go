package main

import "testing"

func TestNextMode(t *testing.T) {
	if next := nextMode("a", []string{"a", "b"}); next != "b" {
		t.Errorf("wanted b, got %s", next)
	}

	if next := nextMode("b", []string{"a", "b"}); next != "a" {
		t.Errorf("wanted a, got %s", next)
	}
}
