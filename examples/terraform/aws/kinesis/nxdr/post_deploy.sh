AWS_DEFAULT_REGION=$AWS_REGION python3 ../build/scripts/aws/kinesis/put_records.py substation terraform/aws/kinesis/nxdr/data.jsonl  --print-response
