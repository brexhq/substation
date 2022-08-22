# config

Contains functions for loading configurations and handling data. 

## data encapsulation

Substation encapsulates data during ingest and decapsulates it during load; data in transit is stored in "capsules." Capsules contain two fields:
* data: stores structured or unstructured data
* metadata: stores structured metadata that describes the data

The metadata field is accessed through a special JSON key named "!metadata", any references to this key will get or set the structured data stored in the field. JSON values can be freely moved between the data and metadata fields.

Capsules can be created and initialized using this pattern, where b is a []byte and v is an interface{}:

```go
	cap := NewCapsule()
	cap.SetData(b).SetMetadata(v)
```

Substation applications follow these rules when handling capsules:
* Sources set the initial metadata, but this can be modified in transit by applying processors
* Sinks only output data, but metadata can be retained by copying it from metadata into data
