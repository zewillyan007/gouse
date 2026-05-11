#!/usr/bin/env bash
# install.sh — instalador do gouse.
#
# Uso público:
#   curl -fsSL https://raw.githubusercontent.com/zewillyan007/gouse/main/install.sh | sh
#
# Modos:
#   (sem flag)   baixa o binário do GitHub Releases
#   --local      builda a partir do código-fonte no diretório atual (dev)
#
# Garantias:
#   - Sem sudo. Só cria/escreve em ~/.gouse/ e edita ~/.bashrc.
#   - Zero resíduo: nada de .bak, .tmp ou pastas temporárias fora de ~/.gouse/.

set -euo pipefail

REPO="zewillyan007/gouse"
GOUSE_DIR="$HOME/.gouse"
GOUSE_BIN_DIR="$GOUSE_DIR/bin"
GOUSE_BIN="$GOUSE_BIN_DIR/gouse"
SHELL_SH="$GOUSE_DIR/shell.sh"
BASHRC="$HOME/.bashrc"
SOURCE_LINE='source ~/.gouse/shell.sh'

MODE="release"
for arg in "$@"; do
  case "$arg" in
    --local) MODE="local" ;;
    *) echo "flag desconhecida: $arg" >&2; exit 2 ;;
  esac
done

log() { printf '%s\n' "$*"; }
err() { printf '%s\n' "$*" >&2; }

check_platform() {
  local os arch
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    *) err "Arquitetura $arch não suportada nesta versão do gouse."; exit 1 ;;
  esac
  if [ "$os" != "linux" ] || [ "$arch" != "amd64" ]; then
    err "OS/Arch $os/$arch não suportado nesta versão do gouse (somente linux/amd64)."
    exit 1
  fi
}

check_gouse_dir() {
  if [ ! -e "$GOUSE_DIR" ]; then
    return 0
  fi
  if [ -d "$GOUSE_DIR" ] && [ -f "$GOUSE_DIR/shell.sh" ]; then
    # install anterior do gouse — upgrade ok
    return 0
  fi
  err "$GOUSE_DIR já existe e não parece ser um install do gouse."
  err "Mova ou remova esse caminho manualmente e rode novamente."
  exit 1
}

ensure_dirs() {
  mkdir -p "$GOUSE_BIN_DIR"
}

build_local() {
  if ! command -v go >/dev/null 2>&1; then
    err "Modo --local exige 'go' instalado no PATH."
    exit 1
  fi
  log "Compilando gouse (modo local)..."
  go build -o "$GOUSE_BIN" ./cmd/gouse
}

fetch() {
  local url="$1" dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$dest" "$url"
  else
    err "É necessário 'curl' ou 'wget' para baixar arquivos."
    exit 1
  fi
}

download_release() {
  local base="https://github.com/$REPO/releases/latest/download"
  local tmp_bin="$GOUSE_DIR/.tmp-gouse"
  local tmp_sums="$GOUSE_DIR/.tmp-sums"
  trap 'rm -f "$tmp_bin" "$tmp_sums"' EXIT

  log "Baixando $base/gouse-linux-amd64..."
  fetch "$base/gouse-linux-amd64" "$tmp_bin"
  fetch "$base/SHA256SUMS" "$tmp_sums"

  local expected got
  expected=$(awk '/gouse-linux-amd64/ {print $1}' "$tmp_sums")
  got=$(sha256sum "$tmp_bin" | awk '{print $1}')
  if [ -z "$expected" ]; then
    err "SHA256SUMS não contém entrada para gouse-linux-amd64."
    exit 1
  fi
  if [ "$expected" != "$got" ]; then
    err "Falha na verificação SHA256."
    err "  esperado: $expected"
    err "  obtido:   $got"
    exit 1
  fi

  mv "$tmp_bin" "$GOUSE_BIN"
  chmod +x "$GOUSE_BIN"
}

run_init() {
  "$GOUSE_BIN" init >/dev/null
}

update_bashrc() {
  touch "$BASHRC"
  if grep -Fxq "$SOURCE_LINE" "$BASHRC"; then
    return 0
  fi
  printf '\n# gouse — gerenciador de versões do Go\n%s\n' "$SOURCE_LINE" >> "$BASHRC"
}

cleanup_residue() {
  # Garante que nada além de ~/.gouse/ foi criado.
  rm -rf "$GOUSE_DIR/tmp" 2>/dev/null || true
}

main() {
  check_platform
  check_gouse_dir
  ensure_dirs
  case "$MODE" in
    local)   build_local ;;
    release) download_release ;;
  esac
  run_init
  update_bashrc
  cleanup_residue
  log ""
  log "Pronto. Abra um novo terminal (ou rode 'source ~/.bashrc') e use:"
  log "  gouse install latest"
  log "  gouse use go1.x.y"
}

main
