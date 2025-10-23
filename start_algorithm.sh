#!/bin/bash

set -e

if [ $# -ne 2 ]; then
    echo "Usage: $0 <node_address> <machine_id>"
    echo "Example: $0 localhost:8080 node-1"
    exit 1
fi

NODE_ADDRESS="$1"
MACHINE_ID="$2"

ENVELOPE=$(cat <<EOF
{
  "type": "StartAlgorithmRequest",
  "to": {
    "machine_id": "$MACHINE_ID",
    "actor_id": "coordinator"
  },
  "payload": {}
}
EOF
)

echo "Sending StartAlgorithmRequest to $NODE_ADDRESS..."
echo "Target PID: $MACHINE_ID/coordinator"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "$ENVELOPE" \
  "http://$NODE_ADDRESS/message")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "Message sent successfully!"
else
    echo "Failed to send message"
    echo "HTTP Status: $HTTP_CODE"
    exit 1
fi
