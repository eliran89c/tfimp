package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/buger/jsonparser"
	tfjson "github.com/hashicorp/terraform-json"
)

type Config struct {
	Version string `json:"version"`
	Steps   []step `json:"steps,omitempty"`
}

type step struct {
	ImportName string       `json:"import_name"`
	ForEach    forEachBlock `json:"for_each,omitempty"`
	Condition  condition    `json:"condition,omitempty"`
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
