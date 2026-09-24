# Scenario

Manages a Pathly **HTTP** scenario (`POST/PATCH/DELETE /v1/scenarios`).

Browser journeys are intentionally unsupported. See [overview.md](overview.md).

## Example

```yaml
apiVersion: pathly.pathlyhq.com/v1alpha1
kind: Scenario
metadata:
  name: api-health
spec:
  providerConfigRef:
    name: default
  forProvider:
    name: API health
    type: http
    url: https://api.example.com/health
    method: GET
    expectedStatus: 200
    intervalSec: 300
    enabled: true
```
