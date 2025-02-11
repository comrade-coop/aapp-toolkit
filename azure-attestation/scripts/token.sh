#!/bin/bash
# Ensure the nonce parameter is captured properly.
nonce="$1"

# Run the AttestationClient command and capture its output.
token=$(sudo /root/aapp-toolkit/azure-attestation/app/AttestationClient \
  -o token \
  -n "$nonce" \
  -c $(cat /root/aapp-toolkit/cert.sha1) \
  -m $(cat /root/aapp-toolkit/manifest.sha1) 2>&1)

# Optionally, debug by printing the token value:
# echo "DEBUG: token is: $token" >&2

# Output the JSON object.
echo "{\"token\": \"$token\"}"