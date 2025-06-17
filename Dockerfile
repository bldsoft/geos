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
ARG VERSION_FILE=VERSION
RUN go build -o geos -ldflags="\
    -X github.com/bldsoft/gost/version.GitCommit=${COMMIT} \
    -X github.com/bldsoft/gost/version.GitBranch=${BRANCH} \
    -X github.com/bldsoft/gost/version.Version=$(cat ${VERSION_FILE})" cmd/geos/main.go

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates wget

ARG GEONAME_DUMP_DIR=/etc/geos/geonames
RUN mkdir -p ${GEONAME_DUMP_DIR}
RUN wget -P ${GEONAME_DUMP_DIR} https://download.geonames.org/export/dump/admin1CodesASCII.txt && \
    wget -P ${GEONAME_DUMP_DIR} https://download.geonames.org/export/dump/cities500.zip && \
    wget -P ${GEONAME_DUMP_DIR} https://download.geonames.org/export/dump/countryInfo.txt

COPY --from=builder /go/src/${SRC_DIR}/geos /geos
EXPOSE 8505
EXPOSE 8506
ENTRYPOINT /geos