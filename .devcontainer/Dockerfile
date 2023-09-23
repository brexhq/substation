FROM mcr.microsoft.com/vscode/devcontainers/go:1.19
RUN apt-get update -y && \
    # https://developer.hashicorp.com/terraform/downloads
    wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list && \
    apt update -y && apt install -y terraform && \
    # go-jsonnet
    go install github.com/google/go-jsonnet/cmd/jsonnet@latest && \
    go install github.com/google/go-jsonnet/cmd/jsonnetfmt@latest && \
    go install github.com/google/go-jsonnet/cmd/jsonnet-lint@latest && \
    # AWS CLI
    apt-get install -y awscli && \
    # python
    apt-get install -y python3 python3-boto3 black
