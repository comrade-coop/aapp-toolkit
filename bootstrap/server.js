const https = require('https');
const fs = require('fs');
const path = require('path');
const tar = require('tar');

const options = {
  key: fs.readFileSync(path.join(__dirname, 'server.key')),
  cert: fs.readFileSync(path.join(__dirname, 'server.pem')),
  requestCert: true,
  rejectUnauthorized: true // reject unauthorized clients at TLS layer
};

const PORT = 54321;
const ALLOWED_DOMAIN = process.env.ALLOWED_DOMAIN;

if (!ALLOWED_DOMAIN) {
  console.error('Missing required environment variable: ALLOWED_DOMAIN');
  process.exit(1);
}

const server = https.createServer(options, (req, res) => {
  const socket = req.socket;
  const authorized = socket.authorized;
  const cert = socket.getPeerCertificate();

  if (!authorized || !cert || !cert.subject) {
    res.writeHead(401);
    return res.end('Unauthorized: Invalid client certificate');
  }

  const domain = cert.subject.CN;
  if (domain !== ALLOWED_DOMAIN) {
    res.writeHead(403);
    return res.end(`Forbidden: Invalid domain ${domain}`);
  }

  // Stream tar.gz of /cloud-app-volume
  res.writeHead(200, {
    'Content-Type': 'application/gzip',
    'Content-Disposition': 'attachment; filename="cloud-app-volume.tar.gz"'
  });

  tar.c(
    {
      gzip: true,
      cwd: '/', // root, so path is absolute
      portable: true
    },
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
