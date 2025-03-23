export DOCKER_BUILDKIT=1
GCP_PROJECT_ID=my-cloud-collection
SERVICE_NAME=power-price
CONTAINER_NAME=europe-west1-docker.pkg.dev/$(GCP_PROJECT_ID)/cloud-run/$(SERVICE_NAME)

PORT?=8080

.PHONY: *

run: signin
	@if ! gcloud auth application-default print-access-token >/dev/null 2>&1; then echo 'Login to `gcloud` using the command\n\n\033[0;34mgcloud auth application-default login\033[0m\n' && exit 1; fi
	SECURITY_TOKEN=$$(op item get entsoe.eu --fields "Web Api Security Token") \
	PORT=$(PORT) go run .
build: test
	docker build -t $(CONTAINER_NAME) .
push: build
	docker push $(CONTAINER_NAME)
deploy: signin push
	gcloud run deploy $(SERVICE_NAME)\
		--project $(GCP_PROJECT_ID)\
		--allow-unauthenticated\
		-q\
		--region europe-west1\
		--platform managed\
		--set-env-vars SECURITY_TOKEN=$$(op item get entsoe.eu --fields "Web Api Security Token")\
		--memory 128Mi\
		--image $(CONTAINER_NAME)
deploy-staging: signin push
	gcloud run deploy $(SERVICE_NAME)\
		--project $(GCP_PROJECT_ID)\
		--allow-unauthenticated\
		-q\
		--region europe-west1\
		--platform managed\
		--set-env-vars SECURITY_TOKEN=$$(op item get entsoe.eu --fields "Web Api Security Token")\
		--memory 128Mi\
		--no-traffic\
		--image $(CONTAINER_NAME)
use-latest:
	gcloud run services update-traffic $(SERVICE_NAME)\
		--to-latest\
		--project $(GCP_PROJECT_ID)\
		--region europe-west1\
		--platform managed
watch-test:
	find -type f -name "*.go" | entr go test ./...
signin:
	@op account get >/dev/null 2>&1 || (echo '❌ 1password is not signed in, run:\n\n\t\e[96m\e[1meval $$(op signin)\e[0m\n'; exit 1)
test:
	go test ./...
