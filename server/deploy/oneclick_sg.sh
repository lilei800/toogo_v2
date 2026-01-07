#!/usr/bin/env bash
# Toogo/HotGo one-click deploy on Ubuntu 24.04 (Singapore VPS)
# - Docker + Docker Compose plugin install
# - Configure config.yaml for docker
# - Bring up: postgres + redis + hotgo + nginx
# - Obtain Let's Encrypt cert for www.toogo.my + toogo.my
#
# Usage (recommended):
#   sudo PG_PASSWORD='***' REDIS_PASSWORD='***' EMAIL='admin@toogo.my' bash server/deploy/oneclick_sg.sh
#
# Notes:
# - This will create files under /opt/toogo/toogo_v2/server/deploy_data (DB/Redis/Certs).
# - First run will import ./storage/data/initial_data.sql into Postgres automatically.

set -euo pipefail

PROJECT_DIR_DEFAULT="/opt/toogo/toogo_v2/server"
PROJECT_DIR="${PROJECT_DIR:-$PROJECT_DIR_DEFAULT}"

DOMAIN_MAIN="${DOMAIN_MAIN:-www.toogo.my}"
DOMAIN_ALT="${DOMAIN_ALT:-toogo.my}"
EMAIL="${EMAIL:-admin@toogo.my}"
TZ="${TZ:-Asia/Singapore}"

PG_DB="${PG_DB:-hotgo}"
PG_USER="${PG_USER:-hotgo_user}"
PG_PASSWORD="${PG_PASSWORD:-}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"

TOKEN_SECRET_KEY="${TOKEN_SECRET_KEY:-}"
TCP_CRON_SECRET_KEY="${TCP_CRON_SECRET_KEY:-hotgo_cron_secret}"
TCP_AUTH_SECRET_KEY="${TCP_AUTH_SECRET_KEY:-hotgo_auth_secret}"

need_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "ERROR: please run as root (use sudo)." >&2
    exit 1
  fi
}

require_vars() {
  if [[ -z "${PG_PASSWORD}" ]]; then
    echo "ERROR: PG_PASSWORD is required. Example:" >&2
    echo "  sudo PG_PASSWORD='Toogo2027!#$888' REDIS_PASSWORD='Redis2027!@#$888' bash server/deploy/oneclick_sg.sh" >&2
    exit 1
  fi
  if [[ -z "${REDIS_PASSWORD}" ]]; then
    echo "ERROR: REDIS_PASSWORD is required." >&2
    exit 1
  fi
}

install_docker() {
  if command -v docker >/dev/null 2>&1; then
    return 0
  fi

  apt-get update -y
  apt-get install -y ca-certificates curl gnupg lsb-release
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg

  local codename
  codename="$(. /etc/os-release && echo "$VERSION_CODENAME")"
  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
    ${codename} stable" | tee /etc/apt/sources.list.d/docker.list >/dev/null

  apt-get update -y
  apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
}

