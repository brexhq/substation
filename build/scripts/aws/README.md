# aws

This directory contains scripts that are used to manage Substation deployments in AWS.

## appconfig

Contains scripts for uploading, deleting, and validating Substation configurations in AWS AppConfig. These are best used in a CI / CD pipeline, such as GitHub Actions:

```yaml

  deploy_substation:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: write
    steps:
      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@master
        with:
          aws-region: us-east-1
          role-to-assume: arn:aws:iam::012345678901:role/substation_cicd
          role-session-name: substation_cicd

      - name: Checkout Code
        uses: actions/checkout@v2
        with:
          fetch-depth: 2

      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23

      - name: Install Jsonnet
        run: |
          go install github.com/google/go-jsonnet/cmd/jsonnet@latest

      - uses: actions/setup-python@v2
      - name: Deploy Configs
        run: |
          pip3 install -r requirements.txt 
          bash compile.sh
          python3 appconfig_upload.py
        env:
          AWS_DEFAULT_REGION: "us-east-1"
          # These are required by the appconfig_upload.py script.
          AWS_APPCONFIG_DEPLOYMENT_STRATEGY: "Instant"
          AWS_APPCONFIG_ENVIRONMENT: "example"
          AWS_APPCONFIG_APPLICATION_NAME: "substation"
          SUBSTATION_CONFIG_DIRECTORY: "path/to/configs"
```

## dynamodb

### bulk_delete_items.py

Python script that deletes items from a DynamoDB table based on a JSON Lines file.

## kinesis

### put_records.py

Python script that puts records into a Kinesis stream by reading a text file. Each line in the text file is sent as a record to the Kinesis stream.

## lambda

### get_appconfig_extension.sh

Bash script that is used to download the [AWS AppConfig Lambda extension](https://docs.aws.amazon.com/appconfig/latest/userguide/appconfig-integration-lambda-extensions.html) for any AWS region. This extension is required for deploying Substation to AWS Lambda.

## s3

### s3_rehydration.py

Python script that rehydrates data from an S3 bucket into an SNS topic by simulating S3 
object creation events.
