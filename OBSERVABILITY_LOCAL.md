# Local Observability Stack

This repository can run a local log observability stack with:

- Loki for log storage and query
- Promtail for collecting Luca API JSON logs and shipping to Loki
- Grafana for exploration

## Why Promtail

Promtail is selected for local stack because it is the smallest reliable integration path for this codebase:

- Luca API already writes structured JSON logs to stdout
- Promtail can tail local log files directly with minimal setup
- Native Loki client and label pipeline are straightforward for Admin log query use cases

## Files

- `docker-compose.observability.yml`
- `observability/loki-config.yaml`
- `observability/promtail-config.yaml`
- `observability/grafana/provisioning/datasources/loki.yaml`
- `observability_up.sh`
- `observability_down.sh`
- `run_with_observability.sh`

## Start Stack And API

```bash
./run.sh --observable
```

Production-shaped local run:

```bash
./run.sh --observable --env=production
```

This command will:

- start Loki, Promtail, and Grafana
- mirror Luca API stdout/stderr into the Promtail watched file
- run the backend with the selected `APP_ENV`

If you want to start only the observability stack without running the API, you can still use:

```bash
./observability_up.sh
```

After start:

- Grafana: `http://127.0.0.1:3001` (`admin/admin`)
- Loki ready endpoint: `http://127.0.0.1:3100/ready`

Promtail tails:

- `tmp/observability/logs/luca_api.json.log`

## Pipeline

Promtail pipeline in `observability/promtail-config.yaml`:

1. Input: tail `tmp/observability/logs/*.log`
2. Parse: JSON fields (`level`, `ts`, `message`, `service`, `module`, `env`, `request_id`, `user_id`, `department_id`, `source`, `stacktrace`)
3. Labels: stable fields (`app`, `job`, `level`, `service`, `module`, `env`)
4. Output: push full log line to Loki (`/loki/api/v1/push`)

High-cardinality fields (`request_id`, `user_id`, `department_id`) are parsed but not promoted to labels.

## Quick Check

Query recent warn/error logs from Loki:

```bash
curl -G "http://127.0.0.1:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={app="luca_api"} | json | level=~"warn|error"' \
  --data-urlencode 'limit=20'
```

## Stop Stack

```bash
./observability_down.sh
```
