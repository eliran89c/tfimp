package config

import (
	"fmt"
)

type ValueTransform struct {
	Action string      `json:"action"`
	Value  interface{} `json:"value"`
}

func (vt *ValueTransform) Transform(value string) (string, error) {
	// specific implementaion to support bucket suffix imports
	if vt.Action == "" {
		return value, nil
	}

	switch vt.Action {
	case "useSuffix":
		{
			return vt.useSuffix(value)
		}
	default:
		{
			return "", fmt.Errorf("Unsupported transform action: %v, choose one from: [useSuffix]", vt.Action)
		}
	}
}

func (vt *ValueTransform) useSuffix(value string) (string, error) {
	switch vt.Value.(type) {
	case float64:
		{
			suffixSize := int(vt.Value.(float64))
			newVal := value[len(value)-suffixSize:]
			return string(newVal), nil
		}
	default:
		{
			return "", fmt.Errorf("Please enter an integer value for transform of type `useSuffix`")
		}
	}
}
