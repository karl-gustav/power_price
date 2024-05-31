GPC_PROJECT_ID=my-cloud-collection
SERVICE_NAME=power-price
CONTAINER_NAME=eu.gcr.io/$(GPC_PROJECT_ID)/$(SERVICE_NAME)

ENTSOE_SECRET_KEY=entsoe-web-api-security-token

run: build
	docker run -p 8080:8080 $(CONTAINER_NAME)
build:
	docker build -t $(CONTAINER_NAME) .
push: build
	docker push $(CONTAINER_NAME)
deploy: push
	gcloud beta run deploy $(SERVICE_NAME)\
		--project $(GPC_PROJECT_ID)\
		--allow-unauthenticated\
		-q\
		--region europe-west1\
		--platform managed\
		--memory 128Mi\
		--image $(CONTAINER_NAME)
upload-secrets:
	gcloud secrets create $(ENTSOE_SECRET_KEY)\
		--replication-policy="automatic"\
		--project $(GPC_PROJECT_ID)
	lpass show entsoe.eu --field=web-api-security-token | gcloud secrets versions add $(ENTSOE_SECRET_KEY)\
		--data-file=-\
		--project $(GPC_PROJECT_ID)
remove-secrets:
	gcloud secrets delete $(ENTSOE_SECRET_KEY)\
		--project $(GPC_PROJECT_ID)
use-latest-version:
	gcloud alpha run services update-traffic $(SERVICE_NAME)\
		--to-latest\
		--project $(GPC_PROJECT_ID)\
		--region europe-west1\
		--platform managed
test:
	go test ./...
