.PHONY: build
SVC := wylis
COMMIT := $(shell git log -1 --pretty='%h')
REPOSITORY := 172.16.16.2:2443/go
PUBLIC_REPOSITIORY = icydoge/web

.PHONY: pull build push

all: pull build push clean
publish: pull build push-public

build:
	docker build -t ${SVC} .

pull:
	docker pull golang:alpine

push:
	docker tag ${SVC}:latest ${REPOSITORY}:${SVC}-${COMMIT}
	docker push ${REPOSITORY}:${SVC}-${COMMIT}

push-public:
		docker tag ${SVC}:latest ${PUBLIC_REPOSITIORY}:${SVC}
		docker tag ${SVC}:latest ${PUBLIC_REPOSITIORY}:${SVC}-${COMMIT}
		docker push ${PUBLIC_REPOSITIORY}:${SVC}
		docker push ${PUBLIC_REPOSITIORY}:${SVC}-${COMMIT}

clean:
	docker image prune -f