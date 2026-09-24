# Webhook

Manages a Pathly outbound webhook (`POST/DELETE /v1/webhooks`).

The destination URL is write-only. The signing secret is returned once on create
and published to `writeConnectionSecretToRef` when set.

## Example

See `examples/webhook.yaml`.
