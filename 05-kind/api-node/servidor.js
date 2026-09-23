// B7 - La MISMA API en Node, para que el limite sea el mismo para los dos.
const http = require('node:http');
const PUERTO = process.env.PORT || 8080;

http.createServer((req, res) => {
  if (req.url === '/salud') { res.end('ok desde node\n'); return; }
  const m = req.url.match(/^\/libros\/(.+)$/);
  if (m) { res.end(`libro ${m[1]}\n`); return; }
  res.statusCode = 404; res.end('no\n');
}).listen(PUERTO, () => console.log('api-node escuchando en ' + PUERTO));
