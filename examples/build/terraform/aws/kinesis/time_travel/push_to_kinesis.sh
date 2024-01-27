# Sends example event data to the API Gateway endpoint. The events containing
# context are sent after the events without context to demonstrate the pipeline's
# ability to enrich events in real-time using "time travel".

# This is a placeholder that must be replaced with the API Gateway endpoint produced by Terraform.
url=https://9s6fewf1kg.execute-api.us-east-1.amazonaws.com/gateway

curl -X POST -H "Content-Type: application/json" -d '{"ip":"8.8.8.8"}' $url; echo &
curl -X POST -H "Content-Type: application/json" -d '{"ip":"9.9.9.9"}' $url; echo &
curl -X POST -H "Content-Type: application/json" -d '{"ip":"1.1.1.1"}' $url; echo &
curl -X POST -H "Content-Type: application/json" -d '{"ip":"8.8.8.8","context":"GOOGLE"}' $url; echo &
curl -X POST -H "Content-Type: application/json" -d '{"ip":"9.9.9.9","context":"QUAD9"}' $url; echo &
curl -X POST -H "Content-Type: application/json" -d '{"ip":"1.1.1.1","context":"CLOUDFLARENET"}' $url; echo &

wait
