package ui

import (
	"os"

	"github.com/pterm/pterm"
)

var (
	Info    = pterm.Info
	Success = pterm.Success
	Warning = pterm.Warning
	Error   = pterm.Error
	Fatal   = pterm.Fatal
)

func ExitError(msg string, err error) {
	Error.Printf("%s: %v\n", msg, err)
	os.Exit(1)
}

func PrintBanner() {
	pterm.DefaultHeader.WithMargin(1).WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).Println("OTAship CLI")
}

func StartSpinner(msg string) (*pterm.SpinnerPrinter, error) {
	return pterm.DefaultSpinner.Start(msg)
}

func StopSpinner(spinner *pterm.SpinnerPrinter) {
	spinner.Stop()
}
