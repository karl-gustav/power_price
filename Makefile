GPC_PROJECT_ID=my-cloud-collection
SERVICE_NAME=power-price
CONTAINER_NAME=eu.gcr.io/$(GPC_PROJECT_ID)/$(SERVICE_NAME)

run: build
	docker run -p 8080:8080 $(CONTAINER_NAME)
build:
	docker build -t $(CONTAINER_NAME) .
push: build
	docker push $(CONTAINER_NAME)
deploy: login-lpass push
	gcloud beta run deploy $(SERVICE_NAME)\
		--project $(GPC_PROJECT_ID)\
		--allow-unauthenticated\
		-q\
		--region europe-west1\
		--platform managed\
		--set-env-vars SECURITY_TOKEN=$$(lpass show entsoe.eu --field=web-api-security-token)\
		--memory 128Mi\
		--image $(CONTAINER_NAME)
use-latest-version:
	gcloud alpha run services update-traffic $(SERVICE_NAME)\
		--to-latest\
		--project $(GPC_PROJECT_ID)\
		--region europe-west1\
		--platform managed
login-lpass:
	lpass sync
test:
	go test ./...
