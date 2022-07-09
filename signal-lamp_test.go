package main

import "testing"

func TestNextModeCycle(t *testing.T) {
	c := &config{
		currentMode: "b",
		order: []string{
			"a", "b",
		},
	}
	if err := c.next(); err != nil {
		t.Fatalf("couldn't next(): %s", err)
	}

	if c.currentMode != "a" {
		t.Errorf("wanted a, got %s", c.currentMode)
	}

	if err := c.next(); err != nil {
		t.Fatalf("couldn't next(): %s", err)
	}

	if c.currentMode != "b" {
		t.Errorf("wanted b, got %s", c.currentMode)
	}
}

func TestNextModeEmptyWanted(t *testing.T) {
	c := &config{
		currentMode: "",
		order: []string{
			"a", "b",
		},
	}

	if err := c.next(); err != nil {
		t.Fatalf("couldn't next(): %s", err)
	}

	if c.currentMode != "a" {
		t.Errorf("wanted a, got %s", c.currentMode)
	}
}
