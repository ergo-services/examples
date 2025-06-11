# Docker Ergo Example with etcd

This example demonstrates how to run multiple Ergo nodes using etcd as a registrar for service discovery. It showcases service discovery, actor communication, typed configuration management, and **real-time configuration event monitoring** across a cluster.

## Architecture

- **etcd**: Distributed key-value store for service discovery and configuration
- **Node1 & Node2**: Running `myapp` application with `myactor` actors
- **Node3**: Client node with resolver actor for service discovery

## Features Demonstrated

1. **Service Discovery**: Nodes automatically discover each other through etcd
2. **Actor Communication**: Actors can send messages across nodes
3. **Typed Configuration**: Hierarchical configuration with automatic type conversion
4. **Real-time Event Monitoring**: Actors monitor registrar events using `MonitorEvent()`
5. **Configuration Updates**: Live configuration changes trigger events to actors
6. **Event System**: Cluster events for node join/leave and application lifecycle
7. **Health Monitoring**: Docker health checks and service monitoring

## Quick Start

```bash
make up              # Start all services
make setup-config    # Setup sample configuration
make logs-nodes      # View node logs
```

## Real-time Configuration Monitoring

The example demonstrates **live configuration event monitoring** using Ergo's event system:

### Event Monitoring Implementation

Each actor monitors configuration changes in real-time:

```go
// Setup event monitoring in actor Init()
func (a *myActor) Init(args ...any) error {
    // Defer monitoring setup until process is fully initialized
    a.Send(a.PID(), "init")
    return nil
}

func (a *myActor) HandleMessage(from gen.PID, message any) error {
    switch message {
    // Handle init message to setup monitoring
    case "init":
        // Get registrar event name and monitor it
        registrar, _ := a.Node().Network().Registrar()
        event, _ := registrar.Event()
        a.MonitorEvent(event)
    }
    return nil
}

// Handle events from registrar
func (a *myActor) HandleEvent(event gen.MessageEvent) error {
    switch e := event.(type) {
    case gen.MessageEventConfig:
        // Handle configuration updates
        a.handleConfigEvent(e)
    case gen.MessageEventNodeDown:
        // Handle node departures
    case gen.MessageEventNodeUp:
        // Handle node arrivals
    // ... other events
    }
}
```

### Live Configuration Updates

Test real-time configuration changes:

```bash
# Setup initial configuration
make setup-config

# Start nodes and watch logs
make logs-nodes

# In another terminal, trigger live updates
make update-config
```

You'll see actors receiving events immediately when configuration changes!

## Configuration Management

The example demonstrates the etcd registrar's hierarchical configuration system with automatic type conversion:

### Configuration Hierarchy (from highest to lowest priority):
1. **Node-specific**: `services/ergo/cluster/{cluster}/config/{node-name}/*`
2. **Application-specific**: `services/ergo/cluster/{cluster}/config/{app-name}/*` 
3. **Cluster-wide**: `services/ergo/cluster/{cluster}/config/*`
4. **Global**: `services/ergo/config/*`

### Type Conversion Examples:
- `"hello"` → `string("hello")` (no prefix)
- `"int:123"` → `int64(123)`
- `"float:3.14"` → `float64(3.14)`
- `"bool:true"` → `bool(true)`, `"bool:false"` → `bool(false)`

### Configuration Demo Workflow:

```bash
# Setup sample configuration values
make setup-config

# Test configuration retrieval
make test-config

# View all configuration in etcd
make show-config

# Demonstrate real-time updates
make update-config

# Start nodes and watch them load configuration
make logs-nodes
```

## Actor Event Handling

### myActor Events
- **Configuration Changes**: Reacts to `myapp` and node-specific configuration
- **Database Settings**: Monitors `database.host` and `database.port` updates
- **Cache Settings**: Responds to `cache.size`, `cache.ratio`, and `cache.enabled` changes
- **Heartbeat Config**: Adjusts to `heartbeat.interval` modifications
- **Security Config**: Monitors `ssl.enabled` and other boolean settings

### clientActor Events
- **Resolver Settings**: Monitors `resolver.timeout` and `resolver.max_retries`
- **Connection Management**: Auto-reconnects when target nodes go down
- **Application Lifecycle**: Responds to `myapp` start/stop events
- **Node Discovery**: Attempts reconnection when new nodes join
- **Feature Flags**: Monitors `compression.enabled` and other boolean settings

## Usage

### Start Services
```bash
make up
```

