FROM golang:1.10-alpine as builder
WORKDIR /go/src/amdgpu-metrics
COPY . .
RUN \ 
	apk add git && \
	go get -d -v ./... && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -ldflags="-w -s" -v ./...

FROM ubuntu:bionic
COPY --from=builder /go/bin/amdgpu-metrics /
WORKDIR /
CMD ["/amdgpu-metrics"]
