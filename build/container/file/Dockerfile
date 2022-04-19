FROM golang as build
ENV GO111MODULE=on

# install deps
WORKDIR /usr/local/go/src/substation/
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

# copy src, unit test, build
COPY . /usr/local/go/src/substation/
RUN go test -timeout 30s -v ./... 
WORKDIR /usr/local/go/src/substation/cmd/file/substation
RUN CGO_ENABLED=0 go build -o /bin/substation

FROM scratch
COPY --from=build /bin/substation /bin/substation
