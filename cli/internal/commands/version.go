package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show CLI version",
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("otaship version %s\n", Version)
	fmt.Printf("built: %s\n", BuildDate)
}
