// Package grafana contains option and resource builder for grafana dashboards
package grafana

import (
	"github.com/telcoltar/options/pkg/option"

	corev1 "k8s.io/api/core/v1"
)

// ImageValues represents the image configuration options for Grafana
type ImageValues struct {
	Registry    *option.Base[string]
	Repository  *option.Base[string]
	Tag         *option.Base[string]
	PullPolicy  *option.Base[corev1.PullPolicy]
	PullSecrets *option.Slice[string]

	*option.Container
}

// NewImageValues creates a new ImageValues instance with default values
func NewImageValues() *ImageValues {
	iv := &ImageValues{
		Registry:    option.NewBase[string]("registry").Default("docker.io"),
		Repository:  option.NewBase[string]("repository").Default("grafana/grafana"),
		Tag:         option.NewBase[string]("tag").Default("latest"),
		PullPolicy:  option.NewBase[corev1.PullPolicy]("pullPolicy").Default(corev1.PullIfNotPresent),
		PullSecrets: option.NewSlice[string]("pullSecrets").EmptyDefault(),
	}

	iv.Container = option.NewContainer("image", iv)
	return iv
}

// ServiceValues represents the service configuration options for Grafana
type ServiceValues struct {
	Enabled                  *option.Base[bool]
	Type                     *option.Base[corev1.ServiceType]
	IPFamilyPolicy           *option.Base[corev1.IPFamilyPolicy]
	IPFamilies               *option.Slice[corev1.IPFamily]
	LoadBalancerIP           *option.Base[string]
	LoadBalancerClass        *option.Base[string]
	LoadBalancerSourceRanges *option.Slice[string]
	Port                     *option.Base[int32]
	TargetPort               *option.Base[int32]
	NodePort                 *option.Base[int32]
	Annotations              *option.Map[string]
	Labels                   *option.Map[string]
	PortName                 *option.Base[string]
	AppProtocol              *option.Base[string]
	SessionAffinity          *option.Base[string]

	*option.Container
}

// NewServiceValues creates a new ServiceValues instance with default values
func NewServiceValues() *ServiceValues {
	sv := &ServiceValues{
		Enabled: option.NewBase[bool]("enabled").Default(true),
		Type:    option.NewBase[corev1.ServiceType]("type").Default(corev1.ServiceTypeClusterIP),
		IPFamilyPolicy: option.NewBase[corev1.IPFamilyPolicy]("ipFamilyPolicy").Enum(
			corev1.IPFamilyPolicyPreferDualStack,
			corev1.IPFamilyPolicyRequireDualStack,
			corev1.IPFamilyPolicySingleStack,
		).EmptyDefault(),
		IPFamilies: option.NewSlice[corev1.IPFamily]("ipFamilies").Enum(
			corev1.IPv4Protocol, corev1.IPv6Protocol, corev1.IPFamilyUnknown,
		).EmptyDefault(),
		LoadBalancerIP:           option.NewBase[string]("loadBalancerIP").EmptyDefault(),
		LoadBalancerClass:        option.NewBase[string]("loadBalancerClass").EmptyDefault(),
		LoadBalancerSourceRanges: option.NewSlice[string]("loadBalancerSourceRanges").EmptyDefault(),
		Port:                     option.NewBase[int32]("port").Default(80),
		TargetPort:               option.NewBase[int32]("targetPort").Default(3000),
		NodePort:                 option.NewBase[int32]("nodePort"),
		Annotations:              option.NewMap[string]("annotations").EmptyDefault(),
		Labels:                   option.NewMap[string]("labels").EmptyDefault(),
		PortName:                 option.NewBase[string]("portName").Default("service"),
		AppProtocol:              option.NewBase[string]("appProtocol").EmptyDefault(),
		SessionAffinity: option.NewBase[string]("sessionAffinity").
			Enum("ClientIP", "None").EmptyDefault(),
	}

	sv.Container = option.NewContainer("service", sv)
	return sv
}

// GrafanaHelmValues represents the complete Grafana helm chart values
type GrafanaHelmValues struct {
	LivenessProbe  *ProbeValues
	ReadinessProbe *ProbeValues

	SecurityContext              *option.Base[string]
	Replicas                     *option.Base[int32]
	AutomountServiceAccountToken *option.Base[bool]
	EnvFromSecret                *option.Base[string]
	EnvFromSecrets               *option.Slice[string]
	ExtraLabels                  *option.Map[string]
	Env                          *option.Map[string]
	PodLabels                    *option.Map[string]

	Image   *ImageValues
	Service *ServiceValues

	*option.Container
}

// NewGrafanaHelmValues creates a new GrafanaHelmValues instance with default values
func NewGrafanaHelmValues() *GrafanaHelmValues {
	ghv := GrafanaHelmValues{
		LivenessProbe:  NewLivenessProbeValues(),
		ReadinessProbe: NewReadinessProbeValues(),
		SecurityContext: option.NewBase[string]("securityContext").Default(
			`{"runAsUser": 472, "runAsGroup": 472, "fsGroup": 472}`,
		),
		Replicas:                     option.NewBase[int32]("replicas").Default(1),
		AutomountServiceAccountToken: option.NewBase[bool]("automountServiceAccountToken").Default(true),
		EnvFromSecret:                option.NewBase[string]("envFromSecret").EmptyDefault(),
		EnvFromSecrets:               option.NewSlice[string]("envFromSecrets").EmptyDefault(),
		ExtraLabels:                  option.NewMap[string]("extraLabels").EmptyDefault(),
		Env:                          option.NewMap[string]("env").EmptyDefault(),
		PodLabels:                    option.NewMap[string]("podLabels"),

		Image:   NewImageValues(),
		Service: NewServiceValues(),
	}
	ghv.Container = option.NewContainer("GrafanaHelmValue", ghv)
	return &ghv
}
