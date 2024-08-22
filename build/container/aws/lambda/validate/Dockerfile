ARG ARCH=x86_64
ARG SOURCE="."  # default to local source

FROM golang:1.22 as build
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# install deps
WORKDIR /usr/local/go/src/substation/
COPY ${SOURCE}/go.mod .
COPY ${SOURCE}/go.sum .
RUN go mod download

# copy src, unit test, build app
COPY ${SOURCE} /usr/local/go/src/substation/
RUN go test -timeout 30s -v ./... 
WORKDIR /usr/local/go/src/substation/cmd/aws/lambda/validate
RUN go build -ldflags "-w -s" -o /var/task/main

FROM public.ecr.aws/lambda/provided:al2-$ARCH
COPY --from=build /opt /opt
COPY --from=build /var/task/main /var/task/main
ENTRYPOINT ["/var/task/main"]
