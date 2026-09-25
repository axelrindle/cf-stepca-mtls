package config

import (
	"charm.land/log/v2"
)

func (c Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c Config) CharmLoggerLevel() log.Level {
	var level log.Level

	switch c.Logging.Level {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	}

	return level
}
