# Provider Crossplane Pathly

[English](README.md) · **Français** · [Español](README.es.md)


[![Powered by Pathly](https://img.shields.io/badge/Powered%20by-Pathly-0B5FFF?style=flat-square)](https://pathlyhq.com)
[![Website](https://img.shields.io/badge/Website-pathlyhq.com-111827?style=flat-square)](https://pathlyhq.com)
[![API docs](https://img.shields.io/badge/API-developers-2563eb?style=flat-square)](https://pathlyhq.com/fr/developers)
[![Start free](https://img.shields.io/badge/Solo-start%20free-16a34a?style=flat-square)](https://pathlyhq.com/fr/login?mode=signup)

> **Démarrage en un clic.** Créez un compte gratuit sur [Pathly](https://pathlyhq.com) ([inscription](https://pathlyhq.com/fr/login?mode=signup)), générez une clé API dans la console, puis exportez `PATHLY_API_TOKEN`. Ce dépôt est le pont officiel vers [la surveillance Pathly](https://pathlyhq.com) — contrôles HTTP et parcours navigateur (panier, connexion, disponibilité), données hébergées dans l’UE. Référence API : [pathlyhq.com/fr/developers](https://pathlyhq.com/fr/developers).

Provider Crossplane pour l’API publique [Pathly](https://pathlyhq.com) (`https://api.pathlyhq.com`).

## Ressources gérées (v1alpha1)

| Kind | API Pathly |
|------|------------|
| `Scenario` | `/v1/scenarios` (HTTP uniquement) |
| `Webhook` | `/v1/webhooks` |
| `MaintenanceWindow` | `/v1/maintenance-windows` |
| `SlaTarget` | `/v1/sla-targets` |
| `ProviderConfig` | identifiants depuis un Secret Kubernetes |

## Sécurité

- Le jeton API est lu depuis un Secret référencé par `ProviderConfig`. Ne jamais committer un jeton dans un manifeste.
- **Scénarios HTTP uniquement.** Les parcours navigateur portent des identifiants. Ce provider n’accepte pas l’AST des étapes navigateur, pour que ces secrets n’atterrissent pas dans etcd.
- Les créations envoient une `Idempotency-Key`. Les `429` / `5xx` transitoires respectent `Retry-After` (plafonné).

## Installation (dev)

```bash
kubectl apply -f package/crds/
make build
kubectl apply -f examples/providerconfig.yaml
./bin/provider --debug
```

## Exemples

Voir `examples/*.yaml`.

## Développement

```bash
make test
make build
```

Module : `github.com/pathlyhq/pathly-crossplane`


## Packages associés

| Package | Role |
|---|---|
| [pathly-k8s-operator](https://github.com/pathlyhq/pathly-k8s-operator) | Kubernetes operator |
| [pathly-terraform-provider](https://github.com/pathlyhq/pathly-terraform-provider) | Terraform / OpenTofu |
| [pathly-sdk-go](https://github.com/pathlyhq/pathly-sdk-go) | Go SDK |
| [pathly-pulumi](https://github.com/pathlyhq/pathly-pulumi) | Pulumi |
| [Pathly product](https://pathlyhq.com) | [Pathly monitoring](https://pathlyhq.com) |
## À propos de Pathly

[Pathly](https://pathlyhq.com) surveille les parcours clients des agences et e-commerçants : rejoue le tunnel, détecte un checkout cassé avant l’appel du client, et joint la preuve (capture, étape, consigne) à la facture de maintenance. Produit : [pathlyhq.com](https://pathlyhq.com) · Développeurs : [pathlyhq.com/fr/developers](https://pathlyhq.com/fr/developers) · Tarifs : [pathlyhq.com/fr/pricing](https://pathlyhq.com/fr/pricing).

## Auteur

| | |
|---|---|
| **Entreprise** | Pathly |
| **Auteur** | Simon Raynaud / keyral |

Voir [AUTHORS](AUTHORS).

## Licence
Apache-2.0
