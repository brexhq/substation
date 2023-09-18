ARG AWS_ARCHITECTURE=x86_64

FROM golang:1.21 as build
ENV GO111MODULE=on

# install deps
WORKDIR /usr/local/go/src/substation/
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

# copy src, unit test, build app
COPY . /usr/local/go/src/substation/
RUN go test -timeout 30s -v ./... 
WORKDIR /usr/local/go/src/substation/cmd/aws/lambda/validation
RUN go build -o /var/task/main

FROM public.ecr.aws/lambda/provided:al2-$AWS_ARCHITECTURE
COPY --from=build /opt /opt
COPY --from=build /var/task/main /var/task/main
ENTRYPOINT ["/var/task/main"]
