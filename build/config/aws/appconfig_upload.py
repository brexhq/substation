# manages upload and deployment of Substation configurations in AWS AppConfig
# this script is intended to be deployed to a CI/CD pipeline (e.g., GitHub Actions, Circle CI, Jenkins, etc.) for automated config management
# example invocation: SUBSTATION_CONFIG_DIRECTORY=examples APPCONFIG_APPLICATION_NAME=substation APPCONFIG_ENVIRONMENT=prod APPCONFIG_DEPLOYMENT_STRATEGY=Instant python3 build/config/aws/appconfig_upload.py

import json
import os
import time

import boto3
from botocore.config import Config


def main():
    client = boto3.client("appconfig")

    app_name = os.environ.get("APPCONFIG_APPLICATION_NAME")
    if not app_name:
        print(f"environment variable APPCONFIG_APPLICATION_NAME missing")
        return

    env_name = os.environ.get("APPCONFIG_ENVIRONMENT")
    if not env_name:
        print(f"environment variable APPCONFIG_ENVIRONMENT missing")
        return

    strategy_name = os.environ.get("APPCONFIG_DEPLOYMENT_STRATEGY")
    if not strategy_name:
        print(f"environment variable APPCONFIG_DEPLOYMENT_STRATEGY missing")
        return

    configs = os.environ.get("SUBSTATION_CONFIG_DIRECTORY")
    if not configs:
        print(f"environment variable SUBSTATION_CONFIG_DIRECTORY missing")
        return

    applications = client.list_applications()
    application_map = {}
    for a in applications.get("Items"):
        application_map[a.get("Name")] = a.get("Id")
    application_id = application_map.get(app_name)
    if not application_id:
        print(f"application {app_name} not found in AppConfig")
        return

    deployment_strategies = client.list_deployment_strategies()
    deployment_map = {}
    for s in deployment_strategies.get("Items"):
        deployment_map[s.get("Name")] = s.get("Id")

    if strategy_name not in deployment_map:
        print(
            f"deployment strategy {strategy_name} does not exist in AppConfig application {app_name}"
        )
        return

    environments = client.list_environments(ApplicationId=application_id)
    environment_map = {}
    for e in environments.get("Items"):
        environment_map[e.get("Name")] = e.get("Id")

    if env_name not in environment_map:
        print(
            f"environment {env_name} does not exist in AppConfig application {app_name}"
        )
        return

    # enumerated profiles are later matched against config file directories to identify which profile should receive new configs
    profiles = client.list_configuration_profiles(ApplicationId=application_id)
    profile_map = {}
    for p in profiles.get("Items"):
        profile_map[p.get("Name")] = p.get("Id")

    # profile versions are later used to determine if configs are new and should be deployed, this avoids unnecessary deployments
    profile_versions = {}
    for profile in profile_map:
        token = 1
        while token:
            if token:
                versions = client.list_hosted_configuration_versions(
                    ApplicationId=application_id,
                    ConfigurationProfileId=profile_map[profile],
                )
                token = versions.get("NextToken")
            else:
                versions = client.list_hosted_configuration_versions(
                    ApplicationId=application_id,
                    ConfigurationProfileId=profile_map[profile],
                    NextToken=token,
                )
                token = versions.get("NextToken")

        items = versions.get("Items", [])
        if items:
            profile_versions[profile] = items[0].get("VersionNumber")

    # file_map is populated with entries that match this pattern:
    #   substation_example_dynamodb = examples/aws/config/substation_example_dynamodb/config.json
    #   substation_example_kinesis = examples/aws/config/substation_example_kinesis/config.json
    #   substation_autoscaling = examples/aws/config/substation_autoscaling/config.json
    file_map = {}
    for r, d, f in os.walk(configs):
        for file_ in f:
            if file_ == "config.json":
                path = r.split("/")[-1]
                path_full = f"{r}/{file_}"
                file_map[path] = path_full

    for file_ in file_map:
        profile_id = profile_map.get(file_)
        if not profile_id:
            print(
                f"profile {file_} not found in AppConfig application {app_name}, skipping"
            )
            continue

        with open(file_map[file_], "rb") as fin:
            tmp = json.loads(fin.read())

            create = client.create_hosted_configuration_version(
                ApplicationId=application_id,
                ConfigurationProfileId=profile_id,
                Content=json.dumps(tmp, separators=(",", ":")),
                ContentType="application/json",
            )
            version = create.get("VersionNumber")

            # if the current version matches the previous version, then don't deploy
            if file_ in profile_versions and profile_versions[file_] == version:
                print(f"config for {file_} matches latest version, skipping")
                continue

        deploy = client.start_deployment(
            ApplicationId=application_id,
            EnvironmentId=environment_map[env_name],
            ConfigurationProfileId=profile_id,
            DeploymentStrategyId=deployment_map[strategy_name],
            ConfigurationVersion=str(version),
        )

        while 1:
            state = client.get_deployment(
                ApplicationId=application_id,
                EnvironmentId=environment_map[env_name],
                DeploymentNumber=deploy.get("DeploymentNumber"),
            ).get("State")
            if state == "COMPLETE":
                break

            time.sleep(0.25)

        print(f"deployed latest version of {file_}")


if __name__ == "__main__":
    main()
