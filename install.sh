#!/usr/bin/env bash
# install.sh — instalador do gouse.
#
# Uso público:
#   curl -fsSL https://raw.githubusercontent.com/zewillyan007/gouse/main/install.sh | sh
#
# Flags:
#   --local           builda a partir do código-fonte no diretório atual (dev)
#   --shell <name>    força o shell-alvo (bash|zsh|fish). Default = autodetect via $SHELL.
#
# Garantias:
#   - Sem sudo. Só cria/escreve em ~/.gouse/ e no rcfile do shell detectado.
#   - Zero resíduo: nada de .bak, .tmp ou pastas temporárias fora de ~/.gouse/.

set -euo pipefail

REPO="zewillyan007/gouse"
GOUSE_DIR="$HOME/.gouse"
GOUSE_BIN_DIR="$GOUSE_DIR/bin"
GOUSE_BIN="$GOUSE_BIN_DIR/gouse"
TMP_BIN="$GOUSE_DIR/.tmp-gouse"
TMP_SUMS="$GOUSE_DIR/.tmp-sums"

# Script-level cleanup: ensures the temp files are removed on any exit
# path (success, failure, Ctrl-C). Declared at script scope so the trap
# can reference the variables when EXIT fires after functions return.
trap 'rm -f "$TMP_BIN" "$TMP_SUMS"' EXIT

MODE="release"
SHELL_OVERRIDE=""
while [ $# -gt 0 ]; do
  case "$1" in
    --local) MODE="local" ;;
    --shell) SHELL_OVERRIDE="${2:-}"; shift ;;
    --shell=*) SHELL_OVERRIDE="${1#--shell=}" ;;
    *) echo "flag desconhecida: $1" >&2; exit 2 ;;
  esac
  shift
done

log() { printf '%s\n' "$*"; }
err() { printf '%s\n' "$*" >&2; }

# detect_arch echoes the canonical arch name (amd64 | arm64) used by the
# release asset filenames; aborts on unsupported machines.
detect_arch() {
  local m
  m=$(uname -m)
  case "$m" in
    x86_64|amd64)   echo amd64 ;;
    aarch64|arm64)  echo arm64 ;;
    *) err "Arquitetura $m não suportada (linux/amd64 e linux/arm64)."; exit 1 ;;
  esac
}

check_platform() {
  local os
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  if [ "$os" != "linux" ]; then
    err "OS $os não suportado nesta versão do gouse (somente linux)."
    exit 1
  fi
}

# detect_shell echoes the canonical shell name (bash | zsh | fish).
# $GOUSE_SHELL env var or --shell flag overrides $SHELL detection.
detect_shell() {
  if [ -n "${SHELL_OVERRIDE:-}" ]; then echo "$SHELL_OVERRIDE"; return; fi
  if [ -n "${GOUSE_SHELL:-}" ]; then echo "$GOUSE_SHELL"; return; fi
  case "${SHELL:-}" in
    */bash|bash) echo bash ;;
    */zsh|zsh)   echo zsh ;;
    */fish|fish) echo fish ;;
    *)           echo bash ;;
  esac
}

# rcfile_for echoes the absolute path of the rcfile for the given shell.
rcfile_for() {
  case "$1" in
    bash) echo "$HOME/.bashrc" ;;
    zsh)  echo "$HOME/.zshrc" ;;
    fish) echo "$HOME/.config/fish/config.fish" ;;
    *)    err "shell desconhecido: $1"; exit 1 ;;
  esac
}

# source_line_for echoes the line to add to the rcfile for the given shell.
source_line_for() {
  case "$1" in
    bash) echo 'source ~/.gouse/shell.sh' ;;
    zsh)  echo 'source ~/.gouse/shell.zsh' ;;
    fish) echo 'source ~/.gouse/shell.fish' ;;
    *)    err "shell desconhecido: $1"; exit 1 ;;
  esac
}

check_gouse_dir() {
  if [ ! -e "$GOUSE_DIR" ]; then
    return 0
  fi
  # Treat as previous install if any known integration file is present.
  if [ -d "$GOUSE_DIR" ] && {
       [ -f "$GOUSE_DIR/shell.sh" ] || \
       [ -f "$GOUSE_DIR/shell.zsh" ] || \
       [ -f "$GOUSE_DIR/shell.fish" ]; }; then
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
  local arch asset base
  arch=$(detect_arch)
  asset="gouse-linux-$arch"
  base="https://github.com/$REPO/releases/latest/download"

  log "Baixando $base/$asset..."
  fetch "$base/$asset" "$TMP_BIN"
  fetch "$base/SHA256SUMS" "$TMP_SUMS"

  local expected got
  expected=$(awk -v a="$asset" '$2 == a {print $1}' "$TMP_SUMS")
  got=$(sha256sum "$TMP_BIN" | awk '{print $1}')
  if [ -z "$expected" ]; then
    err "SHA256SUMS não contém entrada para $asset."
    exit 1
  fi
  if [ "$expected" != "$got" ]; then
    err "Falha na verificação SHA256."
    err "  esperado: $expected"
    err "  obtido:   $got"
    exit 1
  fi

  mv "$TMP_BIN" "$GOUSE_BIN"
  chmod +x "$GOUSE_BIN"
}

run_init() {
  local sh="$1"
  "$GOUSE_BIN" init --shell "$sh" >/dev/null
}

update_rcfile() {
  local sh="$1"
  local rcfile source_line
  rcfile=$(rcfile_for "$sh")
  source_line=$(source_line_for "$sh")
  mkdir -p "$(dirname "$rcfile")"
  touch "$rcfile"
  if grep -Fxq "$source_line" "$rcfile"; then
    return 0
  fi
  printf '\n# gouse — gerenciador de versões do Go\n%s\n' "$source_line" >> "$rcfile"
}

cleanup_residue() {
  rm -rf "$GOUSE_DIR/tmp" 2>/dev/null || true
}

main() {
  check_platform
  check_gouse_dir
  ensure_dirs

  local detected_shell
  detected_shell=$(detect_shell)

  case "$MODE" in
    local)   build_local ;;
    release) download_release ;;
  esac
  run_init "$detected_shell"
  update_rcfile "$detected_shell"
  cleanup_residue

  local rcfile
  rcfile=$(rcfile_for "$detected_shell")
  log ""
  log "Pronto. Shell detectado: $detected_shell"
  log "Abra um novo terminal (ou rode 'source $rcfile') e use:"
  log "  gouse install latest"
  log "  gouse use go1.x.y"
}

main
