package cloudflare

import (
	"fmt"
	"time"

	"charm.land/log/v2"
	"github.com/axelrindle/cf-stepca-mtls/pkg/stepca"
	"github.com/axelrindle/cf-stepca-mtls/pkg/util"
	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/origin_tls_client_auth"
)

const (
	certificateStatusPollInterval = 5 * time.Second
	certificateStatusPollTimeout  = 2 * time.Minute
)

func (c *client) DeployCertificate(cert *stepca.Certificate) error {
	newCert, err := c.uploadCertificate(cert)
	if err != nil {
		return fmt.Errorf("uploading certificate: %w", err)
	}
	log.Info("uploaded certificate", "id", newCert.ID)

	if err := c.waitForCertificateActive(newCert.ID); err != nil {
		return fmt.Errorf("waiting for certificate to become active: %w", err)
	}

	if err := c.ensureAuthenticatedOriginPullsEnabled(); err != nil {
		return fmt.Errorf("enabling authenticated origin pulls: %w", err)
	}

	if err := c.deleteOtherCertificates(newCert.ID); err != nil {
		return fmt.Errorf("deleting old certificates: %w", err)
	}

	return nil
}

func (c *client) uploadCertificate(cert *stepca.Certificate) (*origin_tls_client_auth.ZoneCertificateNewResponse, error) {
	ctx, cancel := util.Timeout(10 * time.Second)
	defer cancel()

	return c.api.OriginTLSClientAuth.ZoneCertificates.New(ctx, origin_tls_client_auth.ZoneCertificateNewParams{
		ZoneID:      cloudflare.String(c.zoneID),
		Certificate: cloudflare.String(string(cert.CrtPEM)),
		PrivateKey:  cloudflare.String(string(cert.KeyPEM)),
	})
}

// waitForCertificateActive polls the certificate's deployment status until
// Cloudflare reports it active on the edge, or certificateStatusPollTimeout
// elapses.
func (c *client) waitForCertificateActive(certID string) error {
	deadline := time.Now().Add(certificateStatusPollTimeout)

	for {
		ctx, cancel := util.Timeout(10 * time.Second)
		res, err := c.api.OriginTLSClientAuth.ZoneCertificates.Get(ctx, certID, origin_tls_client_auth.ZoneCertificateGetParams{
			ZoneID: cloudflare.String(c.zoneID),
		})
		cancel()
		if err != nil {
			return err
		}

		switch res.Status {
		case origin_tls_client_auth.ZoneAuthenticatedOriginPullStatusActive:
			log.Info("certificate is active", "id", certID)
			return nil
		case origin_tls_client_auth.ZoneAuthenticatedOriginPullStatusDeploymentTimedOut,
			origin_tls_client_auth.ZoneAuthenticatedOriginPullStatusDeletionTimedOut,
			origin_tls_client_auth.ZoneAuthenticatedOriginPullStatusDeleted:
			return fmt.Errorf("certificate %s reached unexpected status %q", certID, res.Status)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("certificate %s did not become active within %s (last status %q)", certID, certificateStatusPollTimeout, res.Status)
		}

		log.Debug("waiting for certificate to become active", "id", certID, "status", res.Status)
		time.Sleep(certificateStatusPollInterval)
	}
}

// ensureAuthenticatedOriginPullsEnabled checks whether the zone-level
// authenticated origin pulls feature is enabled and turns it on if it isn't.
// Cloudflare ignores an uploaded certificate until this is set.
func (c *client) ensureAuthenticatedOriginPullsEnabled() error {
	ctx, cancel := util.Timeout(10 * time.Second)
	setting, err := c.api.OriginTLSClientAuth.Settings.Get(ctx, origin_tls_client_auth.SettingGetParams{
		ZoneID: cloudflare.String(c.zoneID),
	})
	cancel()
	if err != nil {
		return err
	}

	if setting.Enabled {
		return nil
	}

	log.Info("authenticated origin pulls disabled, enabling")

	ctx, cancel = util.Timeout(10 * time.Second)
	defer cancel()

	_, err = c.api.OriginTLSClientAuth.Settings.Update(ctx, origin_tls_client_auth.SettingUpdateParams{
		ZoneID:  cloudflare.String(c.zoneID),
		Enabled: cloudflare.Bool(true),
	})
	return err
}

// deleteOtherCertificates removes every zone-level client certificate except
// keepID. Cloudflare only uses one certificate at a time, so any other
// certificate left behind is stale.
func (c *client) deleteOtherCertificates(keepID string) error {
	ctx, cancel := util.Timeout(10 * time.Second)
	certs, err := c.api.OriginTLSClientAuth.ZoneCertificates.List(ctx, origin_tls_client_auth.ZoneCertificateListParams{
		ZoneID: cloudflare.String(c.zoneID),
	})
	cancel()
	if err != nil {
		return err
	}

	for _, cert := range certs.Result {
		if cert.ID == keepID {
			continue
		}

		log.Info("deleting old certificate", "id", cert.ID)

		delCtx, delCancel := util.Timeout(10 * time.Second)
		_, err := c.api.OriginTLSClientAuth.ZoneCertificates.Delete(delCtx, cert.ID, origin_tls_client_auth.ZoneCertificateDeleteParams{
			ZoneID: cloudflare.String(c.zoneID),
		})
		delCancel()
		if err != nil {
			return fmt.Errorf("deleting certificate %s: %w", cert.ID, err)
		}
	}

	return nil
}
