// B4 - Node: el bucle de eventos bloqueado.
// Lo importante ocurre en ESTE log, no en la salida de curl.
const http = require('node:http');

const PUERTO = 18092;
const VUELTAS = 3_000_000_000;
const inicio = process.hrtime.bigint();

function t() {
  const seg = Number(process.hrtime.bigint() - inicio) / 1e9;
  return seg.toFixed(3).padStart(6, ' ');
}

const servidor = http.createServer((req, res) => {
  if (req.url.startsWith('/pesado')) {
    console.log(`${t()}s  entra /pesado`);
    // Calculo puro. No hay await, no hay E/S: el unico hilo se queda aqui.
    let suma = 0;
    for (let i = 0; i < VUELTAS; i++) suma += i;
    console.log(`${t()}s  sale  /pesado`);
    res.end(`pesado listo, suma=${suma}\n`);
    return;
  }
  console.log(`${t()}s  entra /rapido`);
  res.end('rapido\n');
  console.log(`${t()}s  sale  /rapido`);
});

servidor.listen(PUERTO, () => console.log(`node escuchando en ${PUERTO}`));
