# Changelog

## [1.1.2](https://github.com/brexhq/substation/compare/v1.1.1...v1.1.2) (2024-03-23)


### Bug Fixes

* **enrich_kv_store_set:** default to 0 duration ([#151](https://github.com/brexhq/substation/issues/151)) ([8a1d31c](https://github.com/brexhq/substation/commit/8a1d31ccc457f92028135176b7f650630d09bee9))

## [1.1.1](https://github.com/brexhq/substation/compare/v1.1.0...v1.1.1) (2024-03-21)


### Performance Improvements

* **aws:** Reduce Aggregated Kinesis Record Size ([#147](https://github.com/brexhq/substation/issues/147)) ([a0ef232](https://github.com/brexhq/substation/commit/a0ef23263c8a34bcaa413c1cdd1a9c80b8d2f55a))
* **transform:** Improve AggregateToArray Throughput ([#150](https://github.com/brexhq/substation/issues/150)) ([d730cc6](https://github.com/brexhq/substation/commit/d730cc6b967b6ee853ebf5f6b18f1ead9c1bcb55))

## [1.1.0](https://github.com/brexhq/substation/compare/v1.0.0...v1.1.0) (2024-03-14)


### Features

* **autoscale:** Dynamic Kinesis Scaling with Thresholds ([#144](https://github.com/brexhq/substation/issues/144)) ([079fda9](https://github.com/brexhq/substation/commit/079fda9e46f70d8544568ed2ca941416212356d0))
* **transform:** Add Metrics Bytes Transform ([#142](https://github.com/brexhq/substation/issues/142)) ([d708580](https://github.com/brexhq/substation/commit/d70858085f1be265a02ac5f1298efd986e7f275e))

## [1.0.0](https://github.com/brexhq/substation/compare/v1.0.0-rc.1...v1.0.0) (2024-03-05)


### ⚠ BREAKING CHANGES

* `cmd/development/substation` refactored into `cmd/client/file/substation`
* `condition.Inspector` is no longer in the public API
* `condition` inspectors refactored into individual functions
   * JSON Schema inspector removed
   * Inspectors no longer directly support negation
* `config.Capsule` refactored into `message` package
* `config.Channel` is no longer in the public API
* `process` package refactored into `transform` package
   * Count processor removed
   * Flatten processor removed
   * IP Database processor removed
   * Processors (Transfoms) no longer directly support conditions
* `internal/transform` package removed
* `internal/sink` package removed
* `proto` removed
* Secrets are now explicitly retrieved and put into the Secrets Store using the `utility_secret` transform
* The `enrich_kv_store_set` transform had it's object.key and object.set_key behavior flipped (key is now the value put into the KV, set_key is now the key used in the KV)
* The `send_http` transform is now `send_http_post`
* All `TTLOffset` settings are now strings instead of integers (e.g., "15m")
* Removed application metrics, added `meta_metric_duration` and `utility_metric_count` in transform package
* Refactored `Transforms` method in substation package
* Moved `cmd/file/client` application to `examples/cmd/file/client`
* Renamed multiple fields based on recommendations from GitHub Copilot
* Sumo Logic support removed (replaced)
* Group processor removed (replaced)

### Features


* Added `substation` package
* Added `message` package
* Updated applications to use new concurrency and data processing model
* Added Kinesis Data Firehose support to `cmd/aws/lambda/substation`
* Added `meta_negate` inspector to `condition` package
* Added `meta_err` transform to `transform` package
* Added `meta_switch` transform to `transform` package
* Added `string_append` transform to `transform` package
* Added `string_uuid` transform to `transform` package
* Added `utility_delay` transform to `transform` package
* Added `utility_err` transform to `transform` package
* Added support for non-aggregated data to AWS Kinesis Data Stream transform
* Added region and assume role support to all AWS transforms
* Added buffering to several `send` transforms
* Removed IAM modules in `build/terraform/aws/`
* Added `build/scripts/config/format.sh`
* Added `build/scripts/terraform/format.sh`
* Added shorthand to `build/config/substation.libsonnet`
* Added `build/config/substation_test.jsonnet`
* Downgraded `go.mod` and development containers to Go 1.19
* Upgraded application containers to Go 1.21
* Refactored all `examples/`
* Added `utility_secret` transform
* All transform object handling patterns (object.key) return the input message if the retrieved key value does not exist
* Secrets Store AWS Secrets Manager backend supports AWS and retry configuration
* KV Store AWS DynamoDB backend supports AWS and retry configuration
* Added example for summarizing multiple events into a single event
* Added example for using MaxMind with the KV transform
* Added JSON array support to `meta_for_each` in condition package
* Upgraded `go.mod` to Go 1.20
* Added multi-region support to Terraform modules
* Added CloudWatch Terraform modules to collect log data
* Added (refactor) Secrets Terraform module
* Added `array_zip` transform
* IAM roles and policies in Terraform use randomized names

## Bug Fixes
* Concurrency bug in internal/aggregate package

### Code Refactoring

* Consistent Environment Variable and Application Names ([#141](https://github.com/brexhq/substation/issues/141)) ([e4062f4](https://github.com/brexhq/substation/commit/e4062f4221f0e9fcc897cda7b40a2b2d9f8aa6b2))

## [0.9.2](https://github.com/brexhq/substation/compare/v0.9.1...v0.9.2) (2023-08-10)


### Features

* add bitmath inspector ([#128](https://github.com/brexhq/substation/issues/128)) ([4721ffa](https://github.com/brexhq/substation/commit/4721ffaf7fa27aa5d343d33422cd56331ceb4d2f))
* Add JSON Lines Support to KV Store ([#126](https://github.com/brexhq/substation/issues/126)) ([667ceb3](https://github.com/brexhq/substation/commit/667ceb34f845655879e07845e1304cf01cb80e57))
* add TTLKey to KV ([#121](https://github.com/brexhq/substation/issues/121)) ([9837287](https://github.com/brexhq/substation/commit/983728745ba4366e87579a570b56c34a159009d7))
* Allow Multiple URL Interpolations in the HTTP Processor ([#124](https://github.com/brexhq/substation/issues/124)) ([f262f79](https://github.com/brexhq/substation/commit/f262f796cbd47793a72222ddcd147ca2cea6a488))


### Bug Fixes

* KV Store Processor TTL Key ([#123](https://github.com/brexhq/substation/issues/123)) ([0ceffc1](https://github.com/brexhq/substation/commit/0ceffc10c89986c0640a8b3d28776b5bc97b4811))
* use valid path to the for_each inspector settings ([#129](https://github.com/brexhq/substation/issues/129)) ([65d838d](https://github.com/brexhq/substation/commit/65d838d8ea3b12d1c27eb13d53c6b3ef49d16be4))

## [0.9.1](https://github.com/brexhq/substation/compare/v0.9.0...v0.9.1) (2023-05-09)


### Features

* Add Benchmark App & No-Op Features ([#108](https://github.com/brexhq/substation/issues/108)) ([ddfb7bc](https://github.com/brexhq/substation/commit/ddfb7bc1f4cd9699766d7673f831d976a150a1fb))
* Add Stream Transform & Streamer Interface ([#106](https://github.com/brexhq/substation/issues/106)) ([8efd82e](https://github.com/brexhq/substation/commit/8efd82ef0d5c1eb28a9d8316fe5abcea50bfa878))
* Add Zstandard & Snappy Readers ([#105](https://github.com/brexhq/substation/issues/105)) ([8c69907](https://github.com/brexhq/substation/commit/8c699070f5f6095e34e8d06e7b356ea0d4d5ed40))
* AWS DynamoDB Ingest (CDC) ([#109](https://github.com/brexhq/substation/issues/109)) ([36c60ac](https://github.com/brexhq/substation/commit/36c60ace4fa829c654c7ec86c606fc8f34ad536b))
* AWS SNS Sink ([#111](https://github.com/brexhq/substation/issues/111)) ([47e948f](https://github.com/brexhq/substation/commit/47e948f70f1a3df722aec262e0a35ef80ad492d0))

## [0.9.0](https://github.com/brexhq/substation/compare/v0.8.4...v0.9.0) (2023-04-19)


### ⚠ BREAKING CHANGES

* Add AWS AppConfig Lambda Validation app ([#92](https://github.com/brexhq/substation/issues/92))

### Features

* Add AWS AppConfig Lambda Validation app ([#92](https://github.com/brexhq/substation/issues/92)) ([f374137](https://github.com/brexhq/substation/commit/f374137066aaf2b4c1043a88f8d4ff11fb042b38))
* add gt, lt ([#98](https://github.com/brexhq/substation/issues/98)) ([110253b](https://github.com/brexhq/substation/commit/110253b646ccabc340abe5b1c9f3b66b26cc512d))
* condition inspector ([#86](https://github.com/brexhq/substation/issues/86)) ([e1fcee6](https://github.com/brexhq/substation/commit/e1fcee60ec377b994e92b5f5b3f64aa7523393ef))
* Customizable Sink Files ([#93](https://github.com/brexhq/substation/issues/93)) ([bee2463](https://github.com/brexhq/substation/commit/bee2463f2a42f7cd5834f04361fedda71db0927b))
* JQ Processor ([#88](https://github.com/brexhq/substation/issues/88)) ([0adf249](https://github.com/brexhq/substation/commit/0adf2493c6c6052fc67ee8ae62689c763d91c024))


### Bug Fixes

* decode object key ([#96](https://github.com/brexhq/substation/issues/96)) ([9e7a6db](https://github.com/brexhq/substation/commit/9e7a6db6cb1124db596e6e27bb3474ba0e16032b))

## [0.8.4](https://github.com/brexhq/substation/compare/v0.8.3...v0.8.4) (2023-03-08)


### Features

* Add Playground Demo ([#82](https://github.com/brexhq/substation/issues/82)) ([f519eaf](https://github.com/brexhq/substation/commit/f519eaff367c0f7b2cebd0ba995f247424dc4d79))
* **CLI:** adds force-sink flag ([#84](https://github.com/brexhq/substation/issues/84)) ([cb7e697](https://github.com/brexhq/substation/commit/cb7e6974993ac9116e9a564f4cdda343ae3f50a3))
* HTTP Processing & Secrets Retrieval ([#77](https://github.com/brexhq/substation/issues/77)) ([f4e7329](https://github.com/brexhq/substation/commit/f4e73296facefebfde9806d7332d2f2411604a94))
* object named groups ([#78](https://github.com/brexhq/substation/issues/78)) ([d5f687c](https://github.com/brexhq/substation/commit/d5f687c83227ec149d37224b0a360c843ae3aacf))
* setkey support ([#81](https://github.com/brexhq/substation/issues/81)) ([5419f5e](https://github.com/brexhq/substation/commit/5419f5ece82ac8dc2ef70412816bada15390da6a))

## [0.8.3](https://github.com/brexhq/substation/compare/v0.8.2...v0.8.3) (2023-01-23)


### Features

* Add MMDB Key-Value Store ([#71](https://github.com/brexhq/substation/issues/71)) ([cee1932](https://github.com/brexhq/substation/commit/cee1932cdb73d3f826361f75a5a3a4c57b01d2fa))
* Add Sync and Async AWS Lambda Ingest ([#72](https://github.com/brexhq/substation/issues/72)) ([141fdf5](https://github.com/brexhq/substation/commit/141fdf543381bd7969a16e65394194ae6042c991))


### Bug Fixes

* Aggregate & Capture Processor Options ([#75](https://github.com/brexhq/substation/issues/75)) ([46233a4](https://github.com/brexhq/substation/commit/46233a4164521b6cc30b0c6bae14f9a88d41ee1a))

## [0.8.2](https://github.com/brexhq/substation/compare/v0.8.1...v0.8.2) (2023-01-11)


### Features

* Add Sort Key Support to the AWS DynamoDB KV Store ([#68](https://github.com/brexhq/substation/issues/68)) ([517e913](https://github.com/brexhq/substation/commit/517e913ef5373e81117e6e57512f0138b2c30333))

## [0.8.1](https://github.com/brexhq/substation/compare/v0.8.0...v0.8.1) (2023-01-10)


### Features

* Add Key-Value Store Functionality ([#66](https://github.com/brexhq/substation/issues/66)) ([39b88c9](https://github.com/brexhq/substation/commit/39b88c94bb0acc0dec6994ea8b0b8076b68a8153))

## [0.8.0](https://github.com/brexhq/substation/compare/v0.7.1...v0.8.0) (2023-01-04)


### ⚠ BREAKING CHANGES

* Breaking Public APIs ([#53](https://github.com/brexhq/substation/issues/53))

### Code Refactoring

* Breaking Public APIs ([#53](https://github.com/brexhq/substation/issues/53)) ([433ec9c](https://github.com/brexhq/substation/commit/433ec9cd5821660549e0ab9d2a81d69fdc49cb1c))

## [0.7.1](https://github.com/brexhq/substation/compare/v0.7.0...v0.7.1) (2022-12-19)


### Bug Fixes

* DNS errors ([#50](https://github.com/brexhq/substation/issues/50)) ([2c9e524](https://github.com/brexhq/substation/commit/2c9e5248aa6e4e7c4c739264cd9e4a822337f076))
* IPDatabase Concurrency ([#49](https://github.com/brexhq/substation/issues/49)) ([f799a6f](https://github.com/brexhq/substation/commit/f799a6f152b2877d7136e1901e06f2fbba137121))

## [0.7.0](https://github.com/brexhq/substation/compare/v0.6.1...v0.7.0) (2022-12-13)


### Features

* DNS and IP Database Processing ([#39](https://github.com/brexhq/substation/issues/39)) ([0e43886](https://github.com/brexhq/substation/commit/0e4388681143a7bd916529116520b0f66a20aa9f))
* process.Replace allow replacing with nothing ([#42](https://github.com/brexhq/substation/issues/42)) ([7aeeb44](https://github.com/brexhq/substation/commit/7aeeb4426794484dee724ab6a4249b399b00184d))


### Bug Fixes

* ms-fontobject false positive ([#46](https://github.com/brexhq/substation/issues/46)) ([56016f2](https://github.com/brexhq/substation/commit/56016f29f58a56f4556a3f3463837b4a6696effd))
* process.IPDatabase Errors, condition.IP Type ([#44](https://github.com/brexhq/substation/issues/44)) ([a2cf347](https://github.com/brexhq/substation/commit/a2cf347d1b018b384476a7cafe44a1309463871e))

## [0.6.1](https://github.com/brexhq/substation/compare/v0.6.0...v0.6.1) (2022-12-05)


### Bug Fixes

* ForEach JSON selection ([#40](https://github.com/brexhq/substation/issues/40)) ([e1a8ae5](https://github.com/brexhq/substation/commit/e1a8ae58f98b0a8d47b578dbbe7e7bc08a089290))

## [0.6.0](https://github.com/brexhq/substation/compare/v0.5.0...v0.6.0) (2022-11-30)


### ⚠ BREAKING CHANGES

* Standardizing Use of io ([#38](https://github.com/brexhq/substation/issues/38))

### Features

* add for_each condition ([#37](https://github.com/brexhq/substation/issues/37)) ([6771180](https://github.com/brexhq/substation/commit/6771180dd1d62dfa936f43e6164aba2bf2bcf6d7))
* Add gRPC Support ([#34](https://github.com/brexhq/substation/issues/34)) ([04b4917](https://github.com/brexhq/substation/commit/04b4917f8dee59bdcec23c7a1af90bd27197beb2))


### Code Refactoring

* Standardizing Use of io ([#38](https://github.com/brexhq/substation/issues/38)) ([0368d78](https://github.com/brexhq/substation/commit/0368d782dd575d996f45b25a72cb40356c01b515))

## [0.5.0](https://github.com/brexhq/substation/compare/v0.4.0...v0.5.0) (2022-10-04)


### ⚠ BREAKING CHANGES

* Update App Concurrency Model (#30)
* Add Forward Compatibility for SNS (#21)

### Features

* Add Forward Compatibility for SNS ([#21](https://github.com/brexhq/substation/issues/21)) ([b93dc1e](https://github.com/brexhq/substation/commit/b93dc1e29b05165ed790eee201e41b2482a967c5))
* Add Initial Support for Application Metrics ([#25](https://github.com/brexhq/substation/issues/25)) ([30f103d](https://github.com/brexhq/substation/commit/30f103d44a5e7075df24a2813aa0c4d50150e276))
* AppConfig Script Updates ([#28](https://github.com/brexhq/substation/issues/28)) ([5261485](https://github.com/brexhq/substation/commit/52614853b3ebd1df587b90f0a20a8e10003d8112))
* Customizable Kinesis Data Stream Autoscaling ([#27](https://github.com/brexhq/substation/issues/27)) ([2dd7ea7](https://github.com/brexhq/substation/commit/2dd7ea74269bbaa9591d9fc50ad3ccae4102a0fd))
* Improvements to JSON Parsing ([#29](https://github.com/brexhq/substation/issues/29)) ([98cac69](https://github.com/brexhq/substation/commit/98cac69fd75a41fc464d3e269b77698f0693c638))
* Improvements to Reading and Decoding Files ([#24](https://github.com/brexhq/substation/issues/24)) ([e310cb5](https://github.com/brexhq/substation/commit/e310cb5a8e1f32e52cb695764b88d58411a94ebc))


### Bug Fixes

* **linter:** fix golangci-lint warnings across substation ([#32](https://github.com/brexhq/substation/issues/32)) ([9b7e077](https://github.com/brexhq/substation/commit/9b7e077750e12147bf456d8ecc95256fb168b0e1))
* streamname bug ([#23](https://github.com/brexhq/substation/issues/23)) ([da9de62](https://github.com/brexhq/substation/commit/da9de628d79aeaffe2192748b8e5bdb1245aed02))


### Code Refactoring

* Update App Concurrency Model ([#30](https://github.com/brexhq/substation/issues/30)) ([d8df4e2](https://github.com/brexhq/substation/commit/d8df4e2d2df63621453ea78b4156f4b8b36ba1bd))

## [0.4.0](https://github.com/brexhq/substation/compare/v0.3.0...v0.4.0) (2022-08-31)


### ⚠ BREAKING CHANGES

* Encapsulation (#15)

### Features

* Add Random Condition ([#18](https://github.com/brexhq/substation/issues/18)) ([302f24a](https://github.com/brexhq/substation/commit/302f24aae56f8f7ed8d7aee1f16ef6a335dee1a2))
* Data Aggregation ([#10](https://github.com/brexhq/substation/issues/10)) ([6cab3f7](https://github.com/brexhq/substation/commit/6cab3f75862d5a299a2aaa33d00f82c42283b895))
* Encapsulation ([#15](https://github.com/brexhq/substation/issues/15)) ([e46e780](https://github.com/brexhq/substation/commit/e46e780a1f3c0544046a41966073ce9b99e7e14f))
* PrettyPrint Processor ([#12](https://github.com/brexhq/substation/issues/12)) ([fa7a8f7](https://github.com/brexhq/substation/commit/fa7a8f7e1d7d326f65ddb95ca92fbf4e08fc2a8f))


### Bug Fixes

* Handling Large S3 Files ([#20](https://github.com/brexhq/substation/issues/20)) ([2791b91](https://github.com/brexhq/substation/commit/2791b912877bd722fea66a0bffb383552cab1400))
* Process Jsonnet Errors ([#11](https://github.com/brexhq/substation/issues/11)) ([9507c83](https://github.com/brexhq/substation/commit/9507c8324dc47a40547bb65c67baf827c422ec4c))
* replace golint with staticcheck ([#16](https://github.com/brexhq/substation/issues/16)) ([3898992](https://github.com/brexhq/substation/commit/3898992e4888e2c7d6a5c9ca0ec54eb0fa993a25))

## [0.3.0](https://github.com/brexhq/substation/compare/v0.2.0...v0.3.0) (2022-07-18)


### ⚠ BREAKING CHANGES

* Migrate to Meta Processors (#7)

### Features

* Migrate to Meta Processors ([#7](https://github.com/brexhq/substation/issues/7)) ([f0aabce](https://github.com/brexhq/substation/commit/f0aabce1e60b6be31ab3151e70b472a912741116))

## [0.2.0](https://github.com/brexhq/substation/compare/v0.1.0...v0.2.0) (2022-06-17)


### ⚠ BREAKING CHANGES

* Pre-release Refactor (#5)

### Features

* Add base64 Processor ([#4](https://github.com/brexhq/substation/issues/4)) ([cc76318](https://github.com/brexhq/substation/commit/cc7631811b59515321478918be5efaa19430649b))
* Adds Gzip Processor and Content Inspector ([#2](https://github.com/brexhq/substation/issues/2)) ([cdd2999](https://github.com/brexhq/substation/commit/cdd29999f850a77458063415dbe6b285ea3ebcc4))


### Code Refactoring

* Pre-release Refactor ([#5](https://github.com/brexhq/substation/issues/5)) ([c89ced4](https://github.com/brexhq/substation/commit/c89ced4fd1a69a23492c163471b7dcc861d0c892))
