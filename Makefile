export DOCKER_BUILDKIT=1
GPC_PROJECT_ID=my-cloud-collection
SERVICE_NAME=power-price
CONTAINER_NAME=eu.gcr.io/$(GPC_PROJECT_ID)/$(SERVICE_NAME)

OP_SESSION=$(eval OP_SESSION := $$(shell op signin --raw --account my.1password.com))$(OP_SESSION)

run: signin
	GOOGLE_APPLICATION_CREDENTIALS=~/gcp/gcp_key.json \
	SECURITY_TOKEN=$$(op item get entsoe.eu --fields "Web Api Security Token" --session=$(OP_SESSION)) \
	go run .
build: test
	docker build -t $(CONTAINER_NAME) .
push: build
	docker push $(CONTAINER_NAME)
deploy: signin push
	gcloud run deploy $(SERVICE_NAME)\
		--project $(GPC_PROJECT_ID)\
		--allow-unauthenticated\
		-q\
		--region europe-west1\
		--platform managed\
		--set-env-vars SECURITY_TOKEN=$$(op item get entsoe.eu --fields "Web Api Security Token" --session=$(OP_SESSION))\
		--memory 128Mi\
		--image $(CONTAINER_NAME)
	# add --no-traffic to not use latest version
use-latest-version:
	gcloud run services update-traffic $(SERVICE_NAME)\
		--to-latest\
		--project $(GPC_PROJECT_ID)\
		--region europe-west1\
		--platform managed
signin:
	@test $(OP_SESSION) && echo ✓ Signin success || (echo ❌ 1password signin failed; exit 1)
test:
	go test ./...
