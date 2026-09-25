package cloudflare

import "time"

// ShouldSkipRenewal checks the expiry of the zone's currently configured
// client certificate and reports whether it is still valid for longer than
// threshold, in which case obtaining a new certificate can be skipped. It
// never skips when no certificate is configured yet.
func ShouldSkipRenewal(cf CloudflareClient, threshold time.Duration) (bool, error) {
	if threshold == 0 {
		return false, nil
	}

	expiresOn, err := cf.GetCertificateExpiry()
	if err != nil {
		return false, err
	}

	if expiresOn.IsZero() {
		return false, nil
	}

	return time.Until(expiresOn) > threshold, nil
}
