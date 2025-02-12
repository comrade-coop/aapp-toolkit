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
AAPPTAG=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.tag')

sed -i "s|__AAPPHOSTNAME__|${DNS_ROOT}|g" "$NGINX_CONFIG"
sed -i "s|__AAPPPORT__|${AAPPPORT}|g" "$NGINX_CONFIG"

cp azure-attestation/scripts/token.sh /var/www/html/
chmod +x /var/www/html/token.sh
cp azure-attestation/web/index.html /var/www/html/
cp /root/aapp-toolkit/bootstrap/reference.json /var/www/html/
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

cd /root
git clone $AAPPREPO aapp-code
cd aapp-code
git checkout tags/$AAPPTAG

AAPPDOCKERFILE=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.dockerfile')
ENCODED_BUILD_ARGS=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.args')
if [[ -z "$ENCODED_BUILD_ARGS" || "$ENCODED_BUILD_ARGS" == "null" ]]; then
    echo "Error: No valid base64 content found in JSON file!"
    exit 1
fi

# Decode and store it into a aapp.args file
echo "$ENCODED_BUILD_ARGS" | base64 --decode > aapp.args

BUILD_ARGS=""

# Read and process the aapp.args file and ensure proper line endings
sed -i 's/\r$//' aapp.args
while IFS='=' read -r key value; do
    BUILD_ARGS+=" --build-arg $key=$value"
done < "aapp.args"

# Install docker
apt-get update -yq
apt-get install -yq docker.io

systemctl start docker
systemctl enable docker

# Run the docker build command with extracted arguments
docker build $BUILD_ARGS -f $AAPPDOCKERFILE -t aapp-image .

# Run the Docker container in the background
docker run -d -p $AAPPPORT:$AAPPPORT  aapp-image