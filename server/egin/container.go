package egin

import (
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/eflag"
	"github.com/gotomicro/ego/core/elog"
)

type Option func(c *Container)

type Container struct {
	config *Config
	name   string
	logger *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	// 修改host信息
	if eflag.String("host") != "" {
		c.config.Host = eflag.String("host")
	}
	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	server := newComponent(c.name, c.config, c.logger)
	server.Use(recoverMiddleware(c.logger, c.config.SlowLogThreshold))

	if c.config.EnableMetricInterceptor {
		server.Use(metricServerInterceptor())
	}

	if c.config.EnableTraceInterceptor {
		server.Use(traceServerInterceptor())
	}
	return server
}
