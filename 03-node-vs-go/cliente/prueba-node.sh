#!/bin/bash
# Terminal C. La cifra del cliente duele; la del servidor explica.
P=18092
curl -s -o /dev/null -w "pesado -> %{time_total}s\n" "http://localhost:$P/pesado" &
sleep 1.2
curl -s -o /dev/null -w "rapido -> %{time_total}s\n" "http://localhost:$P/rapido"
wait
