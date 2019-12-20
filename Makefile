.PHONY: build
SVC := wylis
COMMIT := $(shell git log -1 --pretty='%h')
REPOSITORY := 172.16.16.2:2443/go

.PHONY: pull build push

all: pull build push clean

build:
	docker build -t ${SVC} .

pull:
	docker pull golang:alpine

push:
	docker tag ${SVC}:latest ${REPOSITORY}:${SVC}-${COMMIT}
	docker push ${REPOSITORY}:${SVC}-${COMMIT}

clean:
	docker image prune -f