# Provider Crossplane Pathly

[English](README.md) · [Français](README.fr.md) · **Español**


[![Powered by Pathly](https://img.shields.io/badge/Powered%20by-Pathly-0B5FFF?style=flat-square)](https://pathlyhq.com)
[![Website](https://img.shields.io/badge/Website-pathlyhq.com-111827?style=flat-square)](https://pathlyhq.com)
[![API docs](https://img.shields.io/badge/API-developers-2563eb?style=flat-square)](https://pathlyhq.com/es/developers)
[![Start free](https://img.shields.io/badge/Solo-start%20free-16a34a?style=flat-square)](https://pathlyhq.com/es/login?mode=signup)

> **Empiece en un clic.** Cree una cuenta gratuita en [Pathly](https://pathlyhq.com) ([registro](https://pathlyhq.com/es/login?mode=signup)), genere una clave API en la consola y exporte `PATHLY_API_TOKEN`. Este repositorio es el puente oficial hacia [la monitorización Pathly](https://pathlyhq.com): comprobaciones HTTP y de navegador (carrito, login, disponibilidad), con datos en la UE. Referencia API: [pathlyhq.com/es/developers](https://pathlyhq.com/es/developers).

Provider Crossplane para la API pública [Pathly](https://pathlyhq.com) (`https://api.pathlyhq.com`).

## Recursos gestionados (v1alpha1)

| Kind | API Pathly |
|------|------------|
| `Scenario` | `/v1/scenarios` (solo HTTP) |
| `Webhook` | `/v1/webhooks` |
| `MaintenanceWindow` | `/v1/maintenance-windows` |
| `SlaTarget` | `/v1/sla-targets` |
| `ProviderConfig` | credenciales desde un Secret de Kubernetes |

## Notas de seguridad

- El token de API se lee desde un Secret referenciado por `ProviderConfig`. Nunca ponga tokens en manifiestos CR versionados en git.
- **Solo escenarios HTTP.** Los recorridos de navegador llevan credenciales de acceso. Este provider no acepta el AST de pasos browser, para que esos secretos no acaben en etcd.
- Las creaciones envían una `Idempotency-Key`. Los `429` / `5xx` transitorios respetan `Retry-After` (con tope).

## Instalación (dev)

```bash
# Aplicar CRDs
kubectl apply -f package/crds/

# Compilar y ejecutar el binario del provider contra su clúster
make build
kubectl apply -f examples/providerconfig.yaml
# apunte el Deployment a la imagen que construyó, o ejecute en local:
./bin/provider --debug
```

Los metadatos del paquete están en `package/crossplane.yaml` para builds de paquetes Crossplane.

## Ejemplos

Véase `examples/*.yaml`.

## Desarrollo

```bash
make test
make build
```

Módulo: `github.com/pathlyhq/pathly-crossplane`


## Paquetes relacionados

| Package | Role |
|---|---|
| [pathly-k8s-operator](https://github.com/pathlyhq/pathly-k8s-operator) | Kubernetes operator |
| [pathly-terraform-provider](https://github.com/pathlyhq/pathly-terraform-provider) | Terraform / OpenTofu |
| [pathly-sdk-go](https://github.com/pathlyhq/pathly-sdk-go) | Go SDK |
| [pathly-pulumi](https://github.com/pathlyhq/pathly-pulumi) | Pulumi |
| [Pathly product](https://pathlyhq.com) | [Pathly monitoring](https://pathlyhq.com) |
## Acerca de Pathly

[Pathly](https://pathlyhq.com) es monitorización sintética para agencias y e-commerce: reproduce el recorrido del cliente, detecta un checkout roto antes de la llamada, y deja la prueba (captura, paso, runbook) lista para la factura. Producto: [pathlyhq.com](https://pathlyhq.com) · Desarrolladores: [pathlyhq.com/es/developers](https://pathlyhq.com/es/developers) · Precios: [pathlyhq.com/es/pricing](https://pathlyhq.com/es/pricing).

## Autor

| | |
|---|---|
| **Empresa** | Pathly |
| **Autor** | Simon Raynaud / keyral |

Véase [AUTHORS](AUTHORS).

## Licencia
Apache-2.0
