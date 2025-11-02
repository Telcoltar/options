// Package utils has helper function
package utils

import (
	"bytes"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func MarshalResources(objects []runtime.Object, format string) ([]byte, error) {
	scheme := runtime.NewScheme()

	serializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme,
		scheme,
		json.SerializerOptions{
			Yaml:   format == "yaml",
			Pretty: format == "json",
		},
	)

	var buf bytes.Buffer

	for i, obj := range objects {
		if i > 0 && format == "yaml" {
			buf.WriteString("---\n")
		}

		if err := serializer.Encode(obj, &buf); err != nil {
			return nil, err
		}

		if i < len(objects)-1 && format != "yaml" {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes(), nil
}
