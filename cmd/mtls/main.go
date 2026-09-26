package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"charm.land/log/v2"
	"github.com/axelrindle/cf-stepca-mtls/internal/config"
	"github.com/axelrindle/cf-stepca-mtls/pkg/cloudflare"
	"github.com/axelrindle/cf-stepca-mtls/pkg/stepca"
	"github.com/lithammer/dedent"
	"github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
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

	notify, err := shoutrrr.NewSenderWithOptions(log.StandardLog(), types.SenderOptions{}, cfg.Notify.Targets...)
	if err != nil {
		log.Fatal(err)
	}

	tpl, err := template.New("shoutrrr").Parse(cfg.Notify.Template)
	if err != nil {
		log.Fatal(err)
	}

	msg := &strings.Builder{}
	tplData := map[string]any{
		"NotAfter": "foo bar",
		"Subject":  cfg.Certificate.Subject,
		"ZoneID":   cfg.Cloudflare.ZoneID,
	}
	if err := tpl.Execute(msg, tplData); err != nil {
		log.Error("notification template failed", err)
	} else {
		errs := notify.Send(dedent.Dedent(msg.String()), nil)
		for _, err := range errs {
			log.Error("send notification failed", err)
		}
	}
	return

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

	// msg := &strings.Builder{}
	// tplData := map[string]any{
	// 	"NotAfter": cert.NotAfter.UTC().Format(time.RFC1123),
	// 	"Subject":  cfg.Certificate.Subject,
	// 	"ZoneID":   cfg.Cloudflare.ZoneID,
	// }
	// if err := tpl.Execute(msg, tplData); err != nil {
	// 	log.Error("notification template failed", err)
	// } else {
	// 	errs := notify.Send(dedent.Dedent(msg.String()), nil)
	// 	for _, err := range errs {
	// 		log.Error("send notification failed", err)
	// 	}
	// }
}
