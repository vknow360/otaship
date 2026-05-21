package ui

import (
	"os"

	"github.com/pterm/pterm"
	"golang.org/x/term"
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
	pterm.DefaultHeader.WithMargin(1).WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).Println("OTAShip CLI")
}

func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func StartSpinner(msg string) (*pterm.SpinnerPrinter, error) {
	return pterm.DefaultSpinner.Start(msg)
}

func StopSpinner(spinner *pterm.SpinnerPrinter) {
	spinner.Stop()
}

func Confirm(msg string) (bool, error) {
	return pterm.DefaultInteractiveConfirm.WithConfirmText(msg).Show()
}

func Select(msg string, options []string, defaultVal string) (string, error) {
	return pterm.DefaultInteractiveSelect.
		WithDefaultText(msg).
		WithOptions(options).
		WithDefaultOption(defaultVal).
		Show()
}

func Ask(msg string) (string, error) {
	return pterm.DefaultInteractiveTextInput.WithDefaultText(msg).Show()
}

func AskSecret(msg string) (string, error) {
	return pterm.DefaultInteractiveTextInput.WithDefaultText(msg).WithMask("*").Show()
}

func AskOptional(msg string, defaultValue string) (string, error) {
	resp, err := pterm.DefaultInteractiveTextInput.WithDefaultText(msg + " [" + defaultValue + "]").Show()
	if err != nil || resp == "" {
		return defaultValue, nil
	}
	return resp, nil
}
