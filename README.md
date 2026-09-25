# Cloudflare <> Step CA mTLS automation

Automatically deploys mTLS certificates to Cloudflare issued by Step CA.

## Prerequisites

- [Step CA JWK provisioner](https://smallstep.com/docs/step-ca/provisioners/#jwk)
- Cloudflare-managed domain (zone)
- Cloudflare Access Token with the following permissions:
    - Zone SSL and Certificates Read/Write

## Installation

### Docker

```sh
docker run --rm \
  -e STEP_API_ENDPOINT=https://ca.example.com \
  -e STEP_CERTIFICATE_PROVISIONER=my-provisioner \
  -e STEP_CERTIFICATE_PROVISIONER_PASSWORD=changeme \
  -e STEP_CERTIFICATE_SUBJECT=mtls.example.com \
  -e CF_API_TOKEN=changeme \
  -e CF_ZONE_ID=changeme \
  ghcr.io/axelrindle/cf-stepca-mtls
```

The image is based on `distroless/static-debian13:nonroot` and runs as a non-root user (UID/GID 1000).

### Kubernetes

The bundled Helm chart at [charts/cf-stepca-mtls](charts/cf-stepca-mtls) deploys the tool as a `CronJob`:

```sh
helm install cf-stepca-mtls ./charts/cf-stepca-mtls \
  --set env[0].name=STEP_API_ENDPOINT \
  --set env[0].value=https://ca.example.com
```

Further env vars (also as `secretKeyRef`) can be set via `env`/`envFrom` in [values.yaml](charts/cf-stepca-mtls/values.yaml). The CronJob defaults to running on the 1st of every 2nd month (approximation for a 60-day interval), configurable via `cronjob.schedule`.

## Configuration

Configuration is done entirely via environment variables.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `APP_LOGGING_LEVEL` | no | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `APP_ENVIRONMENT` | no | `production` | `production` or `development` |
| `STEP_API_ENDPOINT` | yes | - | URL of the Step CA API |
| `STEP_CERTIFICATE_PROVISIONER` | yes | - | Name of the Step CA provisioner |
| `STEP_CERTIFICATE_PROVISIONER_PASSWORD` | yes | - | Password of the provisioner |
| `STEP_CERTIFICATE_SUBJECT` | yes | - | Subject/CN of the certificate to issue |
| `STEP_CERTIFICATE_LIFETIME` | no | `24h` | Validity, either as a duration (`24h`) or an RFC 3339 timestamp |
| `CF_API_TOKEN` | yes | - | Cloudflare API token with `Zone SSL and Certificates Read/Write` permission |
| `CF_ZONE_ID` | yes | - | ID of the Cloudflare zone |
| `CF_RENEWAL_THRESHOLD` | no | `168h` | Minimum remaining validity below which renewal triggers |

## Thank you!

- Cloudflare for their awesome platform
- Step CA for the simple CA service

## License

[MIT](LICENSE)
