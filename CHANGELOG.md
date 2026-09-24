# Changelog

## 0.1.0 — 2026-09-23

- First publish: Crossplane provider for Pathly HTTP scenarios, webhooks,
  maintenance windows and SLA targets.
- Auth via Kubernetes Secret mapped to `PATHLY_API_TOKEN`.
- Client retries with `Retry-After` (max 4, 90 s cap) and `Idempotency-Key` on creates.
