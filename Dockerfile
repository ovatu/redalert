FROM golang:1.16-alpine as builder
RUN apk add --update npm

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN cd ui && npm install && NODE_ENV=production ./node_modules/.bin/webpack -p && cd ..
RUN mkdir -p web/assets
RUN cp ui/dist/assets/app.bundle.js web/assets/
RUN cp ui/index.html web/assets

RUN go get github.com/GeertJohan/go.rice
RUN go get github.com/GeertJohan/go.rice/rice
RUN cd web && rice embed-go && cd ..

RUN go build -o redalert .

FROM golang:1.16-alpine

WORKDIR /

COPY --from=builder /build/redalert /usr/local/bin/redalert
RUN chmod +x /usr/local/bin/redalert

EXPOSE 8888

ENTRYPOINT ["redalert", "server"]