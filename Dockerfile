# using a multi-stage build to reduce image size:
# https://docs.docker.com/develop/develop-images/multistage-build/
FROM golang:1.13-alpine AS build

# build outside of GOPATH (simpler when using Go modules)
WORKDIR /src

# copy dependencies
COPY vendor ./vendor
COPY go.mod go.sum ./

# copy everything else
COPY . .

# build application
RUN go install -mod=vendor

# create working image
FROM alpine

# for c++11 program service
RUN apk add g++

# copy application
COPY --from=build /go/bin/fd8-judge /usr/local/bin/

# confirm application works
RUN fd8-judge | grep 'fd8-judge is an open source cloud-native online judge.'
