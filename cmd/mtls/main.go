package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/axelrindle/cf-stepca-mtls/internal/config"
	"github.com/axelrindle/cf-stepca-mtls/pkg/cloudflare"
	"github.com/axelrindle/cf-stepca-mtls/pkg/stepca"
)

//go:embed banner.txt
var banner string

var (
	Version        = "dev"
	CommitHash     = "unknown"
	BuildTimestamp = "unknown"
)

var (
	flagShowVersion bool
)

func buildVersion() string {
	return fmt.Sprintf("%s-%s (%s)", Version, CommitHash, BuildTimestamp)
}

func init() {
	fmt.Fprintf(os.Stderr, banner, buildVersion())

	flag.BoolVar(&flagShowVersion, "version", false, "show program version")
	flag.Parse()
	if flagShowVersion {
		os.Exit(0)
	}

	log.SetOutput(os.Stderr)
}

func main() {
	cfg := config.Load()

	log.SetLevel(cfg.CharmLoggerLevel())

	cf := cloudflare.NewClient(cfg.Cloudflare.Token, cfg.Cloudflare.ZoneID)

	skip, err := cloudflare.ShouldSkipRenewal(cf, cfg.Cloudflare.RenewalThreshold)
	if err != nil {
		log.Fatal(err)
	}
	if skip {
		log.Info("existing certificate still valid, skipping renewal")
		return
	}

	step, err := stepca.NewClient(cfg.Api.Endpoint,
		stepca.NewProvisioner(cfg.Api.Provisioner, cfg.Api.ProvisionerPassword))
	if err != nil {
		log.Fatal(err)
	}

	cert, err := step.RequestCertificate(cfg.Certificate.Subject, cfg.Certificate.Lifetime)
	if err != nil {
		log.Fatal(err)
	}

	if err := cf.DeployCertificate(cert); err != nil {
		log.Fatal(err)
	}
}
