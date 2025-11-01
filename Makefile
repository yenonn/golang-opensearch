.PHONY: start stop delete status help deploy-opensearch deploy-dashboard

PROFILE_NAME := my-elasticsearch-cluster

help:
	@echo "Available targets:"
	@echo "  start             - Start minikube cluster with profile $(PROFILE_NAME) and deploy OpenSearch + Dashboards"
	@echo "  stop              - Stop the minikube cluster and remove deployments"
	@echo "  delete            - Delete the minikube cluster"
	@echo "  status            - Check the status of the minikube cluster"
	@echo "  deploy-opensearch - Deploy OpenSearch to the cluster"
	@echo "  deploy-dashboard  - Deploy OpenSearch Dashboards to the cluster"

start:
	minikube start --profile $(PROFILE_NAME)
	@echo "Deploying OpenSearch..."
	kubectl --context $(PROFILE_NAME) apply -f opensearch-deployment.yaml
	@echo "Waiting for OpenSearch to be ready..."
	kubectl --context $(PROFILE_NAME) wait --for=condition=available --timeout=300s deployment/opensearch
	@echo "OpenSearch is ready!"
	@echo "Deploying OpenSearch Dashboards..."
	kubectl --context $(PROFILE_NAME) apply -f opensearch-dashboard.yaml
	@echo "Waiting for OpenSearch Dashboards to be ready..."
	kubectl --context $(PROFILE_NAME) wait --for=condition=available --timeout=300s deployment/opensearch-dashboards
	@echo ""
	@echo "=== Deployment Complete ==="
	@echo "OpenSearch API:   http://$$(minikube --profile $(PROFILE_NAME) ip):30920"
	@echo "OpenSearch Dashboards: http://$$(minikube --profile $(PROFILE_NAME) ip):30561"
	@echo ""

deploy-opensearch:
	kubectl --context $(PROFILE_NAME) apply -f opensearch-deployment.yaml

deploy-dashboard:
	kubectl --context $(PROFILE_NAME) apply -f opensearch-dashboard.yaml

stop:
	@echo "Stopping OpenSearch Dashboards..."
	-kubectl --context $(PROFILE_NAME) delete -f opensearch-dashboard.yaml
	@echo "Stopping OpenSearch deployment..."
	-kubectl --context $(PROFILE_NAME) delete -f opensearch-deployment.yaml
	@echo "Stopping minikube cluster..."
	minikube stop --profile $(PROFILE_NAME)

delete:
	minikube delete --profile $(PROFILE_NAME)

status:
	minikube status --profile $(PROFILE_NAME)