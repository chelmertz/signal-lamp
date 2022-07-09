package main

import "testing"

func TestNextModeCycle(t *testing.T) {
	c := &config{
		currentTheme: "b",
		order: []string{
			"a", "b",
		},
	}
	if err := c.next(); err != nil {
		t.Fatalf("couldn't next(): %s", err)
	}

	if c.currentTheme != "a" {
		t.Errorf("wanted a, got %s", c.currentTheme)
	}

	if err := c.next(); err != nil {
		t.Fatalf("couldn't next(): %s", err)
	}

	if c.currentTheme != "b" {
		t.Errorf("wanted b, got %s", c.currentTheme)
	}
}

func TestNextModeEmptyWanted(t *testing.T) {
	c := &config{
		currentTheme: "",
		order: []string{
			"a", "b",
		},
	}

	if err := c.next(); err != nil {
		t.Fatalf("couldn't next(): %s", err)
	}

	if c.currentTheme != "a" {
		t.Errorf("wanted a, got %s", c.currentTheme)
	}
}
