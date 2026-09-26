package config

import "time"

type loggingConfig struct {
	// Log level (`debug`, `info`, `warn`, `error`)
	Level string `env:"APP_LOGGING_LEVEL" default:"info" validate:"oneof=debug info warn error"`
}

type apiConfig struct {
	// URL of the Step CA instance
	Endpoint string `env:"STEP_API_ENDPOINT"`

	// Name of the Step CA provisioner
	Provisioner string `env:"STEP_CERTIFICATE_PROVISIONER" validate:"required"`

	// Password of the provisioner
	ProvisionerPassword string `env:"STEP_CERTIFICATE_PROVISIONER_PASSWORD" validate:"required"`
}

type certificateConfig struct {
	// Subject/CN of the certificate to issue
	Subject string `env:"STEP_CERTIFICATE_SUBJECT" validate:"required"`

	// Either a duration string (e.g. "24h") or an RFC 3339
	// timestamp, applied as the certificate's NotAfter.
	Lifetime string `env:"STEP_CERTIFICATE_LIFETIME" default:"24h"`
}

type cloudflareConfig struct {
	// Cloudflare API token with the following permissions:
	// - Zone SSL and Certificates Read/Write
	Token string `env:"CF_API_TOKEN" validate:"required"`

	// ID of the Cloudflare zone
	// Can be copied on the zone dashboard
	ZoneID string `env:"CF_ZONE_ID" validate:"required"`

	// Minimum remaining validity below which renewal triggers
	RenewalThreshold time.Duration `env:"CF_RENEWAL_THRESHOLD" default:"168h"`
}

type notifyConfig struct {
	// Notification targets [supported by shoutrrr](https://shoutrrr.nickfedor.com/v0.21.1/services/overview/)
	Targets []string `env:"APP_NOTIFY_TARGETS" envSeparator:"\n" default:"[]"`

	// Notification message template. [See below](#notifications) for more information.
	Template string `env:"APP_NOTIFY_TEMPLATE" default:"Cloudflare mTLS certificate updated for zone {{.ZoneID}} with expiry after {{.NotAfter}}"`
}

type Config struct {
	Logging     loggingConfig     `default:""`
	Api         apiConfig         `default:""`
	Certificate certificateConfig `default:""`
	Cloudflare  cloudflareConfig  `default:""`
	Notify      notifyConfig      `default:""`

	// `development` should only be used in local development
	Environment string `env:"APP_ENVIRONMENT" default:"production" validate:"oneof=production development"`
}
