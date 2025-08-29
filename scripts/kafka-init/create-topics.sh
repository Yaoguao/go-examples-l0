#!/bin/bash
# Скрипт создания топиков

BROKER=${BROKER1_HOST:-kafka1:29091}


until /usr/bin/kafka-topics --bootstrap-server "$BROKER" --list >/dev/null 2>&1; do
  echo "not connection $BROKER..."
  sleep 10
done
echo "ready connection"

TOPICS=($ORDER_TOPIC)

for topic in "${TOPICS[@]}"; do
  echo "→ CHECK $topic..."
  /usr/bin/kafka-topics --bootstrap-server "$BROKER" --topic "$topic" --describe >/dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "🛠️  CONNECTION $topic..."
    /usr/bin/kafka-topics --bootstrap-server "$BROKER" --create --if-not-exists \
      --topic "$topic" --partitions 3 --replication-factor 1
  else
    echo "✔️  Error:, $topic already there"
  fi
done