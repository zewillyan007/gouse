# gouse

Gerenciador simples de versões do Go para Linux. Instala, remove e troca entre versões do Go **sem reabrir o terminal e sem sudo**.

Inspirado em `nvm`/`pyenv`: uma função shell intercepta `gouse use` e ajusta `PATH`/`GOPATH` no shell atual.

> **Plataformas suportadas**: `linux/amd64`, `linux/arm64`.
> **Shells suportados**: `bash`, `zsh`, `fish` (autodetectados via `$SHELL`).

## Instalação

```sh
curl -fsSL https://raw.githubusercontent.com/zewillyan007/gouse/main/install.sh | sh
```

O instalador (sem sudo):

1. Detecta arquitetura (`uname -m`) e shell (`$SHELL`).
2. Baixa o binário correto (`gouse-linux-amd64` ou `gouse-linux-arm64`) para `~/.gouse/bin/gouse`.
3. Gera `~/.gouse/shell.{sh|zsh|fish}` conforme o shell detectado.
4. Acrescenta a linha `source ...` ao rcfile correspondente (idempotente):

| Shell | rcfile | linha adicionada |
|---|---|---|
| bash | `~/.bashrc` | `source ~/.gouse/shell.sh` |
| zsh | `~/.zshrc` | `source ~/.gouse/shell.zsh` |
| fish | `~/.config/fish/config.fish` | `source ~/.gouse/shell.fish` |

Para forçar um shell específico (útil se você usa bash mas está rodando o instalador em zsh, por ex.):

```sh
curl -fsSL https://raw.githubusercontent.com/zewillyan007/gouse/main/install.sh | sh -s -- --shell zsh
```

Tudo do gouse fica em `~/.gouse/`. As versões do Go ficam em `~/.local/share/gos/` (XDG). Em nenhum momento o gouse pede sudo.

Após o install, abra um novo terminal (ou `source ~/.bashrc`).

### Verificação de integridade

Cada release publica um arquivo `SHA256SUMS` junto dos binários, que são nomeados no formato `gouse-<versão>-<os>-<arch>` (ex: `gouse-v0.2.3-linux-amd64`). O `install.sh` descobre a versão mais recente via header de redirect, baixa o asset correto e compara o hash contra `SHA256SUMS` antes de instalar.

Caso prefira fazer manualmente:

```sh
# Descobre a tag mais recente
tag=$(curl -fsSI https://github.com/zewillyan007/gouse/releases/latest \
        | awk 'tolower($1)=="location:"{print $2}' \
        | tr -d '\r' | tail -1)
tag="${tag##*/}"

# Baixa o binário e o SHA256SUMS
curl -fsSL "https://github.com/zewillyan007/gouse/releases/download/${tag}/gouse-${tag}-linux-amd64" -o gouse
curl -fsSL "https://github.com/zewillyan007/gouse/releases/download/${tag}/SHA256SUMS" -o SHA256SUMS

# Verifica
sha256sum -c SHA256SUMS --ignore-missing
```

## Uso

```sh
gouse list-remote                 # 20 versões mais novas, latest na última linha
gouse list-remote --page 2        # 20 versões anteriores
gouse list-remote --all           # inclui RC/beta
gouse install go1.26.3            # baixa, valida SHA256, extrai
gouse install latest              # resolve para a stable mais nova
gouse install latest --default    # instala e grava como default no mesmo passo
gouse list                        # lista versões instaladas
gouse use go1.26.3                # troca no shell atual (sem reabrir)
gouse default go1.26.3            # define a versão padrão para novos shells
gouse remove go1.21.6             # remove uma versão
gouse --version                   # mostra a versão do gouse
```

### Tab-complete

O instalador deixa o tab-complete pronto em **bash, zsh e fish** com paridade total. Digite `gouse <TAB><TAB>` para listar os subcomandos. Em `gouse use <TAB>`, `gouse remove <TAB>` e `gouse default <TAB>`, o shell sugere as versões já instaladas. A mesma lista aparece nos três shells. Para ativar, basta abrir um terminal novo (ou re-sourcear o rcfile) após o install.

