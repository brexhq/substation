ARG SOURCE="."  # default to local source

FROM --platform=$BUILDPLATFORM golang:1.23 AS build
ENV GO111MODULE=on 
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# install deps
WORKDIR /usr/local/go/src/substation/
COPY ${SOURCE}/go.mod .
COPY ${SOURCE}/go.sum .
RUN go mod download

# copy src, unit test, build app
COPY ${SOURCE} /usr/local/go/src/substation/
RUN go test -timeout 30s -v ./... 
WORKDIR /usr/local/go/src/substation/cmd/gcp/function/substation
RUN go build -o /var/task/main

FROM gcr.io/distroless/static:nonroot
COPY --from=build /var/task/main /var/task/main

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/var/task/main"]
