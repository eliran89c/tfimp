package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"
	tfjson "github.com/hashicorp/terraform-json"
)

type Config struct {
	Version string `json:"version"`
	Steps   []step `json:"steps,omitempty"`
}

type step struct {
	ImportName     string         `json:"import_name"`
	ForEach        forEachBlock   `json:"for_each,omitempty"`
	Condition      condition      `json:"condition,omitempty"`
	ValueTransform ValueTransform `json:"transform,omitempty"`
}

type forEachBlock struct {
	Resource  string `json:"resource"`
	Attribute string `json:"attribute"`
}

type condition struct {
	ConditionKey string `json:"key"`
}

func (c *condition) Check(r *tfjson.StateResource) bool {
	if c.ConditionKey == "" {
		// no condition specified
		return true
	}
	resAttrByte, _ := json.Marshal(r.AttributeValues)
	keys := strings.Split(c.ConditionKey, ".")
	v, t, _, err := jsonparser.Get(resAttrByte, keys...)

	if err != nil {
		fmt.Printf("Missing condition key: %v, for resource: %v\n", c.ConditionKey, r.Address)
		return false
	}

	switch t.String() {
	case "string":
		{
			return string(v) != ""
		}
	case "boolean":
		{
			return string(v) == "true"
		}
	case "array":
		{
			var s []interface{}
			json.Unmarshal(v, &s)
			return len(s) > 0
		}
	case "null":
		{
			return false
		}
	default:
		{
			fmt.Printf("Condition key must be of type: list, bool, string. got: %v\n", t)
			return false
		}
	}
}

func (fe *forEachBlock) IsEmpty() bool {
	return (fe.Attribute == "" || fe.Resource == "")
}

func NewConfigFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	byteJson, _ := ioutil.ReadAll(f)
	json.Unmarshal(byteJson, &config)
	return &config, nil
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
