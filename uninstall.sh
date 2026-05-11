#!/usr/bin/env bash
# uninstall.sh — desinstalador do gouse.
#
# Uso público:
#   curl -fsSL https://raw.githubusercontent.com/zewillyan007/gouse/main/uninstall.sh | sh -s -- --yes
#
# Remove:
#   ~/.gouse/
#   linha 'source ~/.gouse/shell.sh' do ~/.bashrc
#   ~/go/gopaths/                              (todos os GOPATHs gerenciados)
#   $XDG_DATA_HOME/gos ou ~/.local/share/gos/  (todas as versões do Go)
#
# Flags:
#   --yes / -y   pula o prompt. Obrigatório quando rodado via `... | sh`.

set -euo pipefail

GOUSE_DIR="$HOME/.gouse"
SHELL_SH="$GOUSE_DIR/shell.sh"
BASHRC="$HOME/.bashrc"
GOPATHS_DIR="$HOME/go/gopaths"
GOS_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/gos"
SOURCE_LINE='source ~/.gouse/shell.sh'

ASSUME_YES=0
for arg in "$@"; do
  case "$arg" in
    --yes|-y) ASSUME_YES=1 ;;
    *) printf 'flag desconhecida: %s\n' "$arg" >&2; exit 2 ;;
  esac
done

log() { printf '%s\n' "$*"; }
err() { printf '%s\n' "$*" >&2; }

confirm() {
  if [ "$ASSUME_YES" -eq 1 ]; then
    return 0
  fi
  if [ ! -t 0 ]; then
    err "Stdin não é interativo. Rode com --yes para confirmar a remoção."
    exit 1
  fi
  log ""
  log "O uninstall removerá:"
  [ -e "$GOUSE_DIR" ]    && log "  - $GOUSE_DIR"
  grep -Fxq "$SOURCE_LINE" "$BASHRC" 2>/dev/null && log "  - linha '$SOURCE_LINE' do $BASHRC"
  [ -e "$GOPATHS_DIR" ]  && log "  - $GOPATHS_DIR (todos os projetos lá dentro)"
  [ -e "$GOS_DIR" ]      && log "  - $GOS_DIR (todas as versões do Go)"
  log ""
  printf "Confirma? [y/N] "
  read -r ans
  case "$ans" in
    y|Y|yes|YES) return 0 ;;
    *) log "Cancelado."; exit 0 ;;
  esac
}

remove_gouse_dir() {
  if [ -e "$GOUSE_DIR" ]; then
    rm -rf "$GOUSE_DIR"
    log "Removido: $GOUSE_DIR"
  fi
}

remove_bashrc_line() {
  if [ ! -f "$BASHRC" ]; then
    return 0
  fi
  if ! grep -Fxq "$SOURCE_LINE" "$BASHRC"; then
    return 0
  fi
  local tmp
  tmp=$(mktemp)
  # Remove a linha do source e o comentário "# gouse — ..." imediatamente anterior se presente.
  awk -v src="$SOURCE_LINE" '
    BEGIN { skip_next_blank = 0 }
    {
      if ($0 == src) { next }
      if ($0 == "# gouse — gerenciador de versões do Go") { next }
      print
    }
  ' "$BASHRC" > "$tmp"
  mv "$tmp" "$BASHRC"
  log "Linha removida de $BASHRC"
}

remove_gopaths() {
  if [ -e "$GOPATHS_DIR" ]; then
    rm -rf "$GOPATHS_DIR"
    log "Removido: $GOPATHS_DIR"
  fi
}

remove_gos() {
  if [ ! -e "$GOS_DIR" ]; then
    return 0
  fi
  rm -rf "$GOS_DIR"
  log "Removido: $GOS_DIR"
}

main() {
  confirm
  remove_gouse_dir
  remove_bashrc_line
  remove_gopaths
  remove_gos
  log ""
  log "Desinstalação concluída."
}

main
