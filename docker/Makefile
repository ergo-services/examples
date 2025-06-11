# Makefile for Docker Ergo example

.PHONY: help build up down logs clean restart status setup-config show-config test-config update-config

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build all Docker images"
	@echo "  up            - Start all services"
	@echo "  down          - Stop all services"
	@echo "  logs          - Show logs from all services"
	@echo "  clean         - Remove all containers, images, and volumes"
	@echo "  restart       - Restart all services"
	@echo "  status        - Show status of all services"
	@echo "  setup-config  - Setup sample configuration in etcd"
	@echo "  show-config   - Show all configuration values in etcd"
	@echo "  test-config   - Test configuration values in etcd"
	@echo "  update-config - Demonstrate real-time configuration updates"

# Build all Docker images
build:
	@echo "Building Docker images..."
	@docker-compose build

# Start all services
up:
	@echo "Starting all services..."
	@docker-compose up -d
	@echo "Services started. Use 'make logs' to see output."

# Stop all services
down:
	@echo "Stopping all services..."
	@docker-compose down

# Show logs from all services
logs:
	@docker-compose logs -f

# Clean up everything
clean:
	@echo "Cleaning up..."
	@docker-compose down -v --rmi all
	@docker system prune -f

# Restart all services
restart: down up

# Show status of all services
status:
	@docker-compose ps
	@echo ""
	@echo "etcd status:"
	@docker-compose exec etcd etcdctl --endpoints=http://localhost:2379 endpoint health || echo "etcd not accessible"

# Show only node logs
logs-nodes:
	@docker-compose logs -f node1 node2 node3

# Show only etcd logs
logs-etcd:
	@docker-compose logs -f etcd

# Check etcd cluster
check-etcd:
	@echo "Checking etcd cluster..."
	@docker-compose exec etcd etcdctl --endpoints=http://localhost:2379 endpoint health
	@echo "Listing all keys:"
	@docker-compose exec etcd etcdctl --endpoints=http://localhost:2379 get --prefix "services/ergo/"

# Show all configuration values in etcd  
show-config:
	@echo "=== All Configuration Values in etcd ==="
	@echo ""
	@echo "📋 Global Configuration:"
	@docker-compose exec etcd etcdctl --endpoints=http://localhost:2379 get --prefix "services/ergo/config/" || echo "No global config found"
	@echo ""
	@echo "🌐 Cluster Configuration:"
	@docker-compose exec etcd etcdctl --endpoints=http://localhost:2379 get --prefix "services/ergo/cluster/docker-example/config/" || echo "No cluster config found"
	@echo ""
	@echo "📊 Configuration Statistics:"
	@echo -n "  Total entries: "
	@docker-compose exec etcd etcdctl --endpoints=http://localhost:2379 get --prefix "services/ergo/" --keys-only | wc -l || echo "0"

# Setup sample configuration in etcd
setup-config:
	@echo "Setting up sample configuration..."
	@chmod +x setup-config.sh
	@./setup-config.sh

# Test configuration values in etcd
test-config:
	@echo "Testing configuration values in etcd..."
	@chmod +x test-config.sh
	@./test-config.sh

# Demonstrate real-time configuration updates
update-config:
	@echo "Demonstrating real-time configuration updates..."
	@chmod +x update-config.sh
	@./update-config.sh 