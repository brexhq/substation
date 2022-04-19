# container
Contains Docker build files for each application under `cmd/`. Containers should be built from the project root using a command like this: `docker build -f build/container/path/to/app/Dockerfile -t tag:latest .`

## aws
Images to be stored in AWS ECR can be built like this using environment variables: `docker build -f build/container/aws/lambda/substation/Dockerfile -t $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:latest .`

We recommend tagging images with the Semantic Version of each release: `docker tag foo $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:$(git describe --abbrev=0 --tags)`

To specify the instruction set architecture, use the AWS_ARCHITECTURE build arg: `docker build --build-arg AWS_ARCHITECTURE=$AWS_ARCHITECTURE -f build/container/aws/lambda/substation/Dockerfile -t $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:latest-$AWS_ARCHITECTURE .`

## file
Images to be run from local file systems can be built like this: `docker build -f build/container/file/Dockerfile -t file:latest .`

Using the quickstart as an example, the image can be run like this:`docker run -v "$(pwd)":/tmp file:latest /bin/substation -config /tmp/config/quickstart/config.json -input /tmp/quickstart.json`
