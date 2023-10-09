export AWS_DEFAULT_REGION=$AWS_REGION
BUILD_DIR=$SUBSTATION_ROOT/examples/aws/lambda/kinesis/multistream

echo "> Removing Substation configurations from AWS AppConfig" && \
cd $SUBSTATION_ROOT && \
AWS_APPCONFIG_APPLICATION_NAME=substation AWS_APPCONFIG_PROFILE_NAME=publisher python3 build/scripts/aws/appconfig/appconfig_delete.py
AWS_APPCONFIG_APPLICATION_NAME=substation AWS_APPCONFIG_PROFILE_NAME=subscriber python3 build/scripts/aws/appconfig/appconfig_delete.py

echo "> Destroying infrastructure in AWS with Terraform" && \
cd $BUILD_DIR/terraform && \
terraform destroy