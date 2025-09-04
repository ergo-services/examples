#!/bin/bash

# Real-time configuration update script for Docker Ergo example
# This script demonstrates dynamic configuration changes that trigger events

set -e

CLUSTER_NAME="docker-example"

echo "🔄 Starting real-time configuration updates with random values..."
echo ""

# Function to generate random values
random_int() {
    local min=$1
    local max=$2
    echo $((RANDOM % (max - min + 1) + min))
}

random_float() {
    local min=$1
    local max=$2
    echo "scale=2; $min + ($max - $min) * $RANDOM / 32768" | bc -l
}

random_bool() {
    if [ $((RANDOM % 2)) -eq 0 ]; then
        echo "true"
    else
        echo "false"
    fi
}

random_host() {
    local prefixes=("db" "cache" "api" "storage" "backup")
    local suffixes=("local" "prod" "dev" "test" "stage")
    local prefix=${prefixes[$((RANDOM % ${#prefixes[@]}))]}
    local suffix=${suffixes[$((RANDOM % ${#suffixes[@]}))]}
    echo "$prefix-$suffix.example.com"
}

# Function to update configuration
update_config() {
    local key="$1"
    local value="$2"
    local description="$3"
    
    echo "⚙️  Updating $description..."
    docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 put "$key" "$value"
    echo "   ✅ Set $key = $value"
    sleep 2
}

# Wait for etcd to be ready
echo "⏳ Waiting for etcd to be ready..."
until docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 endpoint health > /dev/null 2>&1; do
    echo "   Waiting..."
    sleep 1
done
echo "✅ etcd is ready!"
echo ""

echo "🎭 Demonstrating real-time configuration updates with random values..."
echo "   Watch the node logs with: make logs-nodes"
echo ""

# Generate random values
TIMEOUT=$(random_int 30 120)
DEBUG_ENABLED=$(random_bool)
CACHE_SIZE=$(random_int 64 512)
CACHE_RATIO=$(random_float 0.5 0.9)
HEARTBEAT_INTERVAL=$(random_int 5 30)
CACHE_ENABLED=$(random_bool)
PING_INTERVAL=$(random_int 10 60)
CLIENT_CACHE_SIZE=$(random_int 32 256)
COMPRESSION_ENABLED=$(random_bool)

# Node-specific random values
NODE1_DB_HOST=$(random_host)
NODE1_DB_PORT=$(random_int 5000 6000)
NODE1_SSL_ENABLED=$(random_bool)

NODE2_DB_HOST=$(random_host)
NODE2_WORKER_THREADS=$(random_int 4 32)
NODE2_SSL_ENABLED=$(random_bool)

NODE3_RESOLVER_TIMEOUT=$(random_int 3 15)
NODE3_MAX_RETRIES=$(random_int 2 10)
NODE3_TLS_VERIFY=$(random_bool)

echo "🎲 Generated random configuration values:"
echo "   • Cluster timeout: ${TIMEOUT}s"
echo "   • Debug mode: $DEBUG_ENABLED"
echo "   • Cache size: ${CACHE_SIZE}MB"
echo "   • Cache ratio: ${CACHE_RATIO}"
echo "   • Node1 DB: $NODE1_DB_HOST:$NODE1_DB_PORT"
echo "   • Node2 workers: $NODE2_WORKER_THREADS"
echo "   • Node3 timeout: ${NODE3_RESOLVER_TIMEOUT}s"
echo ""

# Update cluster-wide settings
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/timeout" "int:$TIMEOUT" "cluster timeout"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/debug.enabled" "bool:$DEBUG_ENABLED" "cluster debug mode"

# Update application settings (using wildcard format so all nodes can see them)
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/myapp.cache.size" "int:$CACHE_SIZE" "myapp cache size"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/myapp.cache.ratio" "float:$CACHE_RATIO" "myapp cache ratio"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/myapp.heartbeat.interval" "int:$HEARTBEAT_INTERVAL" "myapp heartbeat interval"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/myapp.cache.enabled" "bool:$CACHE_ENABLED" "myapp cache enabled"

# Update client application settings (using wildcard format so all nodes can see them)
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/client.ping.interval" "int:$PING_INTERVAL" "client ping interval"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/client.cache.size" "int:$CLIENT_CACHE_SIZE" "client cache size"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/*/client.compression.enabled" "bool:$COMPRESSION_ENABLED" "client compression enabled"

# Update node-specific settings
update_config "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/database.host" "$NODE1_DB_HOST" "node1 database host"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/database.port" "int:$NODE1_DB_PORT" "node1 database port"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/ssl.enabled" "bool:$NODE1_SSL_ENABLED" "node1 ssl enabled"

update_config "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/database.host" "$NODE2_DB_HOST" "node2 database host"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/worker.threads" "int:$NODE2_WORKER_THREADS" "node2 worker threads"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/ssl.enabled" "bool:$NODE2_SSL_ENABLED" "node2 ssl enabled"

update_config "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/resolver.timeout" "int:$NODE3_RESOLVER_TIMEOUT" "node3 resolver timeout"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/resolver.max_retries" "int:$NODE3_MAX_RETRIES" "node3 max retries"
update_config "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/tls.verify" "bool:$NODE3_TLS_VERIFY" "node3 tls verify"

echo ""
echo "🏁 Configuration update demonstration completed!"
echo ""
echo "💡 The actors should have received real-time events for each configuration change"
echo "   Check the logs to see the event handling in action!"
echo ""
echo "🎲 Each run generates new random values - run again to see different configurations!"
echo "🔄 To run more updates, execute this script again: ./update-config.sh" 