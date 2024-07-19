"""
Bulk delete items from a DynamoDB table.

The file should contain a list of items, in JSON Lines format, to delete
and each item must match the schema of the table. For example, if the 
primary key of the table is "id" then the file should contain items like:
    {"id": "1"}
    {"id": "2"}
    {"id": "3"}

Example usage:
    python3 bulk_delete_items.py my-table my-file.jsonl
"""

import argparse
import boto3
import json

DDB = boto3.resource("dynamodb")


def main():
    parser = argparse.ArgumentParser(
        description="Bulk delete items from a DynamoDB table."
    )
    parser.add_argument("table_name")
    parser.add_argument("file")
    args = parser.parse_args()

    t = DDB.Table(args.table_name)
    with open(args.file, "rb") as f, t.batch_writer() as batch:
        for item in f.readlines():
            item = item.decode("utf-8").strip()
            batch.delete_item(Key=json.loads(item))


if __name__ == "__main__":
    main()
