# Unique tag per build so every deploy triggers a rollout (also for uncommitted changes).
REGISTRY ?= ghcr.io/raulmoura157
TAG      ?= $(shell date +%Y%m%d%H%M%S)
TAG      := $(TAG)
IMAGE    ?= $(REGISTRY)/calculator-api:$(TAG)
DEFAULT_IMAGE = ghcr.io/raulmoura157/calculator-api:latest
KIND_CLUSTER ?= kind

.PHONY: test cover build deploy kind-deploy apply smoke undeploy

test:
	go test -cover ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

build:
	docker build -t $(IMAGE) .

# Build, push to REGISTRY and deploy (any cluster).
deploy: build
	docker push $(IMAGE)
	$(MAKE) apply IMAGE=$(IMAGE)

# Build and load the image straight into a kind cluster (no registry needed).
kind-deploy: build
	kind load docker-image $(IMAGE) --name $(KIND_CLUSTER)
	$(MAKE) apply IMAGE=$(IMAGE)

apply:
	sed 's|$(DEFAULT_IMAGE)|$(IMAGE)|' k8s/calculator-api.yaml | kubectl apply -f -
	kubectl -n calculator rollout status deployment/calculator-api --timeout=120s

# Calls the API from a throwaway pod inside the cluster.
smoke:
	kubectl -n calculator run smoke-$(TAG) --rm -i --restart=Never --image=curlimages/curl:8.11.1 -- \
		sh -c 'for op in sum sub mul div; do curl -s "http://calculator-api:8000/api/$$op?term_one=8&term_two=2"; done; curl -s http://calculator-api:8000/healthz'

undeploy:
	kubectl delete -f k8s/calculator-api.yaml --ignore-not-found
