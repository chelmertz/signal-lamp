package main

import (
	"fmt"
	"testing"
)

func TestGnomeTerminalProfilesFromDconfDump(t *testing.T) {
	// what I ran: dconf dump /org/gnome/terminal/legacy/profiles:/
	dconfDumpOutput := `[/]
default='b0682302-667e-4ceb-b714-c05924ab92fc'
list=['b1dcc9dd-5262-4d8d-a863-c897e6d979b9', 'b0682302-667e-4ceb-b714-c05924ab92fc']

[:b0682302-667e-4ceb-b714-c05924ab92fc]
audible-bell=false
background-color='rgb(255,255,255)'
foreground-color='rgb(23,20,33)'
use-theme-colors=false
visible-name='light'

[:b1dcc9dd-5262-4d8d-a863-c897e6d979b9]
audible-bell=false
background-color='rgb(0,0,0)'
default-size-columns=100
font='Source Code Pro 14'
foreground-color='rgb(0,255,0)'
palette=['rgb(23,20,33)', 'rgb(192,28,40)', 'rgb(38,162,105)', 'rgb(162,115,76)', 'rgb(18,72,139)', 'rgb(163,71,186)', 'rgb(42,161,179)', 'rgb(208,207,204)', 'rgb(94,92,100)', 'rgb(246,97,81)', 'rgb(51,209,122)', 'rgb(233,173,12)', 'rgb(42,123,222)', 'rgb(192,97,203)', 'rgb(51,199,222)', 'rgb(255,255,255)']
use-system-font=false
use-theme-colors=false
use-theme-transparency=true
visible-name='dark'
`
	current, profiles, err := dconfDumpProfiles(dconfDumpOutput)

	if err != nil {
		t.Fatalf("wanted nil as error, got %v", err)
	}

	if current != "b0682302-667e-4ceb-b714-c05924ab92fc" {
		t.Errorf("wrong current, got %s", current)
	}

	expectedProfiles := map[string]string{
		"light": "b0682302-667e-4ceb-b714-c05924ab92fc",
		"dark":  "b1dcc9dd-5262-4d8d-a863-c897e6d979b9",
	}

	if len(profiles) != 2 {
		t.Errorf("wrong profile len, got %v", profiles)
	}

	fmt.Printf("got map: %v\n", profiles)

	for k, v := range expectedProfiles {
		if actual, ok := profiles[k]; !ok || actual != v {
			t.Errorf("mismatch between actual:\n%v\n\nand expected:\n%v", profiles, expectedProfiles)
		}
	}
}
