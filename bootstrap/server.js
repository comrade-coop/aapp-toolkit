const https = require('https');
const fs = require('fs');
const path = require('path');
const tar = require('tar');
const crypto = require('crypto');
const { jwtVerify, createRemoteJWKSet, decodeProtectedHeader } = require('jose');
const { URL } = require('url');

/**
 * Convert wildcard pattern to RegExp. '*' matches zero or more characters.
 */
function wildcardToRegExp(pattern) {
  const escaped = pattern.replace(/[.+?^${}()|[\]\\]/g, '\\$&');
  const regexString = '^' + escaped.replace(/\*/g, '.*') + '$';
  return new RegExp(regexString);
}

const options = {
  key: fs.readFileSync(path.join(__dirname, 'server.key')),
  cert: fs.readFileSync(path.join(__dirname, 'server.pem')),
  requestCert: true,
  rejectUnauthorized: true // only allow clients with valid, public-CA-signed certs
};

const PORT = 54321;
const ALLOWED_PATTERN = process.env.ALLOWED_PATTERN;
if (!ALLOWED_PATTERN) {
  console.error('Missing required environment variable: ALLOWED_PATTERN');
  process.exit(1);
}
const allowedRegex = wildcardToRegExp(ALLOWED_PATTERN);

const server = https.createServer(options, async (req, res) => {
  const socket = req.socket || req.connection;
  const authorized = socket.authorized;
  const cert = socket.getPeerCertificate();

  if (!authorized || !cert || !cert.subject) {
    res.writeHead(401);
    return res.end('Unauthorized: Invalid client certificate');
  }

  const domain = cert.subject.CN;
  if (!allowedRegex.test(domain)) {
    res.writeHead(403);
    return res.end(`Forbidden: CN "${domain}" does not match allowed pattern "${ALLOWED_PATTERN}"`);
  }

  // 1. Generate nonce and fetch attestation token
  const nonce = crypto.randomBytes(16).toString('base64');
  const tokenEndpoint = `https://${domain}/.well-known/attest/token?nonce=${encodeURIComponent(nonce)}`;
  let attRes;
  try {
    attRes = await fetch(tokenEndpoint);
  } catch (err) {
    res.writeHead(502);
    return res.end('Bad Gateway: failed to fetch attestation token');
  }
  if (!attRes.ok) {
    res.writeHead(502);
    return res.end(`Bad Gateway: attestation server returned ${attRes.status}`);
  }
  let attJson;
  try {
    attJson = await attRes.json();
  } catch (err) {
    res.writeHead(502);
    return res.end('Bad Gateway: invalid JSON in attestation response');
  }
  const { token } = attJson;
  if (!token) {
    res.writeHead(502);
    return res.end('Bad Gateway: no token in attestation response');
  }

  // 2. Verify JWT using jose and remote JWK set
  let header;
  try {
    header = decodeProtectedHeader(token);
  } catch (err) {
    res.writeHead(403);
    return res.end('Forbidden: invalid JWT header');
  }
  const jwksUrl = new URL(header.jku);
  const JWKS = createRemoteJWKSet(jwksUrl);
  let verified;
  try {
    verified = await jwtVerify(token, JWKS);
  } catch (err) {
    res.writeHead(403);
    return res.end('Forbidden: JWT verification failed');
  }

  // 3. Validate nonce in x-ms-runtime.client-payload.nonce
  const payload = verified.payload;
  const runtime = payload['x-ms-runtime'];
  const clientPayload = runtime && runtime['client-payload'];
  if (!clientPayload || clientPayload.nonce !== nonce) {
    res.writeHead(403);
    return res.end('Forbidden: nonce mismatch');
  }

  // Stream tar.gz of /cloud-app-volume
  res.writeHead(200, {
    'Content-Type': 'application/gzip',
    'Content-Disposition': 'attachment; filename="cloud-app-volume.tar.gz"'
  });

  tar.c(
    { gzip: true, cwd: '/', portable: true },
    ['cloud-app-volume']
  ).pipe(res).on('error', err => {
    console.error('Tar error:', err);
    if (!res.headersSent) {
      res.writeHead(500);
      res.end('Internal Server Error');
    }
  });
});

server.listen(PORT, () => {
  console.log(`HTTPS server listening on port ${PORT}`);
});
