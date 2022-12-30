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

#### Environment Variables

Substation uses environment variables to customize runtime settings of the system (e.g., concurrency is controlled by `SUBSTATION_CONCURRENCY`). 

Custom applications should implement their own runtime settings as required; for example, the [AWS Lambda application](/cmd/aws/lambda/substation/) uses `SUBSTATION_HANDLER` to manage [invocation settings](https://docs.aws.amazon.com/lambda/latest/dg/lambda-invocation.html).

#### Configurations

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

#### Factories

Substation relies on [factory methods](https://refactoring.guru/design-patterns/factory-method) to create objects that [satisfy interfaces](https://go.dev/doc/effective_go#interface_methods) across the system. Factories should be combined with the configuration design pattern to create pre-configured, ready-to-use objects.

Factories are the preferred method for allowing users to customize the system. Example factories can be seen in [condition](/condition/condition.go) and [process](/process/process.go).

#### Reading and Writing Streaming Data

We prefer to use the io package for reading (e.g., io.Reader) and writing (e.g., io.Writer) streams of data. This reduces memory usage and decreases the likelihood that we will need to refactor methods and functions that handle streaming data.

Substation commonly uses these io compatible containers:

- open files are created by calling `os.CreateTemp("", "substation")`

- bytes buffers are created by calling `new(bytes.Buffer)`

### Naming Conventions

#### Errors

Errors should always start with `err` (or `Err`, if they are public) and be defined as constants using [internal/errors](/internal/errors/errors.go).

If the error is related to an object created by an interface factory, then the object should be referenced by name in the error. For example, if the factory returns objects named `Foo`, `Bar`, and `Baz`, then related errors should be `errFooShortDescription`, `errBarShortDescription`, and `errBazShortDescription`.

#### Environment Variables

Environment variable keys and values specific to the Substation application should always use SCREAMING_SNAKE_CASE. If the key or value refers to a cloud service provider, then it should always directly refer to that provider (for example, AWS_API_GATEWAY).

Any environment variable that changes a default runtime setting should always start with SUBSTATION (for example, SUBSTATION_CONCURRENCY).

#### Application Variables

Variable names should always follow conventions from [Effective Go](https://go.dev/doc/effective_go#names), the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments#variable-names) and avoid [predeclared identifiers](https://go.dev/ref/spec#Predeclared_identifiers). For example `capsule` is used instead of `cap` to avoid shadowing the capacity function while modifiers and plural usage are `newCapsule` and `capsules`.

#### Source Metadata

Sources that [add metadata during encapsulation](/config/README.md) should use lowerCamelCase for their JSON keys.

#### Package Configurations

Configurations for packages (for example, [process](/process/README.md) and [condition](/condition/README.md)) should always use lower_snake_case in their JSON keys. This helps maintain readability when reviewing large configuration files.

We strongly urge everyone to use Jsonnet for managing configurations.

### Go Style Guide

Go code should follow [Effective Go](https://go.dev/doc/effective_go) as a baseline.

### Python Style Guide

Python code should follow [Google's Python Style Guide](https://google.github.io/styleguide/pyguide.html) as a baseline.
