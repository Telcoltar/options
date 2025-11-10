package grafana

import (
	"strings"

	"github.com/telcoltar/options/pkg/option"
)

type PersistenceValues struct {
	Enabled          *option.Bool
	Type             *option.String
	Size             *option.String
	StorageClassName *option.String
	ExtraPvcLabels   *option.Map[string]
	ExistingClaim    *option.String
	SubPath          *option.String

	*option.Container[PersistenceValues]
}

func NewPersistenceValues() *PersistenceValues {
	pv := &PersistenceValues{
		Enabled: option.NewBool("enabled").Default(false),
		Type: option.NewString("type").
			Enum("pvc", "statefulset", "sts", "StatefulSet", "PVC").
			Default("pvc").
			Transform(func(value string) string {
				return strings.ToLower(value)
			}),
		Size:             option.NewString("size").Regex("^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$"),
		StorageClassName: option.NewString("storageClassName"),
		ExtraPvcLabels:   option.NewMap[string]("extraPvcLabels"),
		ExistingClaim:    option.NewString("existingClaim"),
		SubPath:          option.NewString("subPath"),
	}

	pv.Container = option.NewContainer("persistence", pv)

	pv.Check(func(pv *PersistenceValues) bool {
		if pv.Enabled.Get() {
			if !pv.Size.HasValue() {
				return false
			}
		}
		return true
	},
		map[string]any{
			"if": map[string]any{
				"properties": map[string]any{
					"enabled": map[string]any{
						"const": true,
					},
				},
			},
			"then": map[string]any{
				"required": []string{"size"},
			},
		},
	)

	return pv
}
