BINARY=redalert

VERSION=0.2.4
COMMIT=`git rev-parse HEAD`

LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT}"

install-deps:
	glide install

build: embed-static
	go build ${LDFLAGS} -o ${BINARY}

embed-static: build-ui
	go get github.com/GeertJohan/go.rice
	go get github.com/GeertJohan/go.rice/rice
	cd web && rice embed-go && cd ..

build-ui:
	cd ui && npm install && NODE_ENV=production ./node_modules/.bin/webpack -p && cd ..
	mkdir -p web/assets
	cp ui/dist/assets/app.bundle.js web/assets/
	cp ui/index.html web/assets

run-dev-ui:
	cd ui && npm install && ./node_modules/.bin/webpack-dev-server

build-proto:
	protoc -I servicepb/ servicepb/service.proto --go_out=plugins=grpc:servicepb

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

test-deps:
	docker pull sickp/alpine-sshd
	docker pull postgres

test:
	go test -v -race $(shell glide novendor)

build-docker-image-local: embed-static
	docker run --rm \
		-v "$(shell pwd):/src" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		centurylink/golang-builder \
		ovatu/redalert

build-docker-image-remote: build-docker-image-local
	docker tag ovatu/redalert ovatu/redalert:v${VERSION}
	docker push ovatu/redalert

.PHONY: embed-static build-ui build-proto clean test-deps test build-docker-image build-docker-image-remote
