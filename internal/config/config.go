package config

import "time"

type loggingConfig struct {
	Level string `env:"APP_LOGGING_LEVEL" default:"info" validate:"oneof=debug info warn error"`
}

type apiConfig struct {
	Endpoint            string `env:"STEP_API_ENDPOINT"`
	Provisioner         string `env:"STEP_CERTIFICATE_PROVISIONER" validate:"required"`
	ProvisionerPassword string `env:"STEP_CERTIFICATE_PROVISIONER_PASSWORD" validate:"required"`
}

type certificateConfig struct {
	Subject string `env:"STEP_CERTIFICATE_SUBJECT" validate:"required"`
	// Lifetime is either a duration string (e.g. "24h") or an RFC 3339
	// timestamp, applied as the certificate's NotAfter.
	Lifetime string `env:"STEP_CERTIFICATE_LIFETIME" default:"24h"`
}

type cloudflareConfig struct {
	// Cloudflare API token with the following permissions:
	// - Zone SSL and Certificates Read/Write
	Token  string `env:"CF_API_TOKEN" validate:"required"`
	ZoneID string `env:"CF_ZONE_ID" validate:"required"`
	// RenewalThreshold is the minimum remaining validity the current
	// certificate must have; if it still has more time left than this,
	// renewal is skipped.
	RenewalThreshold time.Duration `env:"CF_RENEWAL_THRESHOLD" default:"168h"`
}

type notifyConfig struct {
	Targets  []string `env:"APP_NOTIFY_TARGETS" envSeparator:"\n" default:"[]"`
	Template string   `env:"APP_NOTIFY_TEMPLATE" default:"Cloudflare mTLS certificate updated for zone {{.ZoneID}} with expiry after {{.NotAfter}}"`
}

type Config struct {
	Logging     loggingConfig     `default:""`
	Api         apiConfig         `default:""`
	Certificate certificateConfig `default:""`
	Cloudflare  cloudflareConfig  `default:""`
	Notify      notifyConfig      `default:""`

	Environment string `env:"APP_ENVIRONMENT" default:"production" validate:"oneof=production development"`
}
