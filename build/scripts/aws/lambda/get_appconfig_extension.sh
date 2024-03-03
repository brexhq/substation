#!/bin/bash
# downloads AWS Lambda AppConfig extension based on architecture and region
set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z $AWS_ARCHITECTURE ]; then
  >&2 echo "Error: AWS_ARCHITECTURE not set."
  exit 1
fi

if [ -z $AWS_REGION ]; then
  >&2 echo "Error: AWS_REGION not set."
  exit 1
fi

if [ $AWS_ARCHITECTURE == "arm64" ]; then
	case $AWS_REGION in
		us-east-1)
			AWS_ARN="arn:aws:lambda:us-east-1:027255383542:layer:AWS-AppConfig-Extension-Arm64:1"
			;;
		us-east-2)
			AWS_ARN="arn:aws:lambda:us-east-2:728743619870:layer:AWS-AppConfig-Extension-Arm64:1"
			;;
		us-west-2)
			AWS_ARN="arn:aws:lambda:us-west-2:359756378197:layer:AWS-AppConfig-Extension-Arm64:2"
			;;
		eu-central-1)
			AWS_ARN="arn:aws:lambda:eu-central-1:066940009817:layer:AWS-AppConfig-Extension-Arm64:1"
			;;
		eu-west-1)
			AWS_ARN="arn:aws:lambda:eu-west-1:434848589818:layer:AWS-AppConfig-Extension-Arm64:6"
			;;
		eu-west-2)
			AWS_ARN="arn:aws:lambda:eu-west-2:282860088358:layer:AWS-AppConfig-Extension-Arm64:1"
			;;
		ap-northeast-1)
			AWS_ARN="arn:aws:lambda:ap-northeast-1:980059726660:layer:AWS-AppConfig-Extension-Arm64:1"
			;;
		ap-southeast-1)
			AWS_ARN="arn:aws:lambda:ap-southeast-1:421114256042:layer:AWS-AppConfig-Extension-Arm64:2"
			;;
		ap-southeast-2)
			AWS_ARN="arn:aws:lambda:ap-southeast-2:080788657173:layer:AWS-AppConfig-Extension-Arm64:1"
			;;
		ap-south-1)
			AWS_ARN="arn:aws:lambda:ap-south-1:554480029851:layer:AWS-AppConfig-Extension-Arm64:1"
			;;
	esac
elif [ $AWS_ARCHITECTURE == "x86_64" ]; then
	case $AWS_REGION in 
		us-east-1)
			AWS_ARN="arn:aws:lambda:us-east-1:027255383542:layer:AWS-AppConfig-Extension:61"
			;;
		us-east-2)
			AWS_ARN="arn:aws:lambda:us-east-2:728743619870:layer:AWS-AppConfig-Extension:47"
			;;
		us-west-1)
			AWS_ARN="arn:aws:lambda:us-west-1:958113053741:layer:AWS-AppConfig-Extension:61"
			;;
		us-west-2)
			AWS_ARN="arn:aws:lambda:us-west-2:359756378197:layer:AWS-AppConfig-Extension:89"
			;;
		ca-central-1)
			AWS_ARN="arn:aws:lambda:ca-central-1:039592058896:layer:AWS-AppConfig-Extension:47"
			;;
		eu-central-1)
			AWS_ARN="arn:aws:lambda:eu-central-1:066940009817:layer:AWS-AppConfig-Extension:54"
			;;
		eu-west-1)
			AWS_ARN="arn:aws:lambda:eu-west-1:434848589818:layer:AWS-AppConfig-Extension:59"
			;;
		eu-west-2)
			AWS_ARN="arn:aws:lambda:eu-west-2:282860088358:layer:AWS-AppConfig-Extension:47"
			;;
		eu-west-3)
			AWS_ARN="arn:aws:lambda:eu-west-3:493207061005:layer:AWS-AppConfig-Extension:48"
			;;
		eu-north-1)
			AWS_ARN="arn:aws:lambda:eu-north-1:646970417810:layer:AWS-AppConfig-Extension:86"
			;;
		eu-south-1)
			AWS_ARN="arn:aws:lambda:eu-south-1:203683718741:layer:AWS-AppConfig-Extension:44"
			;;
		cn-north-1)
			AWS_ARN="arn:aws-cn:lambda:cn-north-1:615057806174:layer:AWS-AppConfig-Extension:43"
			;;
		cn-northwest-1)
			AWS_ARN="arn:aws-cn:lambda:cn-northwest-1:615084187847:layer:AWS-AppConfig-Extension:43"
			;;
		ap-east-1)
			AWS_ARN="arn:aws:lambda:ap-east-1:630222743974:layer:AWS-AppConfig-Extension:44"
			;;
		ap-northeast-1)
			AWS_ARN="arn:aws:lambda:ap-northeast-1:980059726660:layer:AWS-AppConfig-Extension:45"
			;;
		ap-northeast-2)
			AWS_ARN="arn:aws:lambda:ap-northeast-2:826293736237:layer:AWS-AppConfig-Extension:54"
			;;
		ap-northeast-3)
			AWS_ARN="arn:aws:lambda:ap-northeast-3:706869817123:layer:AWS-AppConfig-Extension:42"
			;;
		ap-southeast-1)
			AWS_ARN="arn:aws:lambda:ap-southeast-1:421114256042:layer:AWS-AppConfig-Extension:45"
			;;
		ap-southeast-2)
			AWS_ARN="arn:aws:lambda:ap-southeast-2:080788657173:layer:AWS-AppConfig-Extension:54"
			;;
		ap-southeast-3)
			AWS_ARN="arn:aws:lambda:ap-southeast-3:418787028745:layer:AWS-AppConfig-Extension:13"
			;;
		ap-south-1)
			AWS_ARN="arn:aws:lambda:ap-south-1:554480029851:layer:AWS-AppConfig-Extension:55"
			;;
		sa-east-1)
			AWS_ARN="arn:aws:lambda:sa-east-1:000010852771:layer:AWS-AppConfig-Extension:61"
			;;
		af-south-1)
			AWS_ARN="arn:aws:lambda:af-south-1:574348263942:layer:AWS-AppConfig-Extension:44"
			;;
		me-south-1)
			AWS_ARN="arn:aws:lambda:me-south-1:559955524753:layer:AWS-AppConfig-Extension:44"
			;;
		us-gov-east-1)
			AWS_ARN="arn:aws-us-gov:lambda:us-gov-east-1:946561847325:layer:AWS-AppConfig-Extension:20"
			;;
		us-gov-west-1)
			AWS_ARN="arn:aws-us-gov:lambda:us-gov-west-1:946746059096:layer:AWS-AppConfig-Extension:20"
			;;
	esac
else
	echo "Error: invalid AWS_ARCHITECTURE."
	exit 1
fi

aws lambda get-layer-version-by-arn \
  --region $AWS_REGION --arn $AWS_ARN \
  | jq -r '.Content.Location' \
  | xargs curl -s -o $DIR/extension.zip 
