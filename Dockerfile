FROM golang:1.16-alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o redalert .

FROM golang:1.16-alpine

WORKDIR /

COPY --from=builder /build/redalert /usr/local/bin/redalert
RUN chmod +x /usr/local/bin/redalert

EXPOSE 8888

ENTRYPOINT ["redalert", "server"]