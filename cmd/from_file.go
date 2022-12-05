package cmd

import (
	"fmt"

	"github.com/eliran89c/tfimp/pkg/config"
	"github.com/eliran89c/tfimp/pkg/tfimp"
	"github.com/spf13/cobra"
)

var (
	configPath        string
	fromConfigFileCmd = &cobra.Command{
		Use:   "from-file",
		Short: "Import new resources based on configuration file",
		Long:  ``,
		RunE:  fromConfigFile,
	}
)

func init() {
	fromConfigFileCmd.Flags().StringVarP(&configPath, "config", "f", "", "The location of the configuration file.")
	fromConfigFileCmd.MarkFlagRequired("config")
}

func fromConfigFile(cmd *cobra.Command, _ []string) error {
	tfImport, err := tfimp.TfImporter(workingDir, noDryRun)
	if err != nil {
		return err
	}
	config, err := config.NewConfigFromFile(configPath)
	if err != nil {
		return err
	}
	for _, step := range config.Steps {
		if step.ForEach.IsEmpty() {
			fmt.Printf("[WARN] for_each block is missing for %v import\n", step.ImportName)
			continue
		}

		for _, r := range tfImport.GetResources(step.ForEach.Resource) {
			attrVal, ok := r.AttributeValues[step.ForEach.Attribute]
			if !ok {
				fmt.Printf("[WARN] Missing attribute `%v` for resource %v\n", step.ForEach.Attribute, r.Address)
				continue
			}

			//match found
			importAddr, err := tfimp.SetImportAddrFromResource(step.ImportName, r)
			if err != nil {
				return err
			}

			// check conditions
			if pass := step.Condition.Check(r); !pass {
				fmt.Printf("skipping resource: `%v` due to condition check\n", r.Address)
				continue
			}

			if err = tfImport.Import(importAddr, attrVal.(string)); err != nil {
				return err
			}
		}
	}

	return nil
}
