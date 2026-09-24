# Pathly Crossplane provider — docs

## HTTP scenarios only

The `Scenario` CRD accepts `type: http` exclusively (**HTTP-only policy**). Pathly
browser journeys store credentials inside an encrypted scenario AST. Exposing
those steps through a Kubernetes Custom Resource would copy them into etcd,
backups and audit logs.

Manage browser scenarios through the [Pathly console](https://pathlyhq.com) or
the HTTP API with a short-lived token
([developers](https://pathlyhq.com/en/developers)). Use this provider for HTTP
health checks, webhooks, maintenance windows and SLA targets.

## Credentials

Create `PATHLY_API_TOKEN` at [pathlyhq.com](https://pathlyhq.com), store it in a
Kubernetes Secret, and reference that Secret from `ProviderConfig`:

```yaml
apiVersion: pathly.pathlyhq.com/v1alpha1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentialsSource: Secret
  credentialsSecretRef:
    namespace: crossplane-system
    name: pathly-creds
    key: token
```

The Secret must contain a Pathly organisation API token (`sp_…`) with the scopes
required by the resources you manage (`scenarios:write`, `webhooks:write`, etc.).

## Idempotency and retries

Creates send `Idempotency-Key` derived from the object UID. The shared HTTP
client retries on transport errors, `429` and `5xx`, honouring `Retry-After`
(capped at 90 seconds). Client errors (`4xx` other than 429) are not retried.

## External name

After create, the Pathly resource id is stored in
`metadata.annotations["crossplane.io/external-name"]`.
