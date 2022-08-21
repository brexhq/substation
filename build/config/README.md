# config

## *.libsonnet

Contain [Jsonnet](https://jsonnet.org/) functions for managing configurations. These functions map to the core ingest, transform, and load functionality of Substation. See [examples/](/examples/) for example usage.

## compile.sh

Used for recursively compiling Substation Jsonnet config files ( `config.jsonnet` ) into JSON; compiled files are stored in the same directory as the Jsonnet files. 

This script is intended to be deployed to a CI / CD pipeline (e.g., GitHub Actions, Circle CI, Jenkins, etc.), but can be run locally if needed.

## aws/appconfig_upload.py

Used for uploading and deploying compiled Substation JSON config files to AWS AppConfig. This script has some dependencies:

* boto3 must be installed
* AWS credentials for reading and writing to AppConfig
* AppConfig infrastructure must be ready to use (see [examples/aws/terraform/bootstrap.tf](/examples/aws/terraform/bootstrap.tf) for an example)

This script is intended to be deployed to a CI / CD pipeline (e.g., GitHub Actions, Circle CI, Jenkins, etc.), but can be run locally if needed. See [examples/aws/](/examples/aws/) for example usage.

## Using Jsonnet

We highly recommend the use of Jsonnet for managing Substation configurations. In addition to the [tutorial](https://jsonnet.org/learning/tutorial.html), we've included some supporting information below on getting started with Jsonnet.

### Jsonnet vs JSON

Below is a complete example that shows the advantages of Jsonnet over JSON. The Jsonnet code below contains four processors that process S3 object metadata from AWS CloudTrail events and map the data to the Elastic Common Schema. Note that the config uses imported functions and local variables -- these increase config reuse and reduce the chances of error-prone JSON.

```
local conditionlib = import '../../condition.libsonnet';
local processlib = import '../../process.libsonnet';
local regexp = import '../../regexp.libsonnet';

local path_ecs_file_path = 'file.path';
local path_ecs_file_name = 'file.name';

local condition_object = conditionlib.strings.endswith('eventName', 'Object');
local condition_s3 = conditionlib.strings.equals('eventSource', 's3.amazonaws.com');

local processors = [
  {
    local inputs = [
      'requestParameters.bucketName',
      'requestParameters.key',
    ],
    local output = path_ecs_file_path,
    // if the event is sourced from S3 and contains an object, then run the processors
    local conditions = [
      condition_s3,
      condition_object,
    ],
    // concatenate the bucket name and key into a single string separated by '/'
    processors: [
      processlib.concat(inputs, output, '/', condition_operator='and', condition_inspectors=conditions),
    ],
  },
  {
    local input = 'requestParameters.key',
    local output = path_ecs_file_name,
    local conditions = [
      condition_s3,
      condition_object,
    ],
    // the file name of the key is captured using a regular expression designed for Unix-like files
    processors: [
      processlib.capture(input, output, regexp.file.name.nix, condition_operator='and', condition_inspectors=conditions),
    ],
  },
  {
    local input = path_ecs_file_path,
    local output = 'file.directory',
    // if the input IS NOT empty, then run the processors
    local conditions = [
      conditionlib.strings.empty(input, negate=true),
    ],
    // the file path is derived from the concatenation of the bucket name and key, then captured using a regular expression designed for Unix file directories
    processors: [
      processlib.capture(input, output, regexp.file.directory.nix, condition_operator='and', condition_inspectors=conditions),
    ],
  },
  {
    local input = path_ecs_file_name,
    local output = 'file.extension',
    local conditions = [
      conditionlib.strings.empty(input, negate=true),
    ],
    // the file extension is captured using a regular expression designed for generic files (i.e., any characters after the last '.' observed in the input)
    processors: [
      processlib.capture(input, output, regexp.file.extension.miscellaneous, condition_operator='and', condition_inspectors=conditions),
    ],
  },
];

// loops through the processors above and inserts them as a list into a JSON path named "processors"
// creates an output in the shape of ...
// "processors": [
//  { ... },
//  { ... },
//  { ... },
//  { ... },
// ]
{
  processors: std.flattenArrays([p.processors for p in processors]),
}
```

When the Jsonnet code is compiled, it produces this JSON:

```
"processors": [
  {
      "settings": {
        "condition": {
            "inspectors": [
              {
                  "settings": {
                    "expression": "s3.amazonaws.com",
                    "function": "equals",
                    "negate": false,
                    "key": "eventSource"
                  },
                  "type": "strings"
              },
              {
                  "settings": {
                    "expression": "Object",
                    "function": "endswith",
                    "negate": false,
                    "key": "eventName"
                  },
                  "type": "strings"
              }
            ],
            "operator": "and"
        },
        "input": {
            "keys": [
              "requestParameters.bucketName",
              "requestParameters.key"
            ]
        },
        "options": {
            "separator": "/"
        },
        "output": {
            "key": "file.path"
        }
      },
      "type": "concat"
  },
  {
      "settings": {
        "condition": {
            "inspectors": [
              {
                  "settings": {
                    "expression": "s3.amazonaws.com",
                    "function": "equals",
                    "negate": false,
                    "key": "eventSource"
                  },
                  "type": "strings"
              },
              {
                  "settings": {
                    "expression": "Object",
                    "function": "endswith",
                    "negate": false,
                    "key": "eventName"
                  },
                  "type": "strings"
              }
            ],
            "operator": "and"
        },
        "input": {
            "key": "requestParameters.key"
        },
        "options": {
            "count": 1,
            "expression": "/([^/]+)$"
        },
        "output": {
            "key": "file.name"
        }
      },
      "type": "capture"
  },
  {
      "settings": {
        "condition": {
            "inspectors": [
              {
                  "settings": {
                    "expression": "",
                    "function": "equals",
                    "negate": true,
                    "key": "file.path"
                  },
                  "type": "strings"
              }
            ],
            "operator": "and"
        },
        "input": {
            "key": "file.path"
        },
        "options": {
            "count": 1,
            "expression": "(.*)/[^/]+$"
        },
        "output": {
            "key": "file.directory"
        }
      },
      "type": "capture"
  },
  {
      "settings": {
        "condition": {
            "inspectors": [
              {
                  "settings": {
                    "expression": "",
                    "function": "equals",
                    "negate": true,
                    "key": "file.name"
                  },
                  "type": "strings"
              }
            ],
            "operator": "and"
        },
        "input": {
            "key": "file.name"
        },
        "options": {
            "count": 1,
            "expression": "\\.([^\\.]+)$"
        },
        "output": {
            "key": "file.extension"
        }
      },
      "type": "capture"
  }
]
```

Put simply, Jsonnet is easier to use and more manageable than JSON for projects with large configurations.

### organization

Substation config files should be organized by pipeline and app using a hierarchical folder structure: `root/[pipeline]/[app]/` .

This folder structure supports three levels of configuration:

* global -- configs used in multiple pipelines, stored in `root/foo.libsonnet`
* regional -- configs used in multiple apps of a single pipeline, stored in `root/[pipeline]/foo.libsonnet`
* local -- configs used in one component of a single pipeline, stored in `root/[pipeline]/[app]/foo.libsonnet`

Further segmentation of files at the local level is recommended if users want to logically group configs or if a single config becomes too large (the larger the config, the harder it is to understand). For example, configs for processing event data into the [Elastic Common Schema](https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html) (ECS) are easier to manage if they are logically grouped according to the ECS data model (e.g., `client.*` fields are in `client.libsonnet` , `process.*` fields are in `process.libsonnet` , `user.*` fields are in `user.libsonnet` , etc.).

### style guide

#### jsonnetfmt

Use jsonnetfmt!

#### variables

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

#### functions

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

#### for loops

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
