FROM golang as build
ENV GO111MODULE=on

# install deps
WORKDIR /usr/local/go/src/substation/
COPY ./go.mod .
COPY ./go.sum .
RUN apt-get update && \
	apt-get install -y protobuf-compiler && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
	go mod download

# copy src, build proto, unit test, build app
COPY . /usr/local/go/src/substation/
RUN sh /usr/local/go/src/substation/build/scripts/proto/compile.sh
RUN go test -timeout 30s -v ./... 
WORKDIR /usr/local/go/src/substation/cmd/file/substation
RUN CGO_ENABLED=0 go build -o /bin/substation

FROM scratch
COPY --from=build /bin/substation /bin/substation
