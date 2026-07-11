package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "fluxend",
	Short: "Fluxend CLI for managing the BaaS platform",
	Long:  `Fluxend CLI allows you to start the server, run seeders, and inspect routes.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func init() {
	RootCmd.AddCommand(aboutCmd)
	RootCmd.AddCommand(serverCmd)
	RootCmd.AddCommand(seedCmd)
	RootCmd.AddCommand(routesCmd)
	RootCmd.AddCommand(udbStats)
	RootCmd.AddCommand(optimizeCmd)
}
