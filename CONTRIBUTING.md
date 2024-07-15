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

Applications may implement runtime settings that are managed by environment variables. For example, the [AWS Lambda application](/cmd/aws/lambda/substation/) uses `SUBSTATION_LAMBDA_HANDLER` to manage [invocation settings](https://docs.aws.amazon.com/lambda/latest/dg/lambda-invocation.html). These should reference the application by name, if possible.

#### Configurations

Substation uses a single configuration pattern for all components in the system (see `Config` in [config/config.go](/config/config.go)). This pattern is highly reusable and should be embedded to create custom configurations. Below is an example that shows how configurations should be designed:

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

Repeating this pattern allows components and applications to integrate with Substation's factory patterns.

#### Factories

Substation relies on [factory methods](https://refactoring.guru/design-patterns/factory-method) to create objects that [satisfy interfaces](https://go.dev/doc/effective_go#interface_methods) across the project. Factories should be combined with the configuration design pattern to create new components.

Factories are the preferred method for allowing users to customize the system. Example factories can be seen in [condition](/condition/condition.go) and [transform](/transform/transform.go).

#### Reading and Writing Streaming Data

We prefer to use the io package for reading (e.g., io.Reader) and writing (e.g., io.Writer) streams of data. This reduces memory usage and decreases the likelihood that we will need to refactor methods and functions that handle streaming data.

Substation commonly uses these io compatible containers:

- open files are created by calling `os.CreateTemp("", "substation")`

- bytes buffers are created by calling `new(bytes.Buffer)`

### Naming Conventions

#### Breaking Changes

Any change that modifies the public API of Go packages and applications is a breaking change, and any source code that has non-obvious impact on the public API should be tagged with `BREAKING CHANGE` in a comment.

#### Errors

Errors should always start with `err` (or `Err`, if they are public). Commonly used errors are defined in [internal/errors.go](internal/errors.go).

If the error is related to a specific component, then the component name should be included in the error. For example, if the error is related to the `Foo` component, then the error should be named `errFooShortDescription`.

#### Environment Variables

Environment variable keys and values specific to the Substation application should always use SCREAMING_SNAKE_CASE. If the key or value refers to a cloud service provider, then it should always directly refer to that provider (for example, AWS_API_GATEWAY).

Any environment variable that changes a default runtime setting should always start with SUBSTATION (for example, SUBSTATION_CONCURRENCY).

#### Application Variables

Variable names should always follow conventions from [Effective Go](https://go.dev/doc/effective_go#names), the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments#variable-names) and avoid [predeclared identifiers](https://go.dev/ref/spec#Predeclared_identifiers).

#### Source Metadata

Sources that [add metadata during message creation](/message/) should use lowerCamelCase for their JSON keys.

#### Package Configurations

Configurations for packages (for example, [condition](/condition/README.md) and [transform](/transform/README.md)) should always use lower_snake_case in their JSON keys. This helps maintain readability when reviewing large configuration files.

We strongly urge everyone to use Jsonnet for managing configurations.

### Go Style Guide

Go code should follow [Effective Go](https://go.dev/doc/effective_go) as a baseline.

### Python Style Guide

Python code should follow [Google's Python Style Guide](https://google.github.io/styleguide/pyguide.html) as a baseline.
