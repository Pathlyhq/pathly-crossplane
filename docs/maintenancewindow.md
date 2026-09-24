# MaintenanceWindow

Manages a Pathly maintenance window (`POST/DELETE /v1/maintenance-windows`).

Supports absolute windows (`startsAt` / `endsAt`) or weekly recurrence
(`weekday`, `startMinute`, `durationMin`).

## Example

See `examples/maintenancewindow.yaml`.
