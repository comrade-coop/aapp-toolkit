const https = require('https');
const fs = require('fs');
const path = require('path');
const tar = require('tar');

/**
 * Convert wildcard pattern to RegExp.
 * '*' matches zero or more characters.
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

const server = https.createServer(options, (req, res) => {
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
