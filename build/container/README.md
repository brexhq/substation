# container

This directory contains Docker build files for applications. Containers should be built from the project root using a command such as `docker build -f build/container/path/to/app/Dockerfile -t tag:latest .`

## aws

Images stored in AWS ECR should be built using environment variables: `docker build -f build/container/aws/lambda/substation/Dockerfile -t $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:latest .`

We recommend tagging images with the Semantic Version of each release: `docker tag foo $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:$(git describe --abbrev=0 --tags)`

To specify the instruction set architecture, use the ARCH build arg: `docker build --build-arg ARCH=$ARCH -f build/container/aws/lambda/substation/Dockerfile -t $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:latest-$ARCH .`
