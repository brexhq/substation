# config

## *.libsonnet

Contain [Jsonnet](https://jsonnet.org/) functions for managing configurations. These functions map to the core ingest, transform, and load functionality of Substation. See [examples/](/examples/) for example usage.

## compile.sh

Used for recursively compiling Substation Jsonnet config files ( `config.jsonnet` ) into JSON; compiled files are stored in the same directory as the Jsonnet files. 

This script is intended to be deployed to a CI / CD pipeline (e.g., GitHub Actions, Circle CI, Jenkins, etc.), but can be run locally if needed.

## aws/appconfig_delete.py

Deletes application profiles, including all hosted configurations, in AWS AppConfig. This script has some dependencies:

* boto3 must be installed
* AWS credentials for reading from and deleting in AppConfig

This script can be used after deleting any AWS Lambda that have application profiles.

## aws/appconfig_upload.py

Manages the upload and deployment of compiled Substation JSON configuration files in AWS AppConfig. This script has some dependencies:

* boto3 must be installed
* AWS credentials for reading from and writing to AppConfig
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
