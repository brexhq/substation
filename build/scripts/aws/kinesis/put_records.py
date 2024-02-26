"""
Puts records into a Kinesis Data Stream.

Usage example:
    python3 put_records.py my-stream my-file.txt --print-response
"""


import argparse
import boto3
import uuid

CLIENT = boto3.client("kinesis")


def main():
    parser = argparse.ArgumentParser(
        description="Puts records into a Kinesis Data Stream"
    )
    parser.add_argument("stream_name", help="The name of the stream")
    parser.add_argument(
        "file", help="The file containing data that is put into the stream"
    )
    parser.add_argument(
        "--print-response",
        help="Determines if the response is printed to the console",
        action="store_true",
    )
    args = parser.parse_args()

    with open(args.file, "rb") as f:
        for line in f.readlines():
            data = line.decode("utf-8").strip()
            resp = CLIENT.put_record(
                StreamName=args.stream_name, Data=data, PartitionKey=str(uuid.uuid4())
            )

            if args.print_response:
                print(resp)


if __name__ == "__main__":
    main()
