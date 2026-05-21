package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pterm/pterm"
)

func TestDoctorCommand(t *testing.T) {
	// Disable global colors for raw string matching
	pterm.DisableColor()
	defer pterm.EnableColor() // cleanup after test

	// Capture all output to a buffer
	var buf bytes.Buffer
	pterm.SetDefaultOutput(&buf)
	// Make sure we restore output after test
	defer pterm.SetDefaultOutput(nil)

	// Provide the buffer to Cobra explicitly
	DoctorCmd.SetOut(&buf)
	DoctorCmd.SetErr(&buf)

	// Execute without real config, should fail
	err := DoctorCmd.Execute()

	if err == nil {
		t.Fatalf("Expected doctor to fail lacking config file")
	}

	output := buf.String()

	// We check the returned error since rootCmd error handling is in main.go
	if !strings.Contains(err.Error(), "no server configured") && !strings.Contains(err.Error(), "not in an OTAShip project") && !strings.Contains(err.Error(), "cannot find") && !strings.Contains(err.Error(), "no such file") && !strings.Contains(err.Error(), "The system cannot find the file specified") && !strings.Contains(err.Error(), "system cannot find") && !strings.Contains(err.Error(), "file does not exist") {
		t.Errorf("Expected error complaining about missing file or config, got: %q", err.Error())
	}

	// The buffer should not contain ERROR since the ERROR prefix is printed by main.go
	// However, if there was intermediate output (e.g. "Config loaded"), we would see it here.
	if strings.Contains(output, "ERROR") {
		t.Errorf("Did not expect ERROR tag in command buffer (handled in main.go), got: %q", output)
	}
}
