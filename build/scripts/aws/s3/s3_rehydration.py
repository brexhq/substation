"""
Rehydrates data from an AWS S3 bucket into an SNS topic by simulating S3 
object creation events.

Typical usage example:

    python3 s3_rehydration.py --bucket my-bucket --topic my-topic 
    --prefix my-prefix --filter my-filter my-other-filter
"""

import argparse
import boto3
import json
import logging
import os
import time

from botocore.exceptions import ClientError

logging.getLogger().setLevel(logging.INFO)

S3 = boto3.client("s3")
SNS = boto3.client("sns")


def main():
    args = argparse.ArgumentParser(
        description="""Rehydrates data from an AWS S3 bucket into an SNS topic by 
        simulating S3 object creation events. If no --prefix and --filter are specified, all objects in the bucket are rehydrated.""",
        add_help=True,
    )
    args.add_argument("--bucket", required=True, help="S3 bucket name")
    args.add_argument("--topic", required=True, help="SNS topic ARN")
    args.add_argument("--prefix", required=False, help="S3 prefix")
    args.add_argument(
        "--filter",
        default=[],
        nargs="+",
        required=False,
        help="filter S3 object keys using substrings",
    )
    args = args.parse_args()

    try:
        S3.head_bucket(Bucket=args.bucket)
    except ClientError as e:
        logging.exception(f'bucket "{args.bucket}" not found')

    try:
        SNS.get_topic_attributes(TopicArn=args.topic)
    except ClientError as e:
        logging.exception(f'topic "{args.topic}" not found')

    continuation_token = None
    while 1:
        objects = S3.list_objects_v2(
            Bucket=args.bucket,
            Prefix=args.prefix,
            MaxKeys=1000,
            EncodingType="url",
        )
        if continuation_token:
            objects["NextContinuationToken"] = continuation_token

        count = 0
        for o in objects.get("Contents", []):
            if not args.filter or all(x in o.get("Key") for x in args.filter):
                event = {
                    "Records": [
                        {
                            "eventVersion": "2.2",
                            "eventSource": "aws:s3",
                            "awsRegion": os.environ.get("AWS_REGION"),
                            "eventTime": time.strftime(
                                "%Y-%m-%dT%H:%M:%SZ", time.gmtime()
                            ),
                            "eventName": "ObjectCreated:*",
                            "userIdentity": {
                                "principalId": os.environ.get("AWS_ACCOUNT_ID"),
                            },
                            "s3": {
                                "s3SchemaVersion": "1.0",
                                "configurationId": "substation_s3_rehydrate",
                                "bucket": {
                                    "name": args.bucket,
                                    "arn": f"arn:aws:s3:::{args.bucket}",
                                },
                                "object": {
                                    "key": o.get("Key"),
                                    "size": o.get("Size"),
                                    "eTag": o.get("ETag"),
                                },
                            },
                        }
                    ]
                }

                SNS.publish(TopicArn=args.topic, Message=json.dumps(event))
                count += 1

        logging.info(f"rehydrated {count} object(s)")

        continuation_token = objects.get("NextContinuationToken")
        if not continuation_token:
            break


if __name__ == "__main__":
    main()
