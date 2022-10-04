# Contributing to Substation

Thank you so much for your interest in contributing to Substation! This document contains guidelines to follow when contributing to the project.

## Table Of Contents

[Code of Conduct](#code-of-conduct)

[Submissions](#submissions)
  + [Changes](#submitting-changes)
  + [Bugs](#submitting-bugs)
  + [Enhancements](#submitting-enhancements)

[Development](#development)
  + [Development Environment](#development-environment)
  + [Testing](#testing)

[Style Guides](#style-guides)
  + [Design Patterns](#design-patterns)
  + [Naming Conventions](#naming-conventions)
  + [Go](#go-style-guide)
  + [Python](#python-style-guide)
  + [Jsonnet](#jsonnet-style-guide)

## Code of Conduct

The Code of Conduct can be reviewed [here](CODE_OF_CONDUCT.md).

## Submissions

### Submitting Changes

Pull requests should be submitted using the pull request template. Changes will be validated through automation and by the project maintainers before merging to main.

### Submitting Bugs

Bugs should be submitted as issues using the issue template.

### Submitting Enhancements

Enhancements should be submitted as issues using the issue template.

## Development

### Development Environment

The project supports development through the use of [Visual Studio Code configurations](https://code.visualstudio.com/docs/remote/containers). The VS Code [development container](.devcontainer/Dockerfile) contains all packages required to develop and test changes locally before submitting pull requests.

### Testing

We rely on contributors to test changes before they are submitted as pull requests. Any components added or changed should be tested and public packages should be supported by unit tests.

## Style Guides

### Design Patterns

##### Configurations

Substation uses a single configuration pattern for all components in the system (see `Config` in [config/config.go](/config/config.go)). This pattern is highly reusable and should be nested to create complex configurations supported by Jsonnet. Below is an example that shows how configurations should be designed:

```json
   "foo": {
	  "settings": { ... },
	  "type": "fooer"
   },
   "bar": {
      "settings": {
         "baz": [
            {
               "settings": { ... },
               "type": "bazar"
            },
         ]
      },
      "type": "barre"
   }
```

Repeating this pattern allows components and applications to integrate with the factory pattern.

##### Factories

Substation relies on [factory methods](https://refactoring.guru/design-patterns/factory-method) to create objects that [satisfy interfaces](https://go.dev/doc/effective_go#interface_methods) across the system. Factories should be combined with the configuration design pattern to create pre-configured, ready-to-use objects.

Factories are the preferred method for allowing users to customize the system. Example factories can be seen in [condition](/condition/condition.go) and [process](/process/process.go).

### Naming Conventions

#### Environment Variables

Substation uses environment variables to customize runtime settings of the system (e.g., concurrency is controlled by `SUBSTATION_CONCURRENCY`). 

Custom applications should implement their own runtime settings as required; for example, the [AWS Lambda application](/cmd/aws/lambda/substation/) uses `SUBSTATION_HANDLER` to manage [invocation settings](https://docs.aws.amazon.com/lambda/latest/dg/lambda-invocation.html).

#### Errors

Errors should always start with `err` (or `Err`, if they are public) and be defined as constants using [internal/errors](/internal/errors/errors.go).

If the error is related to an object created by an interface factory, then the object should be referenced by name in the error. For example, if the factory returns objects named `Foo`, `Bar`, and `Baz`, then related errors should be `errFooShortDescription`, `errBarShortDescription`, and `errBazShortDescription`.

#### Environment Variables

Environment variable keys and values specific to the Substation application should always use SCREAMING_SNAKE_CASE. If the key or value refers to a cloud service provider, then it should always directly refer to that provider (for example, AWS_API_GATEWAY).

Any environment variable that changes a default runtime setting should always start with SUBSTATION (for example, SUBSTATION_CONCURRENCY).

#### Application Variables

Variable names should always follow conventions from [Effective Go](https://go.dev/doc/effective_go#names), the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments#variable-names) and avoid [predeclared identifiers](https://go.dev/ref/spec#Predeclared_identifiers). For example `capsule` is used instead of `cap` to avoid shadowing the capacity function, modifiers and plural usage are `newCapsule`, and `capsules`.

#### Source Metadata

Sources that [add metadata during encapsulation](/config/README.md) should use lowerCamelCase for their JSON keys.

#### Package Configurations

Configurations for packages (for example, [process](/process/README.md) and [condition](/condition/README.md)) should always use lower_snake_case in their JSON keys. This helps maintain readability when reviewing large configuration files.

We strongly urge everyone to use Jsonnet for managing configurations.

### Go Style Guide

Go code should follow [Effective Go](https://go.dev/doc/effective_go) as a baseline.

### Python Style Guide

Python code should follow [Google's Python Style Guide](https://google.github.io/styleguide/pyguide.html) as a baseline.

### Jsonnet Style Guide

Although this repository isn't meant to store Substation configuration files, internally the Security team at Brex follows the guidance described below as a baseline. 

#### jsonnetfmt

Use jsonnetfmt!

#### Variables

If you are referencing a single value more than twice, then it should almost always be defined as a variable. In some cases you may even want to use a variable if it describes a value more clearly than the value's literal JSON key does.

```
bad:
	{
		processors: [
			processlib.copy(
				"ip_address", "client_ip",
				condition_inspectors=[
					conditionlib.regexp("req.details.LocalAddr", "([0-9]{1,3}\.){3}[0-9]{1,3}"),
					conditionlib.regexp("networkDirection", "OUTBOUND"),
				]
			)
		]
	},
	{
		processors: [
			processlib.copy(
				"ip_address", "server_ip",
				condition_inspectors=[
					conditionlib.regexp("req.details.LocalAddr", "([0-9]{1,3}\.){3}[0-9]{1,3}"),
					conditionlib.regexp("networkDirection", "INBOUND"),
				]
			)
		]
	}

better:
    local public_ip_address = req.details.LocalAddr;
    local condition_ip_address = conditionlib.regexp(public_ip_address, "([0-9]{1,3}\.){3}[0-9]{1,3}");

	{
		processors: [
			processlib.copy(
				public_ip_address, "client_ip",
				condition_inspectors=[
					condition_ip_address,
					conditionlib.regexp("networkDirection", "OUTBOUND"),
				]
			)
		]
	},
	{
		processors: [
			processlib.copy(
				public_ip_address, "server_ip",
				condition_inspectors=[
					condition_ip_address,
					conditionlib.regexp("networkDirection", "INBOUND"),
				]
			)
		]
	}
```

#### Functions

If you are using the same config blocks many times, then they should almost always be defined as a function. If the function can be reused across many pipelines, then it should be defined globally.

```
bad:
	"conditions": [
		{
			"type": "regexp",
			"settings": {
				"key": public_ip_address,
				"expression": "([0-9]{1,3}\.){3}[0-9]{1,3}",
				"negate": false,
			},
		},
		{
			"type": "regexp",
			"settings": {
				"key": "event_type",
				"expression": "network_connect",
				"negate": true,
			},
		},
		{
			"type": "regexp",
			"settings": {
				"key": "network_direction",
				"expression": "OUTBOUND",
				"negate": false,
			},
		},
	]

better:
	local regexp(path, expression, negate=false): {
		"type": "regexp",
		"settings": {
			"key": path,
			"expression": expression,
			"negate": negate,
		},
	};

	"conditions": [
		regexp(public_ip_address, "([0-9]{1,3}\.){3}[0-9]{1,3}"),
		regexp("event_type", "network_connect", negate=true),
		regexp("network_direction", "OUTBOUND"),
	]
```

#### For Loops

If you are repeatedly using the same config block, then it should almost always be defined using a for loop.

```
bad:
	[
		{
			local output = "cloud.account.name",
			local conditions = [
				conditions.strings.equals("recipientAccountId", "123")
			],
			"processors": [
				processlib.insert(output, "foo", condition_inspectors=conditions),
			],
		},
		{
			local output = "cloud.account.name",
			local conditions = [
				conditions.strings.equals("recipientAccountId", "456")
			],
			"processors": [
				processlib.insert(output, "bar", condition_inspectors=conditions),
			],
		},
		{
			local output = "cloud.account.name",
			local conditions = [
				conditions.strings.equals("recipientAccountId", "789")
			],
			"processors": [
				processlib.insert(output, "baz", condition_inspectors=conditions),
			],
		},
	]

better:
	local cloudtrail_accounts = {
		"123": "foo",
		"456": "bar",
		"789": "baz",
	};

	[
		{
			local output = "cloud.account.name",
			local conditions = [
				conditions.strings.equals("recipientAccountId", id)
			],
			"processors": [
				processlib.insert(output, cloudtrail_accounts[id], condition_inspectors=conditions),
			],
		}

	for id in std.objectFields(cloudtrail_accounts)
	]
```

#### Organization

Substation configuration files should be organized by pipeline and resource using a hierarchical folder structure: `root/[pipeline]/[resource]/` .

This folder structure supports three levels of configuration:

* global -- configs used in multiple pipelines, stored in `root/foo.libsonnet`
* regional -- configs used in multiple resources of a single pipeline, stored in `root/[pipeline]/foo.libsonnet`
* local -- configs used in one resource of a single pipeline, stored in `root/[pipeline]/[resource]/foo.libsonnet`

Further segmentation of files at the local level is recommended if users want to logically group configs or if a single config becomes too large (the larger the config, the harder it is to understand). For example, configs for processing event data into the [Elastic Common Schema](https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html) (ECS) are easier to manage if they are logically grouped according to the ECS data model (e.g., `client.*` fields are in `client.libsonnet` , `process.*` fields are in `process.libsonnet` , `user.*` fields are in `user.libsonnet` , etc.).
