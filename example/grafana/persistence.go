package grafana

import (
	"strings"

	"github.com/telcoltar/options/pkg/option"
)

type PersistenceValues struct {
	Enabled *option.Bool
	Type    *option.String
	Size    *option.String
}

func NewPersistenceValues() *PersistenceValues {
	return &PersistenceValues{
		Enabled: option.NewBool("enabled").Default(false),
		Type: option.NewString("type").
			Enum("pvc", "statefulset", "sts", "StatefulSet", "PVC").
			Default("pvc").
			Transform(func(value string) string {
				return strings.ToLower(value)
			}),
		Size: option.NewString("size"),
	}
}