prepare_files() {
  if [[ ! -d "${PROJECT_DIR}" ]]; then
    echo "ERROR: PROJECT_DIR not found: ${PROJECT_DIR}" >&2
    exit 1
  fi

  cd "${PROJECT_DIR}"

  mkdir -p deploy_data/certbot/www deploy_data/certbot/conf deploy_data/nginx_logs logs storage

  # Create .env if missing
  if [[ ! -f ".env" ]]; then
    cp -f deploy/env.prod.example .env
  fi

  # Generate token secret if not set
  if [[ -z "${TOKEN_SECRET_KEY}" ]]; then
    TOKEN_SECRET_KEY="$(openssl rand -hex 32)"
  fi

  # Write required envs (overwrite/add)
  # shellcheck disable=SC2016
  sed -i \
    -e "s/^TZ=.*/TZ=${TZ}/" \
    -e "s/^PG_DB=.*/PG_DB=${PG_DB}/" \
    -e "s/^PG_USER=.*/PG_USER=${PG_USER}/" \
    -e "s/^PG_PASSWORD=.*/PG_PASSWORD=${PG_PASSWORD}/" \
    -e "s/^REDIS_PASSWORD=.*/REDIS_PASSWORD=${REDIS_PASSWORD}/" \
    -e "s/^TOKEN_SECRET_KEY=.*/TOKEN_SECRET_KEY=${TOKEN_SECRET_KEY}/" \
    -e "s/^TCP_CRON_SECRET_KEY=.*/TCP_CRON_SECRET_KEY=${TCP_CRON_SECRET_KEY}/" \
    -e "s/^TCP_AUTH_SECRET_KEY=.*/TCP_AUTH_SECRET_KEY=${TCP_AUTH_SECRET_KEY}/" \
    .env

  # Ensure docker config is active as config.yaml
  if [[ -f "manifest/config/config.docker.yaml" ]]; then
    cp -f manifest/config/config.docker.yaml manifest/config/config.yaml
  fi

  # Replace placeholders in config.yaml (gf won't expand ${VAR:default} reliably per repo notes)
  sed -i \
    -e "s/CHANGE_ME_PG_PASSWORD/${PG_PASSWORD//\//\\/}/g" \
    -e "s/CHANGE_ME_REDIS_PASSWORD/${REDIS_PASSWORD//\//\\/}/g" \
    -e "s/CHANGE_ME_TOKEN_SECRET_KEY/${TOKEN_SECRET_KEY//\//\\/}/g" \
    -e "s/CHANGE_ME_TCP_CRON_SECRET_KEY/${TCP_CRON_SECRET_KEY//\//\\/}/g" \
    -e "s/CHANGE_ME_TCP_AUTH_SECRET_KEY/${TCP_AUTH_SECRET_KEY//\//\\/}/g" \
    manifest/config/config.yaml
}

open_firewall() {
  if command -v ufw >/dev/null 2>&1; then
    ufw allow 22/tcp >/dev/null 2>&1 || true
    ufw allow 80/tcp >/dev/null 2>&1 || true
    ufw allow 443/tcp >/dev/null 2>&1 || true
  fi
}

compose_up_http() {
  cd "${PROJECT_DIR}"
  docker compose -f deploy/docker-compose.prod.yml up -d --build postgres redis hotgo nginx
}

issue_cert() {
  cd "${PROJECT_DIR}"

  # Try to obtain cert. Requires DNS to point to this server and port 80 reachable.
  docker compose -f deploy/docker-compose.prod.yml -f deploy/docker-compose.https.yml run --rm \
    certbot \
    "certbot certonly --webroot -w /var/www/certbot \
      -d ${DOMAIN_MAIN} -d ${DOMAIN_ALT} \
      --email ${EMAIL} --agree-tos --non-interactive --rsa-key-size 4096"
}

compose_up_https() {
  cd "${PROJECT_DIR}"
  docker compose -f deploy/docker-compose.prod.yml -f deploy/docker-compose.https.yml up -d nginx
}

show_status() {
  cd "${PROJECT_DIR}"
  echo "---- docker compose ps ----"
  docker compose -f deploy/docker-compose.prod.yml ps
  echo "---- curl checks ----"
  curl -fsS "http://${DOMAIN_MAIN}/healthz" >/dev/null && echo "OK: http /healthz" || echo "WARN: http /healthz failed"
  curl -kfsS "https://${DOMAIN_MAIN}/healthz" >/dev/null && echo "OK: https /healthz" || echo "WARN: https /healthz failed (cert may not be ready)"
  echo "Done."
}

main() {
  need_root
  require_vars
  install_docker
  open_firewall
  prepare_files
  compose_up_http

  echo "Attempting to issue Let's Encrypt cert for ${DOMAIN_MAIN}, ${DOMAIN_ALT} ..."
  if issue_cert; then
    compose_up_https
    echo "✅ HTTPS enabled."
  else
    echo "⚠️ Cert issuance failed. Common reasons:" >&2
    echo "  - DNS not pointing to this server yet" >&2
    echo "  - Port 80 blocked by firewall/security group" >&2
    echo "You can re-run later:" >&2
    echo "  sudo PG_PASSWORD='***' REDIS_PASSWORD='***' EMAIL='${EMAIL}' bash server/deploy/oneclick_sg.sh" >&2
  fi

  show_status
}

main "$@"

