"""
Manages the validation check before deployment of Substation configurations
in AWS AppConfig.

This is intended to be deployed to a CI/CD
pipeline (e.g., GitHub Actions, Circle CI, Jenkins, etc.) for automated
configuration tests.

Typical usage example:

    SUBSTATION_CONFIG_DIRECTORY=config AWS_APPCONFIG_VALIDATOR=substation_validator python3 appconfig_validate.py
"""
import json
import os
import base64
import sys

import boto3
import botocore

LAMBDA = boto3.client("lambda")


def main():
    appconfig_validator = os.environ.get("AWS_APPCONFIG_VALIDATOR")
    if not appconfig_validator:
        print("environment variable AWS_APPCONFIG_VALIDATOR missing")
        sys.exit(1)

    configs = os.environ.get("SUBSTATION_CONFIG_DIRECTORY")
    if not configs:
        print("environment variable SUBSTATION_CONFIG_DIRECTORY missing")
        sys.exit(1)

    try:
        LAMBDA.get_function(FunctionName=appconfig_validator)
    except botocore.exceptions.ClientError as e:
        print(f'Lambda function "{appconfig_validator}" not found: {e}.')
        sys.exit(1)

    versions = set()
    try:
        resp = LAMBDA.list_aliases(FunctionName=appconfig_validator)
        versions = set([alias.get("Name") for alias in resp.get("Aliases", [])])
    except botocore.exceptions.ClientError as e:
        print(f'Failed getting aliases for "{appconfig_validator}": {e}')
        sys.exit(1)

    # file_map is populated with entries that match this pattern:
    #   substation_example_dynamodb = examples/aws/config/substation_example_dynamodb/config.json
    #   substation_example_kinesis = examples/aws/config/substation_example_kinesis/config.json
    file_map = {}
    for r, _, f in os.walk(configs):
        for file_ in f:
            if file_ == "config.json":
                path = r.split("/")[-1]
                path_full = f"{r}/{file_}"
                file_map[path] = path_full

    for file_ in file_map:
        with open(file_map[file_], "rb") as fin:
            tmp = json.loads(fin.read())

            # only validate when a configuration declares a version, this adds
            # an opt out mechanism that shared Seshat configurations use.
            ver = tmp.get("version")
            if ver is None:
                continue

            # Default to validating with the latest version if
            # the configuration version doesn't exist.
            version = ver if ver in versions else "$LATEST"
            fn = f"{appconfig_validator}:{version}"

            try:
                resp = LAMBDA.invoke(
                    FunctionName=fn,
                    Payload=json.dumps(
                        {
                            "content": base64.b64encode(
                                json.dumps(tmp).encode("ascii")
                            ).decode("ascii"),
                        }
                    ),
                    LogType="None",
                )

                if resp.get("FunctionError", "") == "Unhandled":
                    error = json.loads(resp.get("Payload").read()).get("errorMessage")
                    print(f"Issue validating {file_}: {error}")
                    sys.exit(1)

            except Exception as e:
                print(
                    f"Couldn't invoke function {appconfig_validator} for {file_}: {e}"
                )
                sys.exit(1)

        print(f"verified configuration of {file_} against {fn}.")


if __name__ == "__main__":
    main()
