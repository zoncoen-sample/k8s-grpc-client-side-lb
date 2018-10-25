APPNAME = k8s-grpc-client-side-lb-server
APPVERSION = latest
HTTPAPP = $(APPNAME)-http
GRPCAPP = $(APPNAME)-grpc
HTTPIMAGE = gcr.io/$(PROJECT_ID)/$(HTTPAPP):$(APPVERSION)
GRPCIMAGE = gcr.io/$(PROJECT_ID)/$(GRPCAPP):$(APPVERSION)
HTTPPORT = 8000
GRPCPORT = 9000
LOADBALANCER = load-balancer-service
CLUSTERIP = cluster-ip-service
HEADLESS = headless-service
TIMESTAMP = $(shell date "+%s")

.PHONY: gen
gen: clean
	protoc --go_out=plugins=grpc:pb proto/*.proto

.PHONY: clean
clean:
	rm -r pb
	mkdir pb

.PHONY: build-image
build-image:
	docker build -t ${HTTPIMAGE} . --build-arg target=http
	docker build -t ${GRPCIMAGE} . --build-arg target=grpc

.PHONY: upload-image
upload-image: build-image
	gcloud docker -- push ${HTTPIMAGE}
	gcloud docker -- push ${GRPCIMAGE}

.PHONY: deploy-http
deploy-http:
	cat config/deployment.tmpl.yaml | sed s!{{appname}}!${HTTPAPP}! | sed s!{{image}}!${HTTPIMAGE}! | sed s!{{clusterip}}!${CLUSTERIP}! | sed s!{{headless}}!${HEADLESS}! | sed s!{{replicas}}!1! | sed s!{{httpport}}!${HTTPPORT}! | sed s!{{grpcport}}!${GRPCPORT}! | sed s!{{revision}}!${TIMESTAMP}! | kubectl apply -f -

.PHONY: deploy-grpc
deploy-grpc:
	cat config/deployment.tmpl.yaml | sed s!{{appname}}!${GRPCAPP}! | sed s!{{image}}!${GRPCIMAGE}! | sed s!{{clusterip}}!${CLUSTERIP}! | sed s!{{headless}}!${HEADLESS}! | sed s!{{replicas}}!3! | sed s!{{httpport}}!${HTTPPORT}! | sed s!{{grpcport}}!${GRPCPORT}! | sed s!{{revision}}!${TIMESTAMP}! | kubectl apply -f -

.PHONY: deploy-service
deploy-service:
	cat config/service.tmpl.yaml | sed s!{{httpapp}}!${HTTPAPP}! | sed s!{{grpcapp}}!${GRPCAPP}! | sed s!{{loadbalancer}}!${LOADBALANCER}! | sed s!{{clusterip}}!${CLUSTERIP}! | sed s!{{headless}}!${HEADLESS}! | sed s!{{httpport}}!${HTTPPORT}! | sed s!{{grpcport}}!${GRPCPORT}! | kubectl apply -f -

.PHONY: deploy
deploy: upload-image deploy-http deploy-grpc deploy-service

.PHONY: create-cluster
create-cluster:
	gcloud container clusters create ${APPNAME} --num-nodes=3

.PHONY: delete-cluster
delete-cluster:
	gcloud container clusters delete ${APPNAME}

.PHONY: external-ip
external-ip:
	@kubectl get service | grep ${LOADBALANCER} | awk '{print $$4}'

.PHONY: setup
setup:
	gcloud config set project ${PROJECT_ID}
