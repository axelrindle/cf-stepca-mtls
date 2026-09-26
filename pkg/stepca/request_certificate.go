package stepca

import (
	"crypto"
	"encoding/pem"
	"fmt"

	"charm.land/log/v2"
	"github.com/smallstep/certificates/api"
	"go.step.sm/crypto/keyutil"
	"go.step.sm/crypto/pemutil"
	"go.step.sm/crypto/x509util"
)

func (c client) RequestCertificate(subject, lifetime string) (*Certificate, error) {
	notAfter, err := api.ParseTimeDuration(lifetime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate lifetime %q: %w", lifetime, err)
	}

	log.Info("generating private key")
	_, priv, err := keyutil.GenerateDefaultKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	signer, ok := priv.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("generated private key of type %T does not implement crypto.Signer", priv)
	}

	csr, err := x509util.CreateCertificateRequest(subject, []string{subject}, signer)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate request: %w", err)
	}

	token, err := c.provisioner.Token(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to generate provisioner token: %w", err)
	}

	log.Info("requesting certificate", "subject", subject, "lifetime", lifetime, "provisioner", c.provisioner.Name())
	resp, err := c.provisioner.Sign(&api.SignRequest{
		CsrPEM:   api.NewCertificateRequest(csr),
		OTT:      token,
		NotAfter: notAfter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign certificate request: %w", err)
	}

	crtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: resp.ServerPEM.Raw})
	crtPEM = append(crtPEM, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: resp.CaPEM.Raw})...)

	keyBlock, err := pemutil.Serialize(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize private key: %w", err)
	}

	log.Info("certificate obtained", "subject", subject)
	return &Certificate{
		CrtPEM:   crtPEM,
		KeyPEM:   pem.EncodeToMemory(keyBlock),
		NotAfter: notAfter.Time(),
	}, nil
}
