// Package grafana contains option and resource builder for grafana dashboards
package grafana

import (
	"github.com/telcoltar/options/pkg/option"

	corev1 "k8s.io/api/core/v1"
)

// ImageValues represents the image configuration options for Grafana
type ImageValues struct {
	Registry    *option.Simple[string]
	Repository  *option.Simple[string]
	Tag         *option.Simple[string]
	PullPolicy  *option.Simple[corev1.PullPolicy]
	PullSecrets *option.Slice[string]

	*option.Container[ImageValues]
}

// NewImageValues creates a new ImageValues instance with default values
func NewImageValues() *ImageValues {
	iv := &ImageValues{
		Registry:    option.NewSimple[string]("registry").Default("docker.io"),
		Repository:  option.NewSimple[string]("repository").Default("grafana/grafana"),
		Tag:         option.NewSimple[string]("tag").Default("latest"),
		PullPolicy:  option.NewSimple[corev1.PullPolicy]("pullPolicy").Default(corev1.PullIfNotPresent),
		PullSecrets: option.NewSlice[string]("pullSecrets").EmptyDefault(),
	}

	iv.Container = option.NewContainer("image", iv)
	return iv
}

// ServiceValues represents the service configuration options for Grafana
type ServiceValues struct {
	Enabled                  *option.Simple[bool]
	Type                     *option.Simple[corev1.ServiceType]
	IPFamilyPolicy           *option.Simple[corev1.IPFamilyPolicy]
	IPFamilies               *option.Slice[corev1.IPFamily]
	LoadBalancerIP           *option.Simple[string]
	LoadBalancerClass        *option.Simple[string]
	LoadBalancerSourceRanges *option.Slice[string]
	Port                     *option.Simple[int32]
	TargetPort               *option.Simple[int32]
	NodePort                 *option.Simple[int32]
	Annotations              *option.Map[string]
	Labels                   *option.Map[string]
	PortName                 *option.Simple[string]
	AppProtocol              *option.Simple[string]
	SessionAffinity          *option.Simple[string]

	*option.Container[ServiceValues]
}

// NewServiceValues creates a new ServiceValues instance with default values
func NewServiceValues() *ServiceValues {
	sv := &ServiceValues{
		Enabled: option.NewSimple[bool]("enabled").Default(true),
		Type:    option.NewSimple[corev1.ServiceType]("type").Default(corev1.ServiceTypeClusterIP),
		IPFamilyPolicy: option.NewSimple[corev1.IPFamilyPolicy]("ipFamilyPolicy").Enum(
			corev1.IPFamilyPolicyPreferDualStack,
			corev1.IPFamilyPolicyRequireDualStack,
			corev1.IPFamilyPolicySingleStack,
		).EmptyDefault(),
		IPFamilies: option.NewSlice[corev1.IPFamily]("ipFamilies").ItemOption.Enum(
			corev1.IPv4Protocol, corev1.IPv6Protocol, corev1.IPFamilyUnknown,
		).EmptyDefault(),
		LoadBalancerIP:           option.NewSimple[string]("loadBalancerIP").EmptyDefault(),
		LoadBalancerClass:        option.NewSimple[string]("loadBalancerClass").EmptyDefault(),
		LoadBalancerSourceRanges: option.NewSlice[string]("loadBalancerSourceRanges").EmptyDefault(),
		Port:                     option.NewSimple[int32]("port").Default(80),
		TargetPort:               option.NewSimple[int32]("targetPort").Default(3000),
		NodePort:                 option.NewSimple[int32]("nodePort"),
		Annotations:              option.NewMap[string]("annotations").EmptyDefault(),
		Labels:                   option.NewMap[string]("labels").EmptyDefault(),
		PortName:                 option.NewSimple[string]("portName").Default("service"),
		AppProtocol:              option.NewSimple[string]("appProtocol").EmptyDefault(),
		SessionAffinity: option.NewSimple[string]("sessionAffinity").
			Enum("ClientIP", "None").EmptyDefault(),
	}

	sv.Container = option.NewContainer("service", sv)
	return sv
}

// GrafanaHelmValues represents the complete Grafana helm chart values
type GrafanaHelmValues struct {
	LivenessProbe  *ProbeValues
	ReadinessProbe *ProbeValues

	SecurityContext              *option.Simple[string]
	Replicas                     *option.Simple[int32]
	AutomountServiceAccountToken *option.Simple[bool]
	EnvFromSecret                *option.Simple[string]
	EnvFromSecrets               *option.Slice[string]
	ExtraLabels                  *option.Map[string]
	Env                          *option.Map[string]
	PodLabels                    *option.Map[string]

	Image       *ImageValues
	Service     *ServiceValues
	Persistence *PersistenceValues

	*option.Container[GrafanaHelmValues]
}

// NewGrafanaHelmValues creates a new GrafanaHelmValues instance with default values
func NewGrafanaHelmValues() *GrafanaHelmValues {
	ghv := GrafanaHelmValues{
		LivenessProbe:  NewLivenessProbeValues(),
		ReadinessProbe: NewReadinessProbeValues(),
		SecurityContext: option.NewSimple[string]("securityContext").Default(
			`{"runAsUser": 472, "runAsGroup": 472, "fsGroup": 472}`,
		),
		Replicas:                     option.NewSimple[int32]("replicas").Default(1),
		AutomountServiceAccountToken: option.NewSimple[bool]("automountServiceAccountToken").Default(true),
		EnvFromSecret:                option.NewSimple[string]("envFromSecret").EmptyDefault(),
		EnvFromSecrets:               option.NewSlice[string]("envFromSecrets").EmptyDefault(),
		ExtraLabels:                  option.NewMap[string]("extraLabels").EmptyDefault(),
		Env:                          option.NewMap[string]("env").EmptyDefault(),
		PodLabels:                    option.NewMap[string]("podLabels"),

		Image:       NewImageValues(),
		Service:     NewServiceValues(),
		Persistence: NewPersistenceValues(),
	}
	ghv.Container = option.NewContainer("GrafanaHelmValue", &ghv)
	return &ghv
}
