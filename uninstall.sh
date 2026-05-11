#!/usr/bin/env bash
# uninstall.sh — desinstalador do gouse.
#
# Uso público:
#   curl -fsSL https://raw.githubusercontent.com/zewillyan007/gouse/main/uninstall.sh | sh -s -- --yes
#
# Remove:
#   ~/.gouse/
#   linha de source do gouse em todos os rcfiles conhecidos
#       (~/.bashrc, ~/.zshrc, ~/.config/fish/config.fish)
#   ~/go/gopaths/                              (todos os GOPATHs gerenciados)
#   $XDG_DATA_HOME/gos ou ~/.local/share/gos/  (todas as versões do Go)
#
# Flags:
#   --yes / -y   pula o prompt. Obrigatório quando rodado via `... | sh`.

set -euo pipefail

GOUSE_DIR="$HOME/.gouse"
GOPATHS_DIR="$HOME/go/gopaths"
GOS_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/gos"

# rcfile path | source line. Mantém uma única fonte da verdade pra
# extensão a shells novos no futuro.
RCFILES=(
  "$HOME/.bashrc|source ~/.gouse/shell.sh"
  "$HOME/.zshrc|source ~/.gouse/shell.zsh"
  "$HOME/.config/fish/config.fish|source ~/.gouse/shell.fish"
)

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
  local entry rcfile line
  for entry in "${RCFILES[@]}"; do
    rcfile="${entry%%|*}"
    line="${entry##*|}"
    if [ -f "$rcfile" ] && grep -Fxq "$line" "$rcfile"; then
      log "  - linha '$line' do $rcfile"
    fi
  done
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

remove_rcfile_line() {
  local rcfile="$1" line="$2"
  if [ ! -f "$rcfile" ] || ! grep -Fxq "$line" "$rcfile"; then
    return 0
  fi
  local tmp
  tmp=$(mktemp)
  awk -v src="$line" '
    {
      if ($0 == src) { next }
      if ($0 == "# gouse — gerenciador de versões do Go") { next }
      print
    }
  ' "$rcfile" > "$tmp"
  mv "$tmp" "$rcfile"
  log "Linha removida de $rcfile"
}

remove_all_rcfile_lines() {
  local entry rcfile line
  for entry in "${RCFILES[@]}"; do
    rcfile="${entry%%|*}"
    line="${entry##*|}"
    remove_rcfile_line "$rcfile" "$line"
  done
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
  remove_all_rcfile_lines
  remove_gopaths
  remove_gos
  log ""
  log "Desinstalação concluída."
}

main
