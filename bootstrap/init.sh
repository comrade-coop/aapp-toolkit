#!/bin/bash

# Step 1: Download Base64 encoded YAML file
USER_DATA_BASE64=$(curl -H Metadata:true --noproxy "*" "http://169.254.169.254/metadata/instance/compute/userData?api-version=2021-01-01&format=text")

# Step 2: Decode the YAML file
USER_DATA_YAML=$(echo "$USER_DATA_BASE64" | base64 --decode)

# Step 3: Extract the value of spec.dns.root using yq (install it if necessary)
DNS_ROOT=$(echo "$USER_DATA_YAML" | yq eval '.spec.dns.root' -)

# Step 4: Create CSR for the extracted domain
OPENSSL_CONF="openssl.cnf"
CSR_FILE="${DNS_ROOT}.csr"
KEY_FILE="${DNS_ROOT}.key"

cat > "$OPENSSL_CONF" <<EOF
[ req ]
default_bits       = 2048
default_keyfile    = $KEY_FILE
distinguished_name = req_distinguished_name
req_extensions     = req_ext

[ req_distinguished_name ]
CN = $DNS_ROOT

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = $DNS_ROOT
EOF

openssl req -new -newkey rsa:2048 -nodes -keyout "$KEY_FILE" -out "$CSR_FILE" -config "$OPENSSL_CONF" -subj "/CN=$DNS_ROOT"

# Step 5: Base64 encode CSR and invoke AttestationClient
CSR_BASE64=$(base64 -w 0 "$CSR_FILE")

cd ../azure-attestation/app

sudo bash pre-requisites.sh
cmake .
make

sudo ./AttestationClient -c "$CSR_BASE64" -m "$USER_DATA_BASE64" -o token