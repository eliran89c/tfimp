package tfimp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func getAllResources(root *tfjson.StateModule, resources []*tfjson.StateResource) []*tfjson.StateResource {
	if len(root.Resources) > 0 {
		for _, r := range root.Resources {
			if r.Mode == "managed" {
				resources = append(resources, r)
			}
		}
	}
	if len(root.ChildModules) > 0 {
		for _, c := range root.ChildModules {
			return getAllResources(c, resources)
		}
	}
	return resources
}

func clientInit(dir string) (*tfexec.Terraform, error) {
	execPath, err := exec.LookPath("terraform")
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	tf, err := tfexec.NewTerraform(dir, execPath)
	if err != nil {
		return nil, fmt.Errorf("error running Terraform: %v", err)
	}

	// run tf init
	err = tf.Init(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error running Init: %v", err)
	}
	return tf, nil
}

type TfImport struct {
	TfExec    *tfexec.Terraform
	Resources []*tfjson.StateResource
	DryRun    bool
}

func TfImporter(workingDir string, dryRun bool) (*TfImport, error) {
	tf, err := clientInit(workingDir)
	if err != nil {
		return nil, err
	}

	// get all resources
	state, err := tf.Show(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error showing state: %v", err)
	}

	// get all resources from state (root and submodules)
	allResources := getAllResources(state.Values.RootModule, make([]*tfjson.StateResource, 0))

	return &TfImport{
		TfExec:    tf,
		Resources: allResources,
		DryRun:    dryRun,
	}, nil
}

func (t *TfImport) Import(name string, value string) error {
	if t.DryRun {
		fmt.Printf("[DryRun] Executing: terraform import '%v' '%v'\n", name, value)
	} else {
		if err := t.TfExec.Import(context.TODO(), name, value); err != nil {
			return err
		}
	}
	return nil
}

func SetFromResourceImportName(newResource string, sourceResource *tfjson.StateResource) (string, error) {
	re := regexp.MustCompile("\\[[\\w|\"|\\s]+\\]")
	index := strings.Join(re.FindStringSubmatch(newResource), "")
	newResWithoutIndex := re.ReplaceAllString(newResource, "")
	srcResWithoutIndex := re.ReplaceAllString(sourceResource.Address, "")

	// check number of elements in import resource
	switch e := strings.Split(newResWithoutIndex, "."); {
	case len(e) == 1:
		{
			// user provided resource type only, return source resource with new type
			addr := strings.Replace(srcResWithoutIndex, sourceResource.Type, e[0], 1)
			return fmt.Sprintf("%v%v", addr, index), nil
		}
	case len(e) == 2:
		{
			// user provided resource type and name
			addr := strings.Replace(srcResWithoutIndex, sourceResource.Type, e[0], 1)
			addr = strings.Replace(addr, sourceResource.Name, e[1], 1)
			return fmt.Sprintf("%v%v", addr, index), nil
		}
	default:
		{
			return "", fmt.Errorf("Found more than 2 elements for import: %v", newResource)
		}
	}
}

func (t *TfImport) BackupState(backupDir string) error {
	if t.DryRun {
		fmt.Printf("Running in dry-run mode, no backup needed...skipping\n")
	} else {
		filename := uuid.Must(uuid.NewRandom())
		filePath := path.Join(backupDir, filename.String())
		filePath = filePath + ".json"
		state, err := t.TfExec.StatePull(context.TODO())
		if err != nil {
			return err
		}

		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err = f.WriteString(state); err != nil {
			return err
		}

		fmt.Printf("Save state backup file to: %v\n", filePath)
	}
	return nil
}
