FROM mcr.microsoft.com/vscode/devcontainers/go:1.19
RUN apt-get update -y && \
    # install Terraform
    cd /tmp/ && \
    curl https://releases.hashicorp.com/terraform/1.1.7/terraform_1.1.7_linux_$(dpkg --print-architecture).zip -o ./terraform.zip && \
    unzip terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    # install Jsonnet
    go install github.com/google/go-jsonnet/cmd/jsonnet@latest && \
    go install github.com/google/go-jsonnet/cmd/jsonnetfmt@latest && \
    go install github.com/google/go-jsonnet/cmd/jsonnet-lint@latest && \
    # install AWS CLI
    apt-get install -y awscli && \
    # install pip and boto3
    apt-get install -y python3-pip && pip3 install boto3 && pip3 install black && \
    # install protobuf compiler tools
    # https://grpc.io/docs/languages/go/quickstart/#prerequisites
    apt-get install -y protobuf-compiler && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    export PATH="$PATH:$(go env GOPATH)/bin"
