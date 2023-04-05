ARG AWS_ARCHITECTURE=x86_64

FROM golang:1.19 as build
ENV GO111MODULE=on

# install AppConfig Lambda extension
COPY build/scripts/aws/lambda/extension.zip /tmp/extension.zip
RUN apt-get update && \
	apt-get install -y unzip && \
  	unzip /tmp/extension.zip -d /opt

# install deps
WORKDIR /usr/local/go/src/substation/
COPY ./go.mod .
COPY ./go.sum .
RUN apt-get install -y protobuf-compiler && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
	go mod download

# copy src, build proto, unit test, build app
COPY . /usr/local/go/src/substation/
RUN sh /usr/local/go/src/substation/build/scripts/proto/compile.sh
RUN go test -timeout 30s -v ./... 
WORKDIR /usr/local/go/src/substation/cmd/aws/lambda/validation
RUN go build -o /var/task/main

FROM public.ecr.aws/lambda/provided:al2-$AWS_ARCHITECTURE
COPY --from=build /opt /opt
COPY --from=build /var/task/main /var/task/main
ENTRYPOINT ["/var/task/main"]
