# compile.sh

Used for recursively compiling Substation [Jsonnet](https://jsonnet.org/) config files (`config.jsonnet`) into JSON; compiled files are stored in the same directory as the Jsonnet files. The script should be executed from the project root: `sh build/config/compile.sh`.

# aws/appconfig_upload.py

Used for uploading and deploying compiled Substation JSON config files to AWS AppConfig. This script has some dependencies:

- boto3 must be installed
- AWS credentials for reading and writing to AppConfig
- AppConfig infrastructure must be ready to use (see [examples/aws/terraform/bootstrap.tf](/examples/aws/terraform/bootstrap.tf) for an example)

This script is intended to be deployed to a CI / CD pipeline (e.g., GitHub Actions, Circle CI, Jenkins, etc.), but can be run locally if needed. See [examples/aws/](/examples/aws/) for example usage.
