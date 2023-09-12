FROM golang:1.19 as build
ENV GO111MODULE=on

# install deps
WORKDIR /usr/local/go/src/substation/
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

# copy src, build proto, unit test, build app
COPY . /usr/local/go/src/substation/
RUN sh /usr/local/go/src/substation/build/scripts/proto/compile.sh
RUN go test -timeout 30s -v ./... 
WORKDIR /usr/local/go/src/substation/cmd/development/substation
RUN CGO_ENABLED=0 go build -o /bin/substation

FROM scratch
COPY --from=build /bin/substation /bin/substation
