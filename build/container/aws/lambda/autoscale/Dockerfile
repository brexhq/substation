ARG ARCH=x86_64

FROM golang:1.22 as build
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# install AppConfig Lambda extension
COPY build/scripts/aws/lambda/extension.zip /tmp/extension.zip
RUN apt-get update && \
	apt-get install -y unzip && \
  	unzip /tmp/extension.zip -d /opt

# install deps
WORKDIR /usr/local/go/src/substation/
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

# copy src, unit test, build app
COPY . /usr/local/go/src/substation/
RUN go test -timeout 30s -v ./... 
WORKDIR /usr/local/go/src/substation/cmd/aws/lambda/autoscale
RUN go build -ldflags "-w -s" -o /var/task/main

FROM public.ecr.aws/lambda/provided:al2-$ARCH
COPY --from=build /opt /opt
COPY --from=build /var/task/main /var/task/main
ENTRYPOINT ["/var/task/main"]
