FROM mcr.microsoft.com/vscode/devcontainers/go:1.19
RUN apt-get update -y && \
    # https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli
    wget -O- https://apt.releases.hashicorp.com/gpg | \
    gpg --dearmor | \
    sudo tee /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    gpg --no-default-keyring \
    --keyring /usr/share/keyrings/hashicorp-archive-keyring.gpg \
    --fingerprint && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
    https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
    sudo tee /etc/apt/sources.list.d/hashicorp.list && \
    apt update -y && \
    apt install -y terraform && \
    # Install go-jsonnet
    go install github.com/google/go-jsonnet/cmd/jsonnet@latest && \
    go install github.com/google/go-jsonnet/cmd/jsonnetfmt@latest && \
    go install github.com/google/go-jsonnet/cmd/jsonnet-lint@latest && \
    # Install AWS CLI
    apt-get install -y awscli && \
    # Install pip and boto3
    apt-get install -y python3-pip && pip3 install boto3 && pip3 install black
