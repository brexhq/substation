# aws

Contains functions for managing AWS API calls. Substation follows these rules across every application:
* AWS clients are configured using [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html)
* AWS clients use service interface APIs (e.g., s3iface, kinesisiface, etc.)
* AWS clients enable [X-Ray](https://aws.amazon.com/xray/) for tracing if a [daemon address](https://docs.aws.amazon.com/xray/latest/devguide/xray-sdk-go-configuration.html#xray-sdk-go-configuration-envvars) is found
