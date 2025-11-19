package grafana

import (
	"maps"

	"github.com/telcoltar/options/pkg/option"

	"github.com/Telcoltar/kubernetes-resources/builders"
	"github.com/Telcoltar/kubernetes-resources/utils"
)

func Build(name string, values *GrafanaHelmValues) (builders.Builders, error) {
	resources := builders.Builders{}
	if values.Service.Enabled.Get() {
		resources = append(resources, BuildService(name, values.Service))
	}

	deploy, err := BuildDeployment(name, values)
	if err != nil {
		return nil, err
	}
	resources = append(resources, deploy)

	selectorLabels := map[string]string{
		utils.LabelInstance: name,
		utils.LabelName:     "grafana",
	}
	deploy.Selector(selectorLabels)

	if values.Persistence.Enabled.Get() && !values.Persistence.ExistingClaim.NotZero() {
		resources = append(resources, AddPVC(name, deploy, values.Persistence))
	}

	labels := make(map[string]string)
	if values.Image.Tag.HasValue() {
		labels[utils.LabelVersion] = values.Image.Tag.Get()
	}
	if values.ExtraLabels.NotZero() {
		maps.Copy(labels, values.ExtraLabels.Get())
	}
	maps.Copy(labels, selectorLabels)

	resources.Labels(labels)

	return resources, nil
}

func AddPVC(name string, deploy *builders.DeploymentBuilder, pvcValues *PersistenceValues) *builders.PersistentVolumeClaimBuilder {
	pvc := builders.PVC(name).Size(pvcValues.Size.Get()).
		StorageClassP(pvcValues.StorageClassName.GetPointer()).
		Labels(pvcValues.ExtraPvcLabels.Get())
	deploy.Volumes(builders.Volume("storage").PVC(pvc.GetName()))
	return pvc
}

func BuildDeployment(name string, values *GrafanaHelmValues) (*builders.DeploymentBuilder, error) {
	envs := []builders.EnvBuilderI{}
	for key, value := range values.Env.Get() {
		envs = append(envs, builders.Env(key).Value(value))
	}

	mount := builders.VolumeMount("storage").Path("/var/lib/grafana")
	if values.Persistence.SubPath.HasValue() {
		mount.SubPath(values.Persistence.SubPath.Get())
	}

	container := builders.Container(name).
		Image(
			builders.Image(values.Image.Repository.Get()).
				Registry(values.Image.Registry.Get()).
				Tag(values.Image.Tag.Get()),
		).
		Envs(envs...).
		Ports(builders.ContainerPort().Name("http").Port(values.Service.TargetPort.Get())).
		ReadinessProbe(
			builders.Probe().
				InitialDelay(values.ReadinessProbe.InitialDelay.Get()).
				Timeout(values.ReadinessProbe.Timeout.Get()).
				FailureThreshold(values.ReadinessProbe.FailureThreshold.Get()).
				HTTP().Path("/api/health").Port(values.Service.TargetPort.Get()),
		).
		LivenessProbe(
			builders.Probe().
				InitialDelay(values.LivenessProbe.InitialDelay.Get()).
				Timeout(values.LivenessProbe.Timeout.Get()).
				FailureThreshold(values.LivenessProbe.FailureThreshold.Get()).
				HTTP().Path("/api/health").Port(values.Service.TargetPort.Get()),
		).
		VolumeMounts(mount)

	sc, createSCError := builders.PodSecurityContextFromYAML(values.SecurityContext.Get())
	if createSCError != nil {
		return nil, createSCError
	}

	deployment := builders.Deployment(name).
		Replicas(values.Replicas.Get()).
		SecurityContext(sc).
		Containers(container).
		PodLabels(values.PodLabels.Get())
	return deployment, nil
}

func BuildService(name string, values *ServiceValues) builders.Builder {
	port := builders.ServicePort().Name(values.PortName.Get()).
		Port(values.Port.Get()).TargetPort(values.TargetPort.Get())

	option.SetNotNil(values.NodePort, port.NodePort)

	service := builders.Service(name).
		Type(values.Type.Get()).
		LoadbalancerIP(values.LoadBalancerIP.Get()).
		Ports(port)
	return service
}
