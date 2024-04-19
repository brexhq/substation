sleep 5
AWS_DEFAULT_REGION=$AWS_REGION python3 ../build/scripts/aws/kinesis/put_records.py substation_edr terraform/aws/dynamodb/telephone/edr_data.jsonl  --print-response
AWS_DEFAULT_REGION=$AWS_REGION python3 ../build/scripts/aws/kinesis/put_records.py substation_idp terraform/aws/dynamodb/telephone/idp_data.jsonl  --print-response
AWS_DEFAULT_REGION=$AWS_REGION python3 ../build/scripts/aws/kinesis/put_records.py substation_md terraform/aws/dynamodb/telephone/md_data.jsonl  --print-response
