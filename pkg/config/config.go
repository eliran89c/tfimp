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
	Resource  string   `json:"resource"`
	Attribute string   `json:"attribute"`
	Values    []string `json:"values,omitempty"`
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

func (fe *forEachBlock) Contains(resName string) bool {
	if len(fe.Values) == 0 {
		// no restriction, continue
		return true
	}

	for _, v := range fe.Values {
		if v == resName {
			return true
		}
	}
	return false
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

	// strip resources indexes
	re := regexp.MustCompile("\\[[\\w|\"|\\s]+\\]")
	index := strings.Join(re.FindStringSubmatch(newResource), "")
	newResWithoutIndex := re.ReplaceAllString(newResource, "")
	srcResWithoutIndex := re.ReplaceAllString(sourceResource.Address, "")

	// check number of elements in import resource
	newElements := strings.Split(newResWithoutIndex, ".")
	if len(newElements) <= 1 {
		return "", fmt.Errorf("Found less than 2 elements for import: %v. please enter at least resource type and name (e.g. aws_s3_bucket.my_bucket)", newResource)
	}

	// remove the last 2 elements from source resource
	srcElements := strings.Split(srcResWithoutIndex, ".")
	srcElements = srcElements[:len(srcElements)-2]

	// combine the elements (src first, new later)
	elements := append(srcElements, newElements...)
	addr := strings.Join(elements, ".")

	// user provided resource type and name
	return fmt.Sprintf("%v%v", addr, index), nil
}
