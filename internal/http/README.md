## http
Contains functions for managing HTTP requests. Substation follows these rules across every application:
* HTTP clients are always retryable clients from [this package](github.com/hashicorp/go-retryablehttp)
* For AWS deployments, HTTP clients enable AWS X-Ray
