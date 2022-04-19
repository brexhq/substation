# aws
Contains functions for managing AWS API calls. Substation follows these rules across every application:
* AWS clients are always X-Ray enabled for tracing
* AWS clients are configured using environment variables (https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html)
* AWS clients always use the service's interface API (e.g., s3iface, kinesisiface, etc.)
