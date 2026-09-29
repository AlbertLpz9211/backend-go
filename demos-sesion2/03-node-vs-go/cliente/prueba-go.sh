#!/bin/bash
# Terminal C. El curl a /rapido corre en bucle DESDE ANTES del enter,
# para que la pantalla no este quieta mientras dura el bucle pesado.
P=18090
curl -s -o /dev/null -w "pesado -> %{time_total}s\n" "http://localhost:$P/pesado" &
for i in 1 2 3 4; do
  sleep 0.4
  curl -s -o /dev/null -w "rapido -> %{time_total}s\n" "http://localhost:$P/rapido"
done
wait
