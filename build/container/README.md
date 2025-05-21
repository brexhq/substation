# container

This directory contains Docker build files for applications. Containers should be built from the project root using a command such as `docker build -f build/container/path/to/app/Dockerfile -t tag:latest .`

## aws

We recommend building images from within the development container by running these commands:

```bash
# Set environment variables.
export SUBSTATION_VERSION=$(git describe --abbrev=0 --tags)  # Uses the release as the image tag.
export AWS_ARCHITECTURE=arm64  # Either "arm64" or "x86_64".
export AWS_ACCOUNT_ID=012345678901
export AWS_REGION=us-east-1

# Build the image. We recommend using arm64 for AWS Lambda.
bash build/scripts/aws/lambda/get_appconfig_extension.sh
docker buildx build --platform linux/arm64 --build-arg ARCH=$AWS_ARCHITECTURE -f build/container/aws/lambda/substation/Dockerfile -t substation:latest-arm64 .
docker tag substation:latest-arm64 $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:$SUBSTATION_VERSION

# Push the image to ECR.
aws ecr get-login-password | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:$SUBSTATION_VERSION
```

## gcp

We recommend building images from within the development container by running these commands:

```bash
# Set environment variables.
export SUBSTATION_VERSION=$(git describe --abbrev=0 --tags)  # Uses the release as the image tag.
export GCP_REGION=us-east1
export GCP_PROJECT_ID=my-project-id

# Build the image. Cloud Run requires an amd64 image.
docker buildx build --platform linux/amd64 -f build/container/gcp/function/substation/Dockerfile -t substation:latest .
docker tag substation:latest $GCP_REGION-docker.pkg.dev/$GCP_PROJECT_ID/substation/$SUBSTATION_VERSION

# Push the image to the registry.
gcloud auth configure-docker $GCP_REGION-docker.pkg.dev
docker push $GCP_REGION-docker.pkg.dev/$GCP_PROJECT_ID/substation/$SUBSTATION_VERSION
```
