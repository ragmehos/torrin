#!/bin/sh
set -e

COMPOSE="docker compose -f $(dirname "$0")/docker-compose.yml"
ACCESS="${S3_ACCESS_KEY:-GKdeadbeefdeadbeefdeadbeef}"
SECRET="${S3_SECRET_KEY:-deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef}"
BUCKET="${S3_BUCKET:-torrin}"

echo "waiting for garage..."
until $COMPOSE exec -T garage /garage status >/dev/null 2>&1; do sleep 1; done

NODE=$($COMPOSE exec -T garage /garage node id -q | cut -d@ -f1 | tr -d '\r')
echo "node: $NODE"

$COMPOSE exec -T garage /garage layout assign -z dc1 -c 1G "$NODE" || true
$COMPOSE exec -T garage /garage layout apply --version 1 || true
$COMPOSE exec -T garage /garage bucket create "$BUCKET" || true
$COMPOSE exec -T garage /garage key import --yes "$ACCESS" "$SECRET" -n torrin-key || true
$COMPOSE exec -T garage /garage bucket allow --read --write --owner "$BUCKET" --key torrin-key || true

echo "garage ready — bucket=$BUCKET key=$ACCESS"
