# Alert -> AI Webhook Spec

## 1. Purpose

- Define the inbound webhook contract from Grafana Cloud alerting to the internal AI automation service.
- Standardize authentication, idempotency, retries, and payload parsing.

## 2. Endpoint

- Method: `POST`
- Path: `/v1/alerts/grafana`
- Content-Type: `application/json`

## 3. Authentication and Integrity

- Header: `X-Webhook-Signature: sha256=<hex_hmac>`
- Header: `X-Webhook-Timestamp: <unix_epoch_seconds>`
- HMAC input: `<timestamp>.<raw_request_body>`
- HMAC key: secret stored in secret manager.
- Reject if:
  - signature mismatch
  - timestamp skew > 5 minutes
  - missing required headers

## 4. Idempotency and Ordering

- Header: `X-Event-Id` (preferred from sender; otherwise generated hash of body).
- Store processed event IDs with TTL (24h recommended).
- Duplicate events must return `200` with `"status":"duplicate_ignored"`.
- Ordering is not guaranteed; processing must be event-time aware.

## 5. Request Payload (v1)

```json
{
  "version": "1.0",
  "source": "grafana-cloud-alerting",
  "event_id": "evt_123",
  "fired_at": "2026-02-19T18:45:00Z",
  "environment": "staging",
  "alert": {
    "rule_uid": "abcdEF12",
    "rule_name": "High API Error Rate",
    "state": "firing",
    "severity": "critical",
    "labels": {
      "service": "backend-api",
      "team": "platform",
      "region": "us-central1"
    },
    "annotations": {
      "summary": "5xx error rate exceeded threshold",
      "runbook_url": "https://example.com/runbooks/api-errors"
    }
  },
  "window": {
    "start": "2026-02-19T18:35:00Z",
    "end": "2026-02-19T18:45:00Z"
  },
  "links": {
    "grafana_alert": "https://grafana.example/alerting/...",
    "dashboard": "https://grafana.example/d/..."
  }
}
```

## 6. Required Fields

- `version`
- `source`
- `event_id`
- `fired_at`
- `environment`
- `alert.rule_uid`
- `alert.rule_name`
- `alert.state`
- `alert.severity`
- `window.start`
- `window.end`

## 7. Processing Contract

- Validate auth/signature and required fields.
- Persist raw payload and metadata for audit.
- Build enrichment query plan:
  - metrics queries (same `service`, `environment`, time window)
  - logs queries (same labels + error patterns)
  - trace lookups (high latency/error traces in window)
  - optional Sentry issue lookup
- Run AI diagnosis pipeline and produce:
  - probable root causes (ranked)
  - confidence score
  - suggested next actions
  - evidence links
- Send output to incident destination (incident.io/Slack/ticketing).

## 8. Response Contract

- `200` accepted/processed:

```json
{
  "status": "accepted",
  "event_id": "evt_123",
  "analysis_id": "an_456"
}
```

- `200` duplicate:

```json
{
  "status": "duplicate_ignored",
  "event_id": "evt_123"
}
```

- `400` invalid payload:

```json
{
  "status": "invalid_request",
  "error": "missing_required_field",
  "field": "alert.rule_uid"
}
```

- `401` invalid signature:

```json
{
  "status": "unauthorized",
  "error": "invalid_signature"
}
```

## 9. Retry Policy

- Sender retries on non-2xx with exponential backoff.
- Receiver should be fast-fail on validation/auth errors.
- Receiver should enqueue long-running enrichment/AI work asynchronously.

## 10. Operational Requirements

- Structured logs for each event (`event_id`, `analysis_id`, `alert.rule_uid`, `environment`).
- Metrics:
  - webhook requests total
  - auth failures total
  - parse failures total
  - duplicate events total
  - analysis latency histogram
- Alert if:
  - auth failures spike
  - processing latency breaches SLO
  - enqueue failures occur

## 11. Versioning

- Backward-compatible additions allowed within `1.x`.
- Breaking changes require `2.0` and new endpoint path `/v2/alerts/grafana`.
