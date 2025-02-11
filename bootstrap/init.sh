#!/bin/bash

USER_DATA_BASE64=$(curl -H Metadata:true --noproxy "*" "http://169.254.169.254/metadata/instance/compute/userData?api-version=2021-01-01&format=text")
USER_DATA_JSON=$(echo "$USER_DATA_BASE64" | base64 --decode)

echo -n "$USER_DATA_JSON" | sha1sum | awk '{print $1}' > /root/aapp-toolkit/manifest.sha1

DNS_ROOT=$(echo "$USER_DATA_JSON" | jq -r '.spec.ingress.hostname')

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

CSR_BASE64=$(base64 -w 0 "$CSR_FILE")

cd azure-attestation/app

bash pre-requisites.sh
cmake .
make

# sudo ./AttestationClient -c "$CSR_BASE64" -m "$USER_DATA_BASE64" -o token

cd ../../bootstrap

cmake .
make

echo $CSR_BASE64 > csr.txt
 ./server

awk '
  /-----BEGIN CERTIFICATE-----/ {flag=1}
  flag {print}
  /-----END CERTIFICATE-----/ {flag=0; exit}
' fullchain.pem | openssl x509 -noout -fingerprint -sha1 > /root/aapp-toolkit/cert.sha1

cd ../

apt-get update -yq
apt-get install -yq nginx
apt-get install -yq libnginx-mod-http-lua

NGINX_CONFIG="/etc/nginx/sites-available/default"

cp -f ingress/default.conf $NGINX_CONFIG

AAPPPORT=$(echo "$USER_DATA_JSON" | jq -r '.spec.ingress.port')
AAPPREPO=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.repo')
AAPPCOMMITSHA=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.tag')
AAPPJOB=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.job')

sed -i "s|__AAPPHOSTNAME__|${DNS_ROOT}|g" "$NGINX_CONFIG"
sed -i "s|__AAPPPORT__|${AAPPPORT}|g" "$NGINX_CONFIG"
sed -i "s|__AAPPREPO__|${AAPPREPO%.git}|g" "$NGINX_CONFIG"
sed -i "s|__AAPPCOMMITSHA__|${AAPPCOMMITSHA}|g" "$NGINX_CONFIG"
sed -i "s|__AAPPJOB__|${AAPPJOB}|g" "$NGINX_CONFIG"

cp azure-attestation/scripts/token.sh /var/www/html/
chmod +x /var/www/html/token.sh
cp azure-attestation/web/index.html /var/www/html/
chown www-data:www-data /var/www/html/*

# Define the file that will be created in /etc/sudoers.d
SUDOERS_FILE="/etc/sudoers.d/www-data-attestation"

# Define the command for which www-data will have passwordless sudo access
COMMAND="/root/aapp-toolkit/azure-attestation/app/AttestationClient"

# Create the sudoers entry
echo "www-data ALL=(ALL) NOPASSWD: ${COMMAND}" > "$SUDOERS_FILE"

# Set proper permissions for the sudoers file
chmod 0440 "$SUDOERS_FILE"

service nginx restart