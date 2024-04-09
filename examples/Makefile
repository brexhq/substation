SUBSTATION_DIR ?= $(shell git rev-parse --show-toplevel 2> /dev/null)
SUBSTATION_VERSION ?= $(shell git describe --tags --always 2> /dev/null)
AWS_ACCOUNT_ID ?= $(shell aws sts get-caller-identity --query 'Account' --output text 2> /dev/null)
AWS_REGION ?= $(shell aws configure get region 2> /dev/null)

FONT_RED := $(shell tput setaf 1)
FONT_RESET := $(shell tput sgr0)

check:
	@printf "$(FONT_RED)>> Checking Substation variables...$(FONT_RESET)\n"

ifeq ("${SUBSTATION_DIR}","")
	@echo "SUBSTATION_DIR variable is missing!"
else
	@echo "SUBSTATION_DIR: ${SUBSTATION_DIR}"
endif

ifeq ("${SUBSTATION_VERSION}","")
	@echo "SUBSTATION_VERSION variable is missing!"
else
	@echo "SUBSTATION_VERSION: ${SUBSTATION_VERSION}"
endif

	@printf "$(FONT_RED)>> Checking AWS variables...$(FONT_RESET)\n"

ifeq ("${AWS_ACCOUNT_ID}","")
	@echo "AWS_ACCOUNT_ID variable is missing!"
else
	@echo "AWS_ACCOUNT_ID: ${AWS_ACCOUNT_ID}"
endif

ifeq ("${AWS_REGION}","")
	@echo "AWS_REGION variable is missing!"
else
	@echo "AWS_REGION: ${AWS_REGION}"
endif

.PHONY: build
build:
	@$(MAKE) build-go
	@$(MAKE) build-config
	@$(MAKE) build-images

build-config:
	@printf "$(FONT_RED)>> Building configuration files...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && bash build/scripts/config/compile.sh

build-go:
	@printf "$(FONT_RED)>> Building Go binaries...$(FONT_RESET)\n"
	@for file in $(shell find $(SUBSTATION_DIR) -name main.go); do \
		cd $$(dirname $$file) && go build; \
	done

build-images:
	@printf "$(FONT_RED)>> Building AppConfig extension...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && AWS_ARCHITECTURE=arm64 AWS_REGION=$(AWS_REGION) bash build/scripts/aws/lambda/get_appconfig_extension.sh

	@printf "$(FONT_RED)>> Building Docker images...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && \
		docker buildx build --platform linux/arm64 --build-arg ARCH=arm64 -f build/container/aws/lambda/substation/Dockerfile -t substation:latest-arm64 . && \
		docker buildx build --platform linux/arm64 --build-arg ARCH=arm64 -f build/container/aws/lambda/autoscale/Dockerfile -t autoscale:latest-arm64 . && \
		docker buildx build --platform linux/arm64 --build-arg ARCH=arm64 -f build/container/aws/lambda/validate/Dockerfile -t validate:latest-arm64 .

deploy:
	@$(MAKE) deploy-aws-init
	@$(MAKE) deploy-aws-images
	@$(MAKE) deploy-aws-infra
	@$(MAKE) deploy-aws-config
	@$(MAKE) deploy-aws-post-script

deploy-aws-init:
	@printf "$(FONT_RED)>> Initializing cloud infrastructure in AWS with Terraform...$(FONT_RESET)\n"

	@cd $(EXAMPLE)/terraform && \
	terraform init && \
	terraform apply -auto-approve -compact-warnings \
	-target=module.kms \
	-target=module.ecr \
	-target=module.ecr_autoscale \
	-target=module.ecr_validate

deploy-aws-images:
	@printf "$(FONT_RED)>> Deploying images to AWS ECR with Docker...$(FONT_RESET)\n"

	@aws ecr get-login-password | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

ifneq ("$(shell aws ecr describe-repositories --repository-names substation --output text 2> /dev/null)","")
	@docker tag substation:latest-arm64 $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/substation:$(SUBSTATION_VERSION)
	@docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/substation:$(SUBSTATION_VERSION)
endif

ifneq ("$(shell aws ecr describe-repositories --repository-names validate --output text 2> /dev/null)","")
	@docker tag validate:latest-arm64 $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/validate:$(SUBSTATION_VERSION)
	@docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/validate:$(SUBSTATION_VERSION)
endif

ifneq ("$(shell aws ecr describe-repositories --repository-names autoscale --output text 2> /dev/null)","")
	@docker tag autoscale:latest-arm64 $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/autoscale:$(SUBSTATION_VERSION)
	@docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/autoscale:$(SUBSTATION_VERSION)
endif

deploy-aws-infra:
	@printf "$(FONT_RED)>> Deploying cloud infrastructure in AWS with Terraform...$(FONT_RESET)\n"
	@cd $(EXAMPLE)/terraform && terraform apply -auto-approve -compact-warnings

deploy-aws-config:
	@printf "$(FONT_RED)>> Deploying configurations to AppConfig with Python...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && SUBSTATION_CONFIG_DIRECTORY=examples/$(EXAMPLE) AWS_APPCONFIG_APPLICATION_NAME=substation AWS_APPCONFIG_ENVIRONMENT=example AWS_APPCONFIG_DEPLOYMENT_STRATEGY=Instant python3 build/scripts/aws/appconfig/appconfig_upload.py

deploy-aws-post-script:
ifneq ("$(wildcard $(EXAMPLE)/post_deploy.sh)","")
	@printf "$(FONT_RED)>> Running post-deploy script...$(FONT_RESET)\n"
	@bash $(EXAMPLE)/post_deploy.sh
endif

destroy:
	@printf "$(FONT_RED)>> Destroying configurations in AppConfig with Python...$(FONT_RESET)\n"
	@for file in $(shell find $(EXAMPLE) -name config.jsonnet); do \
		AWS_DEFAULT_REGION=$(AWS_REGION) AWS_APPCONFIG_APPLICATION_NAME=substation AWS_APPCONFIG_PROFILE_NAME=$$(basename $$(dirname $$file)) python3 $(SUBSTATION_DIR)/build/scripts/aws/appconfig/appconfig_delete.py; \
	done

	@printf "$(FONT_RED)>> Destroying cloud infrastructure in AWS with Terraform...$(FONT_RESET)\n"
	@cd $(EXAMPLE)/terraform && terraform destroy -auto-approve -compact-warnings

quickstart:
	@$(MAKE) build-go
	@$(MAKE) build-config
	
	@printf "$(FONT_RED)>> Printing data file...$(FONT_RESET)\n"
	@cat cmd/client/file/substation/data.json 

	@printf "$(FONT_RED)>> Running Substation...$(FONT_RESET)\n"
	@cd cmd/client/file/substation && ./substation -config config.json -file data.json
