# Radar Example

Demonstrates the Radar application with three workers that use health probes and custom metrics.

## Workers

**db_worker** -- Simulates a database service. Registers a liveness+readiness signal with 10-second heartbeat timeout. Sends heartbeats every 3 seconds. Reports a `db_connections` gauge with random fluctuations.

**cache_worker** -- Simulates a cache service. Registers a liveness signal with 8-second heartbeat timeout. Sends heartbeats every 2 seconds. Increments a `cache_operations_total` counter with random hit/miss counts every second.

**api_worker** -- Simulates an API gateway. Registers a readiness+startup signal with no heartbeat (uses manual `ServiceUp`/`ServiceDown`). Records `http_request_duration_seconds` histogram with random latencies for simulated GET/POST requests.

All workers are supervised under a one-for-one supervisor. The workers application depends on Radar -- it starts only after Radar is running.

## Running

```bash
go run ./cmd
```

With mailbox latency metrics enabled:

```bash
go run -tags latency ./cmd
```

## Endpoints

```
http://localhost:9090/health/live
http://localhost:9090/health/ready
http://localhost:9090/health/startup
http://localhost:9090/metrics
```

## Checking

```bash
# health probes
curl http://localhost:9090/health/live
curl http://localhost:9090/health/ready

# prometheus metrics (includes base ergo metrics + custom db/cache/api metrics)
curl http://localhost:9090/metrics
```
