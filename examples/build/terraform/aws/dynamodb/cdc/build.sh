if [ -z "$SUBSTATION_ROOT" ]
then
  >&2 echo "Error: SUBSTATION_ROOT not set. This is the root directory of the Substation repository."
  exit 1
fi

if [ -z "$AWS_ACCOUNT_ID" ]
then
  >&2 echo "Error: AWS_ACCOUNT_ID not set."
  exit 1
fi

if [ -z "$AWS_REGION" ]
then
  >&2 echo "Error: AWS_REGION not set."
  exit 1
fi

if [ -z "$AWS_ARCHITECTURE" ]
then
  >&2 echo "Error: AWS_ARCHITECTURE is empty and must be set. Valid values are x86_64 and arm64."
  exit 1
fi

export AWS_DEFAULT_REGION=$AWS_REGION
BUILD_DIR=$SUBSTATION_ROOT/examples/aws/lambda/dynamodb/cdc

echo "> Deploying infrastructure in AWS with Terraform" && \
cd $BUILD_DIR/terraform && \
terraform init && \
terraform apply \
-target=module.kms_substation \
-target=aws_appconfig_application.substation \
-target=aws_appconfig_environment.prod \
-target=aws_appconfig_environment.dev \
-target=aws_appconfig_deployment_strategy.instant \
-target=module.ecr_substation

echo "> Building Substation container images and pushing to AWS ECR" && \
cd $SUBSTATION_ROOT && \
sh build/scripts/aws/lambda/get_appconfig_extension.sh && \
docker build --build-arg AWS_ARCHITECTURE=$AWS_ARCHITECTURE -f build/container/aws/lambda/substation/Dockerfile -t substation:latest-$AWS_ARCHITECTURE . && \
# In production environments the tag should match the version of Substation being deployed
docker tag substation:latest-$AWS_ARCHITECTURE $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:latest && \
sh build/scripts/aws/ecr_login.sh && \
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/substation:latest

echo "> Deploying Substation nodes with Terraform" && \
cd $BUILD_DIR/terraform && \
terraform apply

echo "> Compiling Substation configurations and uploading to AWS AppConfig" && \
cd $SUBSTATION_ROOT && \
sh build/scripts/config/compile.sh && \
SUBSTATION_CONFIG_DIRECTORY=$BUILD_DIR AWS_APPCONFIG_APPLICATION_NAME=substation AWS_APPCONFIG_ENVIRONMENT=example AWS_APPCONFIG_DEPLOYMENT_STRATEGY=Instant python3 build/scripts/aws/appconfig/appconfig_upload.py
