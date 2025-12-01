#!/bin/bash
set -e
# Ждём готовности RabbitMQ
until rabbitmqctl status; do
  echo "RabbitMQ не готов..."
  sleep 2
done

rabbitmqctl add_vhost foodservice

rabbitmqctl set_permissions -p foodservice admin ".*" ".*" ".*"

rabbitmqctl delete_vhost /

rabbitmqctl import_definitions /opt/definitions.json
