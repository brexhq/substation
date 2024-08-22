sleep 5
AWS_DEFAULT_REGION=$AWS_REGION python3 ../build/scripts/aws/kinesis/put_records.py substation terraform/aws/kinesis/time_travel/data.jsonl  --print-response
