#!/bin/bash

# Setup configuration script for Docker Ergo example
# This script populates etcd with sample configuration values demonstrating typed configuration

set -e

ETCD_ENDPOINTS=${ETCD_ENDPOINTS:-"localhost:2379"}
CLUSTER_NAME="docker-example"

echo "Setting up configuration in etcd..."
echo "Endpoints: $ETCD_ENDPOINTS"
echo "Cluster: $CLUSTER_NAME"

# Wait for etcd to be ready
echo "Waiting for etcd to be ready..."
until docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 endpoint health > /dev/null 2>&1; do
    echo "Waiting for etcd..."
    sleep 1
done
echo "etcd is ready!"

# Global configuration (lowest priority)
echo "Setting global configuration..."
docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/config/global/log.level" "info"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/config/global/metrics.enabled" "bool:true"

# Cluster-wide configuration (using wildcard '*' for all nodes in cluster)
echo "Setting cluster-wide configuration..."
docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/*/timeout" "int:30"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/*/debug.enabled" "bool:false"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/*/max_connections" "int:1000"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/*/log.level" "info"

# Application-specific configuration
echo "Setting application-specific configuration..."

# myapp configuration
docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/myapp/cache.size" "int:128"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/myapp/cache.ratio" "float:0.75"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/myapp/heartbeat.interval" "int:10"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/myapp/log.level" "info"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/myapp/cache.enabled" "bool:true"

# client configuration
docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/client/cache.size" "int:64"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/client/ping.interval" "int:15"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/client/retry.max" "int:3"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/client/log.level" "debug"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/client/compression.enabled" "bool:false"

# Node-specific configuration (highest priority)
# Note: Using static node names that match actual container names in Docker network
echo "Setting node-specific configuration..."

# Node1 configuration
docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/database.host" "node1-db.local"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/database.port" "int:5432"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/worker.threads" "int:4"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/log.level" "info"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/ssl.enabled" "bool:true"

# Node2 configuration
docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/database.host" "node2-db.local"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/database.port" "int:5433"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/worker.threads" "int:8"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/log.level" "debug"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/ssl.enabled" "bool:false"

# Node3 configuration
docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/resolver.timeout" "int:5"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/resolver.max_retries" "int:3"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/connection.pool.size" "int:10"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/log.level" "warning"

docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put \
    "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/tls.verify" "bool:true"

echo "Configuration setup completed!"
echo ""
echo "=== Configuration hierarchy demonstration ====="
echo "1. Node-specific config (highest priority): services/ergo/cluster/$CLUSTER_NAME/config/{node-name}/*"
echo "2. Application config: services/ergo/cluster/$CLUSTER_NAME/config/{app-name}/*"
echo "3. Cluster config (wildcard): services/ergo/cluster/$CLUSTER_NAME/config/*/{item}"
echo "4. Global config (lowest priority): services/ergo/config/global/*"
echo ""
echo "=== Type conversion examples ==="
echo "String values: 'node1-db.local' → string"
echo "Int values: 'int:5432' → int64(5432)"
echo "Float values: 'float:0.75' → float64(0.75)"
echo "Boolean values: 'bool:true' → bool(true), 'bool:false' → bool(false)"
echo ""
echo "=== Log level hierarchy ==="
echo "Node1: info (node-specific)"
echo "Node2: debug (node-specific)"  
echo "Node3: warning (node-specific)"
echo "Apps use application-level log settings unless overridden"
echo ""
echo "Run 'make logs-nodes' to see configuration being loaded by the nodes!"
echo "Run 'make update-config' to demonstrate real-time configuration updates!" 