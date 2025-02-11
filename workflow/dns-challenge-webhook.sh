#!/bin/bash
# dns-challenge-webhook.sh

# Duplicate stdout and stderr to HOOK_LOG.
exec > >(tee -a "$HOOK_LOG") 2> >(tee -a "$HOOK_LOG" >&2)

# The DNS challenge details are passed via Certbot environment variables.
DNS_RECORD="_acme-challenge.${CERTBOT_DOMAIN}"
EXPECTED="${CERTBOT_VALIDATION}"

echo "=================================================================="
echo "MANUAL DNS CHALLENGE REQUIRED FOR DOMAIN: ${CERTBOT_DOMAIN}"
echo "------------------------------------------------------------------"
echo "Please create the following DNS TXT record at your DNS provider:"
echo "Record: ${DNS_RECORD}"
echo "Value: ${EXPECTED}"
echo "------------------------------------------------------------------"
echo "The script will now poll DNS for the TXT record until it appears with the expected value."
echo "=================================================================="

# Configuration: number of attempts and sleep interval (in seconds).
MAX_ATTEMPTS=30
SLEEP_TIME=10

attempt=0
while [ $attempt -lt $MAX_ATTEMPTS ]; do
    txt_record=$(dig +short TXT "${DNS_RECORD}" | tr -d '"')
    echo "Attempt $((attempt + 1)): Found TXT record: ${txt_record}"

    if [[ "${txt_record}" == *"${EXPECTED}"* ]]; then
        echo "Expected DNS record found. Continuing..."
        exit 0
    fi

    attempt=$((attempt + 1))
    sleep $SLEEP_TIME
done

echo "ERROR: DNS record with the expected value was not found after $((MAX_ATTEMPTS * SLEEP_TIME)) seconds."
exit 1
