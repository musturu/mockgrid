#!/bin/sh

set -eu

INIT_DIR=${INIT_DIR:-/docker-entrypoint-initdb.d}
STORAGE_TYPE=${STORAGE_TYPE:-none}
STORAGE_PATH=${STORAGE_PATH:-./data/messages.db}
CONFIG_PATH=${MOCKGRID_CONFIG:-}

log() {
  printf "mockgrid init: %s\n" "$1"
}

has_init_files() {
  [ -d "$INIT_DIR" ] || return 1
  for file in "$INIT_DIR"/*; do
    if [ -e "$file" ]; then
      return 0
    fi
  done
  return 1
}

run_shell_drops() {
  for script in "$INIT_DIR"/*.sh; do
    if [ ! -e "$script" ]; then
      continue
    fi
    log "running shell script $script"
    sh "$script"
  done
}

run_sqlite_scripts() {
  if [ "$STORAGE_TYPE" != "sqlite" ]; then
    return
  fi
  if ! command -v sqlite3 >/dev/null 2>&1; then
    log "sqlite3 binary not available; skipping SQL scripts"
    return
  fi
  log "ensuring sqlite directory exists"
  mkdir -p "$(dirname "$STORAGE_PATH")"
  for script in "$INIT_DIR"/*.sql; do
    if [ ! -e "$script" ]; then
      continue
    fi
    log "applying sqlite script $script"
    sqlite3 "$STORAGE_PATH" < "$script"
  done
}

run_filesystem_json() {
  if [ "$STORAGE_TYPE" != "filesystem" ]; then
    return
  fi
  log "ensuring filesystem store exists"
  mkdir -p "$STORAGE_PATH"
  for script in "$INIT_DIR"/*.json; do
    if [ ! -e "$script" ]; then
      continue
    fi
    dest="$STORAGE_PATH/$(basename "$script")"
    if [ -e "$dest" ]; then
      log "skipping $(basename "$script"): already exists"
      continue
    fi
    log "copying $(basename "$script") into filesystem store"
    cp "$script" "$dest"
  done
}

run_init_scripts() {
  if ! has_init_files; then
    return
  fi
  log "running initialization scripts from $INIT_DIR"
  run_shell_drops
  run_sqlite_scripts
  run_filesystem_json
}

run_init_scripts

if [ -n "$CONFIG_PATH" ]; then
  set -- "$@" "--config" "$CONFIG_PATH"
fi

exec /mockgrid "$@"
