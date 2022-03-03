package bootstrap

import (
	"fmt"
	"xenotification/app/env"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

// InitJaeger :
func (bs *Bootstrap) initJaeger() *Bootstrap {
	def := config.Configuration{
		ServiceName: "xenotification",
		Disabled:    false,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: false,
		},
	}

	jLogger := log.StdLogger
	jMetricsFactory := metrics.NullFactory

	cfg, err := def.FromEnv()
	if err != nil {
		panic("Could not parse Jaeger env vars: " + err.Error())
	}

	if env.IsDevelopment() {
		opentracing.SetGlobalTracer(new(opentracing.NoopTracer))
	} else {
		tracer, _, err := cfg.NewTracer(
			config.Logger(jLogger),
			config.Metrics(jMetricsFactory),
			config.MaxTagValueLength(2048),
		)
		if err != nil {
			panic(fmt.Sprintf("cannot init Jaeger: %v\n", err))
		}
		opentracing.SetGlobalTracer(tracer)
	}

	return bs
}
