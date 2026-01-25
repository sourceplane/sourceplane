package cmd

import (
	"github.com/spf13/cobra"
)

// version is set at build time via ldflags
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "sp",
	Short: "Sourceplane CLI - Component-driven tool for software organizations",
	Long: `Sourceplane CLI is a component-driven tool for defining, understanding, 
and managing software repositories and organizations.

It codifies intent as code and enables organizational introspection and architectural analysis.`,
	Version: version,
}

var thinCIRootCmd = &cobra.Command{
	Use:   "thinci",
	Short: "Thin-CI - Deterministic CI/CD planning engine",
	Long: `Thin-CI generates deterministic execution plans for CI systems.
It does not execute CI, it only creates plans that can be rendered into workflows.`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteThinCI runs the thin-ci standalone command
func ExecuteThinCI() error {
	return thinCIRootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	thinCIRootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add thin-ci as subcommand to main CLI
	rootCmd.AddCommand(thinCICmd)

	// Add plan and run commands to standalone thin-ci CLI
	thinCIRootCmd.AddCommand(thinCIPlanCmd)
	thinCIRootCmd.AddCommand(thinCIRunCmd)
}