### Paginação do `list-remote`

A API oficial do Go retorna mais de 200 releases. O `list-remote` mostra **20 por página**, ordenadas da mais antiga (topo) para a mais nova (rodapé). A página 1 (default) é a página com as 20 versões **mais recentes** — a `(latest)` fica na última linha, sem rolagem. `--page 2` traz as 20 anteriores, e assim por diante. O rodapé indica o total de páginas.

## ⚠ Não use `~/go/gopaths/` para seus projetos

A pasta `~/go/gopaths/<versão>/` é **gerenciada pelo gouse**:

- É segregada por versão do Go e troca conforme `gouse use`.
- É **removida integralmente** pelo `uninstall.sh`.

Crie seus projetos fora dessa árvore (ex: `~/projetos/`, `~/code/`, etc.).

## Migração de instalação antiga (`/usr/local/gos/`)

A v1 do gouse instalava o Go em `/usr/local/gos/` (com sudo). A versão atual usa `~/.local/share/gos/` (sem sudo). Se você tem versões pré-existentes no path antigo:

```sh
mkdir -p ~/.local/share/gos
sudo mv /usr/local/gos/* ~/.local/share/gos/
sudo chown -R "$USER:$USER" ~/.local/share/gos
sudo rmdir /usr/local/gos
```

Pronto: `gouse list` já enxerga as versões migradas.

## Desinstalação

```sh
curl -fsSL https://raw.githubusercontent.com/zewillyan007/gouse/main/uninstall.sh | sh -s -- --yes
```

Remove (sem sudo):

- `~/.gouse/`
- A linha de source do gouse em **todos** os rcfiles conhecidos (`~/.bashrc`, `~/.zshrc`, `~/.config/fish/config.fish`)
- `~/go/gopaths/` (todos os GOPATHs)
- `~/.local/share/gos/` (todas as versões do Go)

Quando executado interativamente (sem `--yes`), mostra um resumo e pede confirmação.

## Comandos avançados

Use `gouse --help-all` para ver todos os comandos, incluindo os internos:

| Comando | Quando |
|---|---|
| `gouse init` | Regerar `~/.gouse/shell.sh` (o `install.sh` já faz isso). |
| `gouse shell-env <versão>` | Imprime exports para `eval`. Usado pela função shell. |
| `gouse default --print` | Imprime só o nome da versão default (uso interno do `shell.sh`). |

Sem a função shell carregada, `gouse use <versão>` imprime os exports no stdout — pode ser usado manualmente com `eval "$(gouse use <versão>)"`.

## Como funciona

```
~/.gouse/
├── bin/gouse              binário Go
├── shell.{sh|zsh|fish}    função gouse() + tab-complete + ativação do default no startup
├── completion.{bash|zsh|fish}    script de tab-complete (gerado pelo `gouse init`)
└── state.json             { "default": "...", "latest_known": "...", "latest_checked_at": ... }

~/.local/share/gos/<versão>/go/    extração do tarball oficial
~/go/gopaths/<versão>/             GOPATH por versão
```

(Honora `$XDG_DATA_HOME` se setado; caso contrário usa `~/.local/share/`.)

Lista de versões disponíveis vem direto da API oficial:
`https://go.dev/dl/?mode=json&include=all`. Cada download valida o SHA256 antes de extrair.

A tag `(latest)` em `gouse list` reflete a comparação com a **versão mais nova publicada online** (cache de 24h em `state.json`, atualizado por `list-remote` e `install`; refresh silencioso quando stale).

## Desenvolvimento

```sh
go build -o ./bin/gouse ./cmd/gouse
./install.sh --local      # instala a partir do build local em vez de baixar do GitHub
```

## Licença

MIT.
