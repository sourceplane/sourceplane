package cmd

import (
	"fmt"

	"github.com/sourceplane/sourceplane/internal/providers"
	"github.com/spf13/cobra"
)

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Manage Sourceplane providers",
	Long:  `Commands for managing provider installations, cache, and versions.`,
}

var providersInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Download and cache providers from intent.yaml",
	Long: `Downloads all providers specified in intent.yaml and caches them locally.
Similar to 'terraform init', this ensures all required providers are available.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		intentFile, _ := cmd.Flags().GetString("intent")

		fmt.Println("Initializing providers...")
		if err := providers.InitProviders(intentFile); err != nil {
			return err
		}

		fmt.Println("\nProviders initialized successfully!")
		return nil
	},
}

var providersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cached providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cache, err := providers.NewProviderCache()
		if err != nil {
			return err
		}

		cached, err := cache.ListCachedProviders()
		if err != nil {
			return err
		}

		if len(cached) == 0 {
			fmt.Println("No cached providers found. Run 'sp providers init' to download providers.")
			return nil
		}

		fmt.Println("Cached providers:")
		for _, p := range cached {
			fmt.Printf("  %s/%s@%s\n", p.Owner, p.Repo, p.Version)
			fmt.Printf("    Path: %s\n", p.Path)
		}

		return nil
	},
}

var providersClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the provider cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		cache, err := providers.NewProviderCache()
		if err != nil {
			return err
		}

		if err := cache.ClearCache(); err != nil {
			return err
		}

		fmt.Println("Provider cache cleared successfully!")
		return nil
	},
}

func init() {
	providersCmd.AddCommand(providersInitCmd)
	providersCmd.AddCommand(providersListCmd)
	providersCmd.AddCommand(providersClearCmd)

	providersInitCmd.Flags().String("intent", "intent.yaml", "Path to intent.yaml file")

	rootCmd.AddCommand(providersCmd)
}
