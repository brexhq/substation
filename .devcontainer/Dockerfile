FROM mcr.microsoft.com/vscode/devcontainers/go:1.17
RUN sudo apt-get update && \
    # install Terraform
    cd /tmp/ && \
    curl https://releases.hashicorp.com/terraform/1.1.7/terraform_1.1.7_linux_$(dpkg --print-architecture).zip -o ./terraform.zip && \
    unzip terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    # install Jsonnet
    sudo apt-get install -y jsonnet && \
    # install AWS CLI
    sudo apt-get install -y awscli && \
    # install pip and boto3
    sudo apt-get install -y python3-pip && pip3 install boto3 && pip3 install black
