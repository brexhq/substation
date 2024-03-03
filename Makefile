SUBSTATION_DIR ?= $(shell git rev-parse --show-toplevel 2> /dev/null)
SUBSTATION_VERSION ?= "latest"
IMAGE_ARCH ?= $(shell uname -m)
AWS_ACCOUNT_ID ?= $(shell aws sts get-caller-identity --query 'Account' --output text 2> /dev/null)
AWS_REGION ?= $(shell aws configure get region 2> /dev/null)
FONT_RED := $(shell tput setaf 1)
FONT_RESET := $(shell tput sgr0)

check:
	@printf "$(FONT_RED)>> Checking Go version...$(FONT_RESET)\n"

ifeq ($(shell go version 2> /dev/null),"")
	@echo "Go is not installed! Please install the latest version: https://go.dev/doc/install."
else
	@echo "$(shell go version 2> /dev/null)"
endif

	@printf "$(FONT_RED)>> Checking Python version...$(FONT_RESET)\n"

ifeq ($(shell python3 --version 2> /dev/null),"")
	@echo "Python is not installed! Please install the latest version: https://www.python.org/downloads."
else
	@echo "$(shell python3 --version 2> /dev/null)"
endif

	@printf "$(FONT_RED)>> Checking Terraform version...$(FONT_RESET)\n"

ifeq ($(shell terraform --version 2> /dev/null),"")
	@echo "Terraform is not installed! Please install the latest version: https://www.terraform.io/downloads.html."
else
	@echo "$(shell terraform --version 2> /dev/null)"
endif

	@printf "$(FONT_RED)>> Checking Jsonnet version...$(FONT_RESET)\n"

ifeq ($(shell jsonnet --version 2> /dev/null),"")
	@echo "Jsonnet is not installed! Please install the latest version: https://github.com/google/go-jsonnet."
else
	@echo "$(shell jsonnet --version 2> /dev/null)"
endif

	@printf "$(FONT_RED)>> Checking Docker version...$(FONT_RESET)\n"

ifeq ($(shell docker --version 2> /dev/null),"")
	@echo "Docker is not installed! Please install the latest version: https://docs.docker.com/get-docker."
else
	@echo "$(shell docker --version 2> /dev/null)"
endif

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

ifeq ("${IMAGE_ARCH}","")
	@echo "IMAGE_ARCH variable is missing!"
else
	@echo "IMAGE_ARCH: ${IMAGE_ARCH}"	
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

ifeq ("${AWS_APPCONFIG_ENV}","")
	@echo "AWS_APPCONFIG_ENV variable is missing! This must match the AppConfig environment defined in Terraform."
else
	@echo "AWS_APPCONFIG_ENV: ${AWS_APPCONFIG_ENV}"
endif

	@printf "$(FONT_RED)>> Checking deployment variables...$(FONT_RESET)\n"

ifeq ("${DEPLOYMENT_DIR}","")
	@echo "DEPLOYMENT_DIR variable is missing! This must point to a directory containing configuration (Jsonnet) and infrastructure (Terraform) files."
else
	@echo "DEPLOYMENT_DIR: ${DEPLOYMENT_DIR}"
endif

build-config:
	@printf "$(FONT_RED)>> Building configuration files...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && bash build/scripts/config/compile.sh

build-go:
	@printf "$(FONT_RED)>> Building Go binaries...$(FONT_RESET)\n"
	@for file in $(shell find $(SUBSTATION_DIR) -name main.go); do \
		cd $$(dirname $$file) && go build; \
	done

build-aws:
	@$(MAKE) build-aws-appconfig
	@$(MAKE) build-aws-substation
	@$(MAKE) build-aws-validate
	@$(MAKE) build-aws-autoscale
	@$(MAKE) build-config

build-aws-appconfig:
	@printf "$(FONT_RED)>> Building AppConfig extension...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && AWS_ARCHITECTURE=$(IMAGE_ARCH) AWS_REGION=$(AWS_REGION) bash build/scripts/aws/lambda/get_appconfig_extension.sh

build-aws-substation:
	@printf "$(FONT_RED)>> Building Substation Lambda function...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && docker build --build-arg ARCH=$(IMAGE_ARCH) -f build/container/aws/lambda/substation/Dockerfile -t substation:latest-$(IMAGE_ARCH) .

build-aws-validate:
	@printf "$(FONT_RED)>> Building Validator Lambda function...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && docker build --build-arg ARCH=$(IMAGE_ARCH) -f build/container/aws/lambda/validate/Dockerfile -t validate:latest-$(IMAGE_ARCH) .

build-aws-autoscale:
	@printf "$(FONT_RED)>> Building Autoscaler Lambda function...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && docker build --build-arg ARCH=$(IMAGE_ARCH) -f build/container/aws/lambda/autoscale/Dockerfile -t autoscale:latest-$(IMAGE_ARCH) .

deploy-aws:
	@$(MAKE) deploy-aws-tf-init
	@$(MAKE) deploy-aws-images
	@$(MAKE) deploy-aws-tf-all
	@$(MAKE) deploy-aws-config

deploy-aws-tf-init:
	@printf "$(FONT_RED)>> Initializing cloud infrastructure to AWS with Terraform...$(FONT_RESET)\n"

	@cd $$(dirname $$(find $(DEPLOYMENT_DIR) -name _provider.tf)) && \
	terraform init && \
	terraform apply \
	-target=module.kms \
	-target=module.ecr \
	-target=module.ecr_autoscale \
	-target=module.ecr_validate

deploy-aws-tf-all:
	@printf "$(FONT_RED)>> Deploying cloud infrastructure to AWS with Terraform...$(FONT_RESET)\n"
	@cd $$(dirname $$(find $(DEPLOYMENT_DIR) -name _provider.tf)) && terraform apply

deploy-aws-images:
	@printf "$(FONT_RED)>> Deploying images to AWS ECR with Docker...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && AWS_ACCOUNT_ID=$(AWS_ACCOUNT_ID) AWS_REGION=$(AWS_REGION) bash build/scripts/aws/ecr_login.sh

	@cd $$(dirname $$(find $(DEPLOYMENT_DIR) -name _provider.tf))
ifneq ("$(shell aws ecr describe-repositories --region $(AWS_REGION) --repository-names substation --output text 2> /dev/null)","")
	@docker tag substation:latest-$(IMAGE_ARCH) $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/substation:$(SUBSTATION_VERSION)
	@docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/substation:$(SUBSTATION_VERSION)
endif

ifneq ("$(shell aws ecr describe-repositories --repository-names validate --output text 2> /dev/null)","")
	@docker tag validate:latest-$(IMAGE_ARCH) $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/validate:$(SUBSTATION_VERSION)
	@docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/validate:$(SUBSTATION_VERSION)
endif

ifneq ("$(shell aws ecr describe-repositories --repository-names autoscale --output text 2> /dev/null)","")
	@docker tag autoscale:latest-$(IMAGE_ARCH) $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/autoscale:$(SUBSTATION_VERSION)
	@docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/autoscale:$(SUBSTATION_VERSION)
endif

deploy-aws-config:
	@printf "$(FONT_RED)>> Deploying configurations to AppConfig with Python...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR) && SUBSTATION_CONFIG_DIRECTORY=$(DEPLOYMENT_DIR) AWS_DEFAULT_REGION=$(AWS_REGION)  AWS_APPCONFIG_APPLICATION_NAME=substation AWS_APPCONFIG_ENVIRONMENT=$(AWS_APPCONFIG_ENV) AWS_APPCONFIG_DEPLOYMENT_STRATEGY=Instant python3 build/scripts/aws/appconfig/appconfig_upload.py

destroy-aws:
	@printf "$(FONT_RED)>> Destroying configurations in AppConfig with Python...$(FONT_RESET)\n"
	@cd $(SUBSTATION_DIR)

	@for file in $(shell find $(DEPLOYMENT_DIR) -name config.jsonnet); do \
		AWS_DEFAULT_REGION=$(AWS_REGION) AWS_APPCONFIG_APPLICATION_NAME=substation AWS_APPCONFIG_PROFILE_NAME=$$(basename $$(dirname $$file)) python3 build/scripts/aws/appconfig/appconfig_delete.py; \
	done

	@printf "$(FONT_RED)>> Destroying cloud infrastructure in AWS with Terraform...$(FONT_RESET)\n"
	@cd $$(dirname $$(find $(DEPLOYMENT_DIR) -name _provider.tf)) && terraform destroy

test-local:
	@$(MAKE) build-go
	@$(MAKE) build-config
	
	@printf "$(FONT_RED)>> Printing data file...$(FONT_RESET)\n"
	@cat examples/cmd/client/file/substation/data.json 

	@printf "$(FONT_RED)>> Running Substation...$(FONT_RESET)\n"
	@cd examples/cmd/client/file/substation && ./substation -config config.json -file data.json

test-aws:
	@$(MAKE) build-aws
	@$(MAKE) deploy-aws DEPLOYMENT_DIR=examples/terraform/aws/kinesis/time_travel AWS_APPCONFIG_ENV=example

	@printf "$(FONT_RED)>> Sending data to AWS Kinesis...$(FONT_RESET)\n"
	@cd build/scripts/aws/kinesis && \
	echo '{"ip":"8.8.8.8"}' >> data.json && \
	echo '{"ip":"9.9.9.9"}' >> data.json && \
	echo '{"ip":"1.1.1.1"}' >> data.json && \
	echo '{"ip":"8.8.8.8","context":"GOOGLE"}' >> data.json && \
	echo '{"ip":"9.9.9.9","context":"QUAD9"}' >> data.json && \
	echo  '{"ip":"1.1.1.1","context":"CLOUDFLARENET"}' >> data.json && \
	AWS_DEFAULT_REGION=$(AWS_REGION) python3 put_records.py substation data.json --print-response && \
	rm data.json

	@printf "$(FONT_RED)>> Log in to the AWS console and review the logs written by the Substation Lambda functions. When finished, run 'make destroy-aws' to clean up the resources. $(FONT_RESET)\n"