### Configuration Management
```bash
make setup-config  # Setup sample configuration
make test-config   # Validate configuration storage
make show-config   # View all configuration
make update-config # Demonstrate real-time updates
make config-demo   # Full configuration workflow
```

### View Logs
```bash
make logs          # All services
make logs-nodes    # Only Ergo nodes
make logs-etcd     # Only etcd
```

### Cleanup
```bash
make down    # Stop services
make clean   # Remove everything
```

## Event Types Monitored

The actors demonstrate monitoring of these registrar events:

### Configuration Events (`gen.MessageEventConfig`)
- **Scope**: Identifies the configuration level (node/app/cluster/global)
- **Key**: The configuration key that changed
- **Value**: New value with automatic type conversion
- **Real-time**: Events fire immediately when etcd values change

### Cluster Events
- **`gen.MessageEventNodeUp`**: Node joins cluster
- **`gen.MessageEventNodeDown`**: Node leaves cluster
- **`gen.MessageEventApplicationStarted`**: Application starts on a node
- **`gen.MessageEventApplicationStopped`**: Application stops on a node

## Manual Configuration

You can manually add configuration values to etcd:

```bash
# Add a string value (no prefix)
docker-compose exec etcd etcdctl put \
    "services/ergo/cluster/docker-example/config/myapp/database.host" "localhost"

# Add an integer value (int: prefix)
docker-compose exec etcd etcdctl put \
    "services/ergo/cluster/docker-example/config/myapp/database.port" "int:5432"

# Add a float value (float: prefix)
docker-compose exec etcd etcdctl put \
    "services/ergo/cluster/docker-example/config/myapp/cache.ratio" "float:0.75"

# Add a boolean value (bool: prefix)
docker-compose exec etcd etcdctl put \
    "services/ergo/cluster/docker-example/config/myapp/cache.enabled" "bool:true"
```

Watch the node logs to see actors receive events immediately!

## Monitoring

Check service status:
```bash
make status
```

Check etcd health:
```bash
make check-etcd
```

## Expected Output

When running `make up`, you should see:

1. **Service Discovery**: Nodes discovering each other
2. **Event Monitoring Setup**: Actors starting to monitor registrar events
3. **Configuration Loading**: Nodes loading typed configuration values
4. **Actor Communication**: myactor actors sending heartbeats
5. **Service Resolution**: Node3 resolving and connecting to myapp services

When running `make update-config`, you'll see:

1. **Real-time Events**: Actors receiving configuration update events
2. **Type Conversion**: Values automatically converted (`int:256` → `int64(256)`, `bool:true` → `bool(true)`)
3. **Scope Detection**: Actors identifying which configs affect them
4. **Live Updates**: Current configuration displayed after each change

The colored logger will display:
- Ergo Framework ASCII logo on startup
- Timestamped colored log messages
- 🔍 Event monitoring setup confirmations
- ⚙️ Configuration update notifications with emojis
- 🎯 Scope-specific configuration targeting
- 🔄 Current configuration values after updates

## Configuration Values Loaded

The demo sets up the following configuration with **hierarchical precedence**:

**Global Config:**
- `log.level`: "info"
- `metrics.enabled`: bool(true)

**Cluster Config:**
- `timeout`: int64(30)
- `debug.enabled`: bool(false) 
- `max_connections`: int64(1000)

**Application Config (myapp):**
- `cache.size`: int64(128)
- `cache.ratio`: float64(0.75)
- `cache.enabled`: bool(true)
- `heartbeat.interval`: int64(10)
- `log.level`: "info"

**Application Config (client):**
- `cache.size`: int64(64)
- `ping.interval`: int64(15)
- `compression.enabled`: bool(false)
- `log.level`: "debug"

**Node-specific Config (highest priority):**
- Node1: `database.host`: "node1-db.local", `database.port`: int64(5432), `ssl.enabled`: bool(true), `log.level`: "info"
- Node2: `database.host`: "node2-db.local", `database.port`: int64(5433), `ssl.enabled`: bool(false), `log.level`: "debug"
- Node3: `resolver.timeout`: int64(5), `resolver.max_retries`: int64(3), `tls.verify`: bool(true), `log.level`: "warning"

## Real-time Update Demonstration

The `update-config.sh` script demonstrates live configuration changes:

1. **Cluster Settings**: timeout, debug mode (boolean)
2. **Application Settings**: cache sizes, heartbeat intervals, feature flags (boolean)
3. **Node Settings**: database hosts, worker threads, SSL/TLS settings (boolean)

Each change triggers **immediate events** to the affected actors!

