package main

import (
	"os"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/commands"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:           "otaship",
	Short:         "OTAShip CLI - Manage OTA updates for Expo apps",
	SilenceErrors: true, // Prevents Cobra from printing the error again
	SilenceUsage:  true, // Prevents Cobra from printing help menu on execution errors
}

func main() {
	rootCmd.AddCommand(commands.VersionCmd)
	rootCmd.AddCommand(commands.InstallCmd)
	rootCmd.AddCommand(commands.UpgradeCmd)
	rootCmd.AddCommand(commands.LoginCmd)
	rootCmd.AddCommand(commands.InitCommand)
	rootCmd.AddCommand(commands.LinkCommand)
	rootCmd.AddCommand(commands.StatusCommand)
	rootCmd.AddCommand(commands.PublishCommand)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.DeleteCmd)
	rootCmd.AddCommand(commands.RollbackCmd)
	rootCmd.AddCommand(commands.ResetCmd)
	rootCmd.AddCommand(commands.DoctorCmd)
	rootCmd.AddCommand(commands.WhoAmICmd)
	if err := rootCmd.Execute(); err != nil {
		errMsg := err.Error()
		if len(errMsg) > 0 {
			runes := []rune(errMsg)
			runes[0] = unicode.ToUpper(runes[0])
			errMsg = string(runes)
		}
		ui.Error.Println(errMsg)
		os.Exit(1)
	}
}
