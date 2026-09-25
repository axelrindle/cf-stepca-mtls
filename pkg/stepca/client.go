package stepca

import (
	"fmt"

	"github.com/smallstep/certificates/ca"
)

type StepCaClient interface {
	RequestCertificate(subject, lifetime string) (*Certificate, error)
}

type provisioner struct {
	name, password string
}

func NewProvisioner(name, password string) provisioner {
	return provisioner{
		name, password,
	}
}

type client struct {
	caUrl       string
	provisioner *ca.Provisioner
}

func NewClient(caUrl string, p provisioner) (StepCaClient, error) {
	provisioner, err := ca.NewProvisioner(p.name, "", caUrl, []byte(p.password))
	if err != nil {
		return nil, fmt.Errorf("failed to load provisioner %q: %w", p.name, err)
	}

	return &client{
		caUrl,
		provisioner,
	}, nil
}
