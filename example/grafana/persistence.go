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

	*option.Container
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
		Size:             option.NewString("size"),
		StorageClassName: option.NewString("storageClassName"),
		ExtraPvcLabels:   option.NewMap[string]("extraPvcLabels"),
		ExistingClaim:    option.NewString("existingClaim"),
		SubPath:          option.NewString("subPath"),
	}

	pv.Container = option.NewContainer("persistence", pv)
	return pv
}
