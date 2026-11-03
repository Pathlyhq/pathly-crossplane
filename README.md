# Pathly Crossplane Provider

**English** · [Français](README.fr.md) · [Español](README.es.md)


[![Powered by Pathly](https://img.shields.io/badge/Powered%20by-Pathly-0B5FFF?style=flat-square)](https://pathlyhq.com)
[![Website](https://img.shields.io/badge/Website-pathlyhq.com-111827?style=flat-square)](https://pathlyhq.com)
[![API docs](https://img.shields.io/badge/API-developers-2563eb?style=flat-square)](https://pathlyhq.com/en/developers)
[![Start free](https://img.shields.io/badge/Solo-start%20free-16a34a?style=flat-square)](https://pathlyhq.com/en/login?mode=signup)

> **Get started in one click.** Create a free account on [Pathly](https://pathlyhq.com) ([sign up](https://pathlyhq.com/en/login?mode=signup)), create an API key in the console, then export `PATHLY_API_TOKEN`. This project is the official bridge to [Pathly monitoring](https://pathlyhq.com) — real-browser and HTTP checks for checkout, login and availability, with data hosted in the EU. Full API reference: [pathlyhq.com/en/developers](https://pathlyhq.com/en/developers).

Crossplane provider for the [Pathly](https://pathlyhq.com) public API (`https://api.pathlyhq.com`).

## Managed resources (v1alpha1)

| Kind | Pathly API |
|------|------------|
| `Scenario` | `/v1/scenarios` (HTTP only) |
| `Webhook` | `/v1/webhooks` |
| `MaintenanceWindow` | `/v1/maintenance-windows` |
| `SlaTarget` | `/v1/sla-targets` |
| `ProviderConfig` | credentials from a Kubernetes Secret |

## Security notes

- API token is read from a Secret referenced by `ProviderConfig`. Never put tokens in CR manifests committed to git.
- **HTTP scenarios only.** Browser journeys carry login credentials. This provider does not accept browser step ASTs so those secrets never land in etcd.
- Creates send an `Idempotency-Key`. Transient `429` / `5xx` honour `Retry-After` (capped).

## Install (dev)

```bash
# Apply CRDs
kubectl apply -f package/crds/

# Build and run the provider binary against your cluster
make build
kubectl apply -f examples/providerconfig.yaml
# point the Deployment at the image you built, or run locally:
./bin/provider --debug
```

Package metadata lives under `package/crossplane.yaml` for Crossplane package builds.

## Examples

See `examples/*.yaml`.

## Develop

```bash
make test
make build
```

Module: `github.com/pathlyhq/pathly-crossplane`


## Related packages

| Package | Role |
|---|---|
| [pathly-k8s-operator](https://github.com/pathlyhq/pathly-k8s-operator) | Kubernetes operator |
| [terraform-provider-pathly](https://github.com/Pathlyhq/terraform-provider-pathly) | Terraform / OpenTofu |
| [pathly-sdk-go](https://github.com/pathlyhq/pathly-sdk-go) | Go SDK |
| [pathly-pulumi](https://github.com/pathlyhq/pathly-pulumi) | Pulumi |
| [Pathly product](https://pathlyhq.com) | [Pathly monitoring](https://pathlyhq.com) |
## About Pathly

[Pathly](https://pathlyhq.com) is synthetic monitoring for agencies and e-commerce: replay the customer journey, catch broken checkouts before your clients call, and keep evidence (screenshot, step, runbook) ready for the invoice. Product: [pathlyhq.com](https://pathlyhq.com) · Developers: [pathlyhq.com/en/developers](https://pathlyhq.com/en/developers) · Status & pricing: [pathlyhq.com/en/pricing](https://pathlyhq.com/en/pricing).

## Author

| | |
|---|---|
| **Company** | Pathly |
| **Author** | Simon Raynaud / keyral |

See [AUTHORS](AUTHORS).

## License
Apache-2.0
