#!/bin/bash

# Test configuration script for Docker Ergo example
# This script tests that configuration values are correctly stored and can be retrieved

set -e

CLUSTER_NAME="docker-example"

echo "=== Testing Configuration in etcd ==="
echo ""

# Function to test a configuration key
test_config_key() {
    local key="$1"
    local expected_value="$2"
    local description="$3"
    
    echo -n "Testing $description... "
    
    actual_value=$(docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 get "$key" --print-value-only 2>/dev/null || echo "")
    
    if [ "$actual_value" = "$expected_value" ]; then
        echo "✅ PASS (value: $actual_value)"
    else
        echo "❌ FAIL (expected: $expected_value, got: $actual_value)"
    fi
}

# Test global configuration
echo "📋 Global Configuration:"
test_config_key "services/ergo/config/log.level" "info" "global log level"
test_config_key "services/ergo/config/metrics.enabled" "bool:true" "global metrics"

echo ""

# Test cluster configuration
echo "🌐 Cluster Configuration:"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/timeout" "int:30" "cluster timeout"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/debug.enabled" "bool:false" "cluster debug"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/max_connections" "int:1000" "cluster max connections"

echo ""

# Test application configuration
echo "📱 Application Configuration:"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/myapp/cache.size" "int:128" "myapp cache size"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/myapp/cache.ratio" "float:0.75" "myapp cache ratio"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/myapp/cache.enabled" "bool:true" "myapp cache enabled"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/client/cache.size" "int:64" "client cache size"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/client/compression.enabled" "bool:false" "client compression enabled"

echo ""

# Test node-specific configuration
echo "🖥️ Node-specific Configuration:"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/database.host" "node1-db.local" "node1 database host"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/database.port" "int:5432" "node1 database port"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/node1@node1/ssl.enabled" "bool:true" "node1 ssl enabled"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/database.host" "node2-db.local" "node2 database host"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/node2@node2/ssl.enabled" "bool:false" "node2 ssl enabled"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/resolver.timeout" "int:5" "node3 resolver timeout"
test_config_key "services/ergo/cluster/$CLUSTER_NAME/config/node3@node3/tls.verify" "bool:true" "node3 tls verify"

echo ""

# Count total configuration entries
echo "📊 Configuration Statistics:"
total_config=$(docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 get --prefix "services/ergo/" --keys-only | wc -l || echo "0")
global_config=$(docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 get --prefix "services/ergo/config/" --keys-only | wc -l || echo "0")
cluster_config=$(docker-compose exec -T etcd etcdctl --endpoints=http://localhost:2379 get --prefix "services/ergo/cluster/$CLUSTER_NAME/config/" --keys-only | wc -l || echo "0")

echo "  Total configuration entries: $total_config"
echo "  Global configuration entries: $global_config"
echo "  Cluster configuration entries: $cluster_config"

echo ""

# Test type prefix examples
echo "🏷️ Type Conversion Examples:"
echo "  String: 'node1-db.local' → string"
echo "  Integer: 'int:5432' → int64(5432)"
echo "  Float: 'float:0.75' → float64(0.75)"
echo "  Boolean: 'bool:true' → bool(true), 'bool:false' → bool(false)"

echo ""
echo "✨ Configuration test completed!"
echo ""
echo "Next steps:"
echo "  1. Run 'make logs-nodes' to see configuration being loaded by nodes"
echo "  2. Run 'make show-config' to view all configuration in etcd"
echo "  3. Check node logs for '=== Configuration Demo ===' sections"
echo "  4. Run 'make update-config' to see real-time configuration events" 