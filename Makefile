.PHONY: start stop delete status help deploy-opensearch deploy-dashboard run build get-minikube-ip test clean port-forward tunnel

PROFILE_NAME := my-elasticsearch-cluster
BINARY_NAME := opensearch-app
BUILD_DIR := bin

help:
	@echo "Available targets:"
	@echo ""
	@echo "Cluster Management:"
	@echo "  start             - Start minikube cluster with profile $(PROFILE_NAME) and deploy OpenSearch + Dashboards"
	@echo "  stop              - Stop the minikube cluster and remove deployments"
	@echo "  delete            - Delete the minikube cluster"
	@echo "  status            - Check the status of the minikube cluster"
	@echo "  deploy-opensearch - Deploy OpenSearch to the cluster"
	@echo "  deploy-dashboard  - Deploy OpenSearch Dashboards to the cluster"
	@echo ""
	@echo "Application:"
	@echo "  run               - Run the Go application"
	@echo "  build             - Build the application binary"
	@echo "  test              - Run tests"
	@echo "  clean             - Remove build artifacts"
	@echo ""
	@echo "Utilities:"
	@echo "  get-minikube-ip   - Get the minikube IP address"
	@echo "  port-forward      - Set up port forwarding to OpenSearch (use with Docker driver)"
	@echo "  tunnel            - Create minikube service tunnel (alternative for Docker driver)"

start:
	minikube start --profile $(PROFILE_NAME)
	@echo "Deploying OpenSearch..."
	kubectl --context $(PROFILE_NAME) apply -f build/opensearch-deployment.yaml
	@echo "Waiting for OpenSearch to be ready..."
	kubectl --context $(PROFILE_NAME) wait --for=condition=available --timeout=300s deployment/opensearch
	@echo "OpenSearch is ready!"
	@echo "Deploying OpenSearch Dashboards..."
	kubectl --context $(PROFILE_NAME) apply -f build/opensearch-dashboard.yaml
	@echo "Waiting for OpenSearch Dashboards to be ready..."
	kubectl --context $(PROFILE_NAME) wait --for=condition=available --timeout=300s deployment/opensearch-dashboards
	@echo ""
	@echo "=== Deployment Complete ==="
	@echo "OpenSearch API:   http://$$(minikube --profile $(PROFILE_NAME) ip):30920"
	@echo "OpenSearch Dashboards: http://$$(minikube --profile $(PROFILE_NAME) ip):30561"
	@echo ""

deploy-opensearch:
	kubectl --context $(PROFILE_NAME) apply -f build/opensearch-deployment.yaml

deploy-dashboard:
	kubectl --context $(PROFILE_NAME) apply -f build/opensearch-dashboard.yaml

stop:
	@echo "Stopping OpenSearch Dashboards..."
	-kubectl --context $(PROFILE_NAME) delete -f build/opensearch-dashboard.yaml
	@echo "Stopping OpenSearch deployment..."
	-kubectl --context $(PROFILE_NAME) delete -f build/opensearch-deployment.yaml
	@echo "Stopping minikube cluster..."
	minikube stop --profile $(PROFILE_NAME)

delete:
	minikube delete --profile $(PROFILE_NAME)

status:
	minikube status --profile $(PROFILE_NAME)

# Application targets
run:
	@echo "Getting Minikube IP..."
	@MINIKUBE_IP=$$(minikube ip --profile $(PROFILE_NAME) 2>/dev/null || echo "192.168.49.2") && \
	echo "Using Minikube IP: $$MINIKUBE_IP" && \
	MINIKUBE_IP=$$MINIKUBE_IP go run main.go

build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary created at: $(BUILD_DIR)/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Utilities
get-minikube-ip:
	@minikube ip --profile $(PROFILE_NAME)

port-forward:
	@echo "Setting up port forwarding to OpenSearch..."
	@echo "OpenSearch will be available at: http://localhost:9200"
	@echo "Press Ctrl+C to stop port forwarding"
	kubectl --context $(PROFILE_NAME) port-forward svc/opensearch 9200:9200

tunnel:
	@echo "Creating minikube service tunnel..."
	@echo "This requires keeping the terminal open"
	@echo "Press Ctrl+C to stop the tunnel"
	minikube service opensearch --profile $(PROFILE_NAME)