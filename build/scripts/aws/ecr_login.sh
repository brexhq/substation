#!/bin/bash
set -euo pipefail

if [ -z $AWS_PROFILE ]; then
  >&2 echo "Error: AWS_PROFILE not set."
  exit 1
fi

if [ -z $AWS_REGION ]; then
  >&2 echo "Error: AWS_REGION not set."
  exit 1
fi

if [ -z $AWS_ACCOUNT_ID ]; then
  >&2 echo "Error: AWS_ACCOUNT_ID not set."
  exit 1
fi

if ! [ -x "$(command -v aws)" ]; then
  >&2 echo "Error: AWS CLI is not installed."
  exit 1
fi

if ! [ -x "$(command -v docker)" ]; then
  >&2 echo "Error: Docker is not installed."
  exit 1
fi

REGISTRY="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"
aws ecr get-login-password | \
  docker login --username AWS --password-stdin "$REGISTRY"
