#!/bin/bash

USER_DATA_BASE64=$(curl -H Metadata:true --noproxy "*" "http://169.254.169.254/metadata/instance/compute/userData?api-version=2021-01-01&format=text")
USER_DATA_JSON=$(echo "$USER_DATA_BASE64" | base64 --decode)

mkdir -p /var/www/html

echo "$USER_DATA_JSON" | sha1sum | awk '{print $1}' > /var/www/html/manifest.sha1

HOSTNAMES=$(echo "$USER_DATA_JSON" | jq -r '.spec.ingress.hostnames[]')
ALLOWED_PATTERN=$(echo "$USER_DATA_JSON" | jq -r '.spec.bootstrapping.pattern')

# Pick the first hostname as the primary CN
DNS_ROOT=$(echo "$HOSTNAMES" | head -n 1)

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
EOF

NGINX_SERVER_NAMES=""

# Append all hostnames to alt_names section
COUNT=1
for HOST in $HOSTNAMES; do
  echo "DNS.$COUNT = $HOST" >> "$OPENSSL_CONF"
  NGINX_SERVER_NAMES+="$HOST "
  COUNT=$((COUNT + 1))
done

NGINX_SERVER_NAMES=$(echo "$NGINX_SERVER_NAMES" | xargs)

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
' fullchain.pem | openssl x509 -noout -fingerprint -sha1 | sed 's/^.*=//; s/://g' | tr '[:upper:]' '[:lower:]' > /var/www/html/cert.sha1

cd ../

apt-get update -yq
apt-get install -yq nginx
apt-get install -yq libnginx-mod-http-lua

NGINX_CONFIG="/etc/nginx/sites-available/default"

cp -f ingress/default.conf $NGINX_CONFIG

AAPPPORT=$(echo "$USER_DATA_JSON" | jq -r '.spec.ingress.port')
AAPPREPO=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.repo')
AAPPTAG=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.tag')

sed -i "s|__DNS_ROOT__|${DNS_ROOT}|g" "$NGINX_CONFIG"
sed -i "s|__NGINX_SERVER_NAMES__|${NGINX_SERVER_NAMES}|g" "$NGINX_CONFIG"

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
BUILD_ARGS_JSON=$(echo "$USER_DATA_JSON" | jq -r '.spec.container.build.args')

BUILD_ARGS=""

# Parse JSON and construct Docker build arguments
while IFS="=" read -r key value; do
    key=$(echo "$key" | xargs)
    value=$(echo "$value" | xargs)
    BUILD_ARGS+=" --build-arg $key=$value"
done < <(echo "$BUILD_ARGS_JSON" | jq -r 'to_entries | map("\(.key)=\(.value)") | .[]')

# Install docker from official docker repository
apt update
apt install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo tee /etc/apt/keyrings/docker.asc > /dev/null
chmod a+r /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

systemctl start docker
systemctl enable docker

# Run the docker build command with extracted arguments
DOCKER_BUILDKIT=1 docker build $BUILD_ARGS -f $AAPPDOCKERFILE -t aapp-image .

# Run the Docker container in the background with mounts
cd /root
CLOUD_MOUNT_HOST_DIR=""
MOUNT_OPTS=""
VOLUMES=$(echo "$USER_DATA_JSON" | jq -c '.spec.container.volumes // []')
for row in $(echo "${VOLUMES}" | jq -c '.[]'); do
  NAME=$(echo "$row" | jq -r '.name')
  MOUNT=$(echo "$row" | jq -r '.mount')
  TYPE=$(echo "$row" | jq -r '.type')

  # Create directory
  HOST_DIR="./volumes/$NAME"
  mkdir -p "$HOST_DIR"

  # Append mount option for Docker
  MOUNT_OPTS="$MOUNT_OPTS -v $(realpath $HOST_DIR):$MOUNT"

  if [[ $TYPE == "cloud" ]]; then
    CLOUD_MOUNT_HOST_DIR=$(realpath $HOST_DIR)
  fi
done

cp /root/aapp-toolkit/bootstrap/fullchain.pem /root/aapp-toolkit/bootstrap/server.pem
cp /root/aapp-toolkit/${DNS_ROOT}.key /root/aapp-toolkit/bootstrap/server.key
cd /root/aapp-toolkit/bootstrap

BOOTSTRAPPING_PARENT=$(echo "$USER_DATA_JSON" | jq -r '.spec.bootstrapping.parent // empty')
if [[ -n $BOOTSTRAPPING_PARENT ]]; then
  curl -sSf --cert server.pem --key server.key https://$BOOTSTRAPPING_PARENT:54321 -o cloud-app-volume.tar.gz
  tar -xzf cloud-app-volume.tar.gz --strip-components=1 -C $CLOUD_MOUNT_HOST_DIR
fi

docker run -d -p 3000:$AAPPPORT $MOUNT_OPTS --restart=always aapp-image

if [[ -n $CLOUD_MOUNT_HOST_DIR ]]; then
  docker build -t aapp-toolkit-server .
  docker run -d -p 54321:54321 -e ALLOWED_PATTERN=$ALLOWED_PATTERN -v $CLOUD_MOUNT_HOST_DIR:/cloud-app-volume --restart=always aapp-toolkit-server
fi