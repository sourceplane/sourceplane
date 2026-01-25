package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sourceplane/sourceplane/internal/thinci"
	"gopkg.in/yaml.v3"
)

func main() {
	data, err := os.ReadFile("providers/helm/provider.yaml")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		fmt.Printf("Error unmarshaling YAML: %v\n", err)
		return
	}

	thinCIConfig, ok := raw["thinCI"].(map[string]interface{})
	if !ok {
		fmt.Println("No thinCI config found")
		return
	}

	actionsRaw, ok := thinCIConfig["actions"].([]interface{})
	if !ok {
		fmt.Println("No actions found")
		return
	}

	for i, actionRaw := range actionsRaw {
		actionData, _ := yaml.Marshal(actionRaw)

		var action thinci.ProviderAction
		if err := yaml.Unmarshal(actionData, &action); err != nil {
			fmt.Printf("Error unmarshaling action: %v\n", err)
			continue
		}

		fmt.Printf("\n=== Action %d: %s ===\n", i, action.Name)
		fmt.Printf("Order: %d\n", action.Order)
		fmt.Printf("JobTemplate is nil: %v\n", action.JobTemplate == nil)
		fmt.Printf("Commands: %+v\n", action.Commands)

		if action.JobTemplate != nil {
			jsonBytes, _ := json.MarshalIndent(action.JobTemplate, "", "  ")
			fmt.Printf("JobTemplate JSON:\n%s\n", string(jsonBytes))
		} else {
			fmt.Println("JobTemplate is nil!")
		}
	}
}
