IMAGE=trace-context-injector

clean:
	@echo 'Removing generated configurations and local image build...'
	rm deploy/mutatingwebhook-ca-bundle.yaml
	docker rmi -f $(IMAGE)

# docker commands
docker:
	@echo 'Building image $(IMAGE)...'
	docker build -t $(IMAGE) .

# cluster commands
cluster-up: deploy-certs deploy-config

cluster-down: delete-config delete-certs

deploy-certs:
	@echo 'Generating certs and deploying into active cluster...'
	hack/webhook-create-signed-cert.sh --service trace-context-injector-webhook-svc --secret trace-context-injector-webhook-certs --namespace default
	cat deploy/mutatingwebhook.yaml | hack/webhook-patch-ca-bundle.sh > deploy/mutatingwebhook-ca-bundle.yaml

delete-certs:
	@echo 'Deleting mutating controller certs...'
	kubectl delete secret trace-context-injector-webhook-certs

deploy-config:
	@echo 'Applying configuration to active cluster...'
	kubectl apply -f deploy/deployment.yaml
	kubectl apply -f deploy/mutatingwebhook-ca-bundle.yaml
	kubectl apply -f deploy/service.yaml

delete-config:
	@echo 'Tearing down mutating controller and associated resources...'
	kubectl delete -f deploy/deployment.yaml
	kubectl delete -f deploy/mutatingwebhook-ca-bundle.yaml
	kubectl delete -f deploy/service.yaml