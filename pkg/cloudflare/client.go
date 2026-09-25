package cloudflare

import (
	"time"

	"github.com/axelrindle/cf-stepca-mtls/pkg/stepca"
	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
)

type CloudflareClient interface {
	// DeployCertificate uploads cert as the zone's client certificate for
	// authenticated origin pulls, waits for Cloudflare to activate it
	// (enabling the feature if it isn't already enabled), and then removes
	// every other certificate configured on the zone.
	DeployCertificate(cert *stepca.Certificate) error

	// GetCertificateExpiry returns the soonest expiry among the zone's
	// currently configured client certificates, or the zero time if none is
	// configured yet.
	GetCertificateExpiry() (time.Time, error)
}

type client struct {
	zoneID string
	api    *cloudflare.Client
}

func NewClient(token, zoneID string) CloudflareClient {
	return &client{
		zoneID: zoneID,
		api:    cloudflare.NewClient(option.WithAPIToken(token)),
	}
}
