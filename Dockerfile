ARG SRC_DIR=github.com/bldsoft/geos
#build stage
FROM golang:1.21-alpine AS builder
WORKDIR /go/src/${SRC_DIR}

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
ARG COMMIT
ARG BRANCH
ARG VERSION_FILE=/cmd/geos/VERSION
RUN go build -o geos -ldflags="\
    -X github.com/bldsoft/gost/version.GitCommit=${COMMIT} \
    -X github.com/bldsoft/gost/version.GitBranch=${BRANCH} \
    -X github.com/bldsoft/gost/version.Version=$(cat ${VERSION_FILE})" cmd/geos/main.go

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/${SRC_DIR}/geos /geos
EXPOSE 8505
EXPOSE 8506
ENTRYPOINT /geos