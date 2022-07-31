# compile.sh

Used for recursively compiling Substation [Jsonnet](https://jsonnet.org/) config files (`config.jsonnet`) into JSON; compiled files are stored in the same directory as the Jsonnet files. The script should be executed from the project root: `sh build/config/compile.sh`.

# aws/appconfig_upload.py

Used for uploading and deploying compiled Substation JSON config files to AWS AppConfig. This script has some dependencies:

- boto3 must be installed
- AWS credentials for reading and writing to AppConfig
- AppConfig infrastructure should be built using `build/terraform/aws/bootstrap.tf` (or similar settings should be managed externally)

This script is designed to run in a CI/CD tool such as GitHub Actions (but can be run locally if needed) and should be executed from the project root: `python3 build/config/aws/appconfig_upload.py`.
