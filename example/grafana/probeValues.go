package grafana

import "github.com/telcoltar/options/pkg/option"

type ProbeValues struct {
	InitialDelay     *option.Simple[int32]
	Timeout          *option.Simple[int32]
	FailureThreshold *option.Simple[int32]

	*option.Container
}

func NewLivenessProbeValues() *ProbeValues {
	pv := &ProbeValues{
		InitialDelay:     option.NewSimple[int32]("initialDelay").Default(60),
		Timeout:          option.NewSimple[int32]("timeout").Default(30),
		FailureThreshold: option.NewSimple[int32]("failureThreshold").Default(10),
	}

	pv.Container = option.NewContainer("livenessProbe", pv)
	return pv
}

func NewReadinessProbeValues() *ProbeValues {
	pv := &ProbeValues{
		InitialDelay:     option.NewSimple[int32]("initialDelay"),
		Timeout:          option.NewSimple[int32]("timeout"),
		FailureThreshold: option.NewSimple[int32]("failureThreshold"),
	}

	pv.Container = option.NewContainer("readinessProbe", pv)
	return pv
}
