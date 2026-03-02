# Makefile for Observability example

.PHONY: help build up down logs clean restart status

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build node Docker images"
	@echo "  up       - Build and start all services (etcd, Prometheus, Grafana, nodes)"
	@echo "  down     - Stop all services"
	@echo "  logs     - Show logs from all services"
	@echo "  clean    - Remove all containers, images, and volumes"
	@echo "  restart  - Restart all services"
	@echo "  status   - Show status of all services"

# Build node images
build:
	@echo "Building node images..."
	@docker-compose build cluster-node1 cluster-node2 cluster-node3

# Start all services
up:
	@echo "Building and starting all services..."
	@docker-compose up -d --build
	@echo ""
	@echo "Services started:"
	@echo "  etcd:       http://localhost:2379"
	@echo "  Prometheus: http://localhost:9091"
	@echo "  Grafana:    http://localhost:8888 (admin/ergo)"
	@echo "  cluster-node1:  radar :9090, mcp :9922"
	@echo "  cluster-node2:  radar :9090"
	@echo "  cluster-node3:  radar :9090"
	@echo "  cluster-node4:  radar :9090"
	@echo "  cluster-node5:  radar :9090"
	@echo "  observer:       http://localhost:9911"
	@echo ""
	@echo "Use 'make logs' to see output."

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
