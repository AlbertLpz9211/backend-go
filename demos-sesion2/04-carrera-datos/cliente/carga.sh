#!/bin/bash
# 200 peticiones concurrentes. El numero exacto varia; el WARNING no.
P=18094
for i in $(seq 1 200); do
  curl -s -o /dev/null "http://localhost:$P/libros/$i" &
done
wait
echo -n "contador final: "; curl -s "http://localhost:$P/total"
