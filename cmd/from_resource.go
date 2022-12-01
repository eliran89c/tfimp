package cmd

import (
	"fmt"

	"github.com/eliran89c/tfimp/tfimp"
	"github.com/spf13/cobra"
)

var (
	resourceType    string
	resourceAttr    string
	fromResourceCmd = &cobra.Command{
		Use:   "from-resource",
		Short: "Import new resources based on existing resource",
		Long:  "Use this feature to import one or more resources for each source resource. Get the import id value from an existing resource (e.g. bucket name)",
		RunE:  fromResource,
	}
)

func init() {
	fromResourceCmd.Flags().StringVarP(&resourceType, "source-resource-type", "t", "", "The source resource type (e.g. aws_s3_bucket).")
	fromResourceCmd.Flags().StringVarP(&resourceAttr, "source-resource-attr", "a", "", "The source resource attribute (e.g. bucket).")
	fromResourceCmd.MarkFlagRequired("source-resource-type")
	fromResourceCmd.MarkFlagRequired("source-resource-attr")
}

func fromResource(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Please enter at least 1 resource type to import")
	}

	tfImport, err := tfimp.TfImporter(workingDir, dryRun)
	if err != nil {
		return err
	}

	// check if need to backup state
	if backup {
		if err = tfImport.BackupState(backupDir); err != nil {
			return err
		}
	}

	// going over all resources that match --source-resource-type
	for _, r := range tfImport.Resources {
		if r.Type != resourceType {
			continue
		}

		attrVal, ok := r.AttributeValues[resourceAttr]
		if !ok {
			fmt.Printf("[WARN] Missing attribute `%v` for resource %v\n", resourceAttr, r.Address)
			continue
		}

		// match found
		for _, newResource := range args {
			newResourceName, err := tfimp.SetFromResourceImportName(newResource, r)
			if err != nil {
				return err
			}
			if err = tfImport.Import(newResourceName, attrVal.(string)); err != nil {
				return err
			}
		}
	}

	return nil
}
