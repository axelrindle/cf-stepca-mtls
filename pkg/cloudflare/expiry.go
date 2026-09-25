package cloudflare

import (
	"time"

	"github.com/axelrindle/cf-stepca-mtls/pkg/util"
	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/origin_tls_client_auth"
)

func (c *client) GetCertificateExpiry() (time.Time, error) {
	ctx, cancel := util.Timeout(10 * time.Second)
	defer cancel()

	certs, err := c.api.OriginTLSClientAuth.ZoneCertificates.List(ctx, origin_tls_client_auth.ZoneCertificateListParams{
		ZoneID: cloudflare.String(c.zoneID),
	})
	if err != nil {
		return time.Time{}, err
	}

	var expiry time.Time
	for _, cert := range certs.Result {
		if expiry.IsZero() || cert.ExpiresOn.Before(expiry) {
			expiry = cert.ExpiresOn
		}
	}

	return expiry, nil
}
