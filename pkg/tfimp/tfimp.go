package tfimp

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

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
	tfExec   *tfexec.Terraform
	state    *tfjson.State
	noDryRun bool
	cache    map[string][]*tfjson.StateResource
}

func TfImporter(workingDir string, noDryRun bool) (*TfImport, error) {
	tf, err := clientInit(workingDir)
	if err != nil {
		return nil, err
	}

	// get tf state
	state, err := tf.Show(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error showing state: %v", err)
	}

	return &TfImport{
		tfExec:   tf,
		state:    state,
		noDryRun: noDryRun,
		cache:    make(map[string][]*tfjson.StateResource),
	}, nil
}

func (t *TfImport) GetResources(resType string) []*tfjson.StateResource {
	if _, ok := t.cache[resType]; !ok {
		t.getAllResourcesForType(resType, t.state.Values.RootModule)
	}

	return t.cache[resType]
}

func (t *TfImport) getAllResourcesForType(resType string, root *tfjson.StateModule) {
	if len(root.Resources) > 0 {
		for _, r := range root.Resources {
			if r.Mode == "managed" && r.Type == resType {
				t.cache[resType] = append(t.cache[resType], r)
			}
		}
	}
	if len(root.ChildModules) > 0 {
		for _, c := range root.ChildModules {
			t.getAllResourcesForType(resType, c)
		}
	}
	return
}

func (t *TfImport) Import(name string, value string) error {
	if !t.noDryRun {
		fmt.Printf("[DryRun] Executing: terraform import '%v' '%v'\n", name, value)
	} else {
		if err := t.tfExec.Import(context.TODO(), name, value); err != nil {
			return err
		}
	}
	return nil
}

func SetImportAddrFromResource(newResource string, sourceResource *tfjson.StateResource) (string, error) {
	if newResource == "" {
		return "", fmt.Errorf("Import name: %s not supported", newResource)
	}

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
