"""Deletes application profiles, including hosted configurations, in AWS AppConfig.

Enumerates hosted configurations in AppConfig for a user-defined application profile and deletes each configuration. When all configurations are deleted, the profile is deleted.

	Typical usage example:

	AWS_APPCONFIG_APPLICATION_NAME=substation AWS_APPCONFIG_PROFILE_NAME=foo python3 appconfig_delete.py
"""
import os

import boto3


def main():
    client = boto3.client("appconfig")

    app_name = os.environ.get("AWS_APPCONFIG_APPLICATION_NAME")
    if not app_name:
        print(f"environment variable AWS_APPCONFIG_APPLICATION_NAME missing")
        return

    profile_name = os.environ.get("AWS_APPCONFIG_PROFILE_NAME")
    if not profile_name:
        print(f"environment variable AWS_APPCONFIG_PROFILE_NAME missing")
        return

    application_map = {}
    applications = client.list_applications()
    for a in applications.get("Items"):
        application_map[a.get("Name")] = a.get("Id")

    application_id = application_map.get(app_name)
    if not application_id:
        print(f"application {app_name} not found in AppConfig")
        return

    profile_map = {}
    profiles = client.list_configuration_profiles(ApplicationId=application_id)
    for p in profiles.get("Items"):
        profile_map[p.get("Name")] = p.get("Id")

    profile_id = profile_map.get(profile_name)
    if not profile_id:
        print(
            f"profile {profile_name} does not exist in AppConfig application {app_name}"
        )
        return

    versions = client.list_hosted_configuration_versions(
        ApplicationId=application_id,
        ConfigurationProfileId=profile_id,
    )

    items = versions.get("Items", [])
    version = items[0].get("VersionNumber", 0)

    if not version:
        print(
            f"profile {profile_name} has no configurations in AppConfig application {app_name}"
        )

    # AppConfig returns version numbers in reverse order (the latest version is listed first), so we iterate, delete, and decrement versions starting from the latest version and break on 0. when 0 is reached, no configurations remain.
    while version:
        client.delete_hosted_configuration_version(
            ApplicationId=application_id,
            ConfigurationProfileId=profile_id,
            VersionNumber=version,
        )

        version -= 1

    client.delete_configuration_profile(
        ApplicationId=application_id,
        ConfigurationProfileId=profile_id,
    )

    print(f"deleted profile {profile_name} from AppConfig application {app_name}")


if __name__ == "__main__":
    main()
