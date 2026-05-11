# Contexto: Gerenciador de versões do Go (`gouse`)

## Situação

- SO: Fedora
- Linguagem: Go
- Versão usada no trabalho: `go1.21.6`
- Versão usada em projetos pessoais: `latest`
- Não quero usar `gvm` (última modificação há ~3 anos, projeto parado).
- Objetivo: trocar de versão do Go no shell **sem reabrir a sessão**.

## Estrutura de diretórios no sistema

As versões do Go estão instaladas em:

```
/usr/local/gos/<versao>/go/bin/go
```

Os GOPATHs ficam separados por versão em:

```
$HOME/go/gopaths/<versao>
```

Exemplos de `<versao>`: `go1.21.6`, `go-latest`.

## Configuração antiga (a ser removida)

No `.bash_profile` havia exports diretos como:

```bash
# Go latest
#export GOPATH=$HOME/go/gopaths/go-latest
#export PATH=$PATH:$GOPATH/bin:/usr/local/gos/go-latest/go/bin

# Go 1.21.6
export GOPATH=$HOME/go/gopaths/go1.21.6
export PATH=$PATH:$GOPATH/bin:/usr/local/gos/go1.21.6/go/bin
```

Problema: pra trocar de versão era preciso editar o arquivo e reabrir a sessão. Essas linhas foram removidas.

## Solução: função `gouse` em `~/.gouse`

Criar o arquivo `~/.gouse` com o conteúdo:

```bash
# ~/.gouse - gerenciador simples de versões do Go

export GO_VERSIONS_DIR=/usr/local/gos
export GO_GOPATHS_DIR=$HOME/go/gopaths

gouse() {
    local version=$1
    if [ -z "$version" ]; then
        echo "Uso: gouse <versão>   (ex: gouse go1.21.6, gouse go-latest)"
        echo "Versão atual: ${GO_CURRENT:-nenhuma}"
        echo "Disponíveis:"
        ls -1 "$GO_VERSIONS_DIR" 2>/dev/null | sed 's/^/  /'
        return 0
    fi

    local go_bin="$GO_VERSIONS_DIR/$version/go/bin"
    local gopath="$GO_GOPATHS_DIR/$version"

    if [ ! -d "$go_bin" ]; then
        echo "Erro: $go_bin não existe"
        return 1
    fi

    # Remove do PATH a versão anterior que esta função adicionou
    if [ -n "$GO_CURRENT" ]; then
        local old_go_bin="$GO_VERSIONS_DIR/$GO_CURRENT/go/bin"
        local old_gopath_bin="$GO_GOPATHS_DIR/$GO_CURRENT/bin"
        PATH=$(echo "$PATH" | tr ':' '\n' | grep -vFx "$old_go_bin" | grep -vFx "$old_gopath_bin" | paste -sd:)
    fi

    export GOPATH="$gopath"
    export PATH="$go_bin:$gopath/bin:$PATH"
    export GO_CURRENT="$version"

    echo "Go $version ativado"
    go version
}

# Versão padrão ao carregar (opcional)
gouse go1.21.6 >/dev/null
```

## Carregamento no `~/.bashrc`

Adicionar **no final** do `~/.bashrc` (não no `.bash_profile`):

```bash
# gouse - gerenciador de versões do Go
[ -f ~/.gouse ] && source ~/.gouse
```

Justificativa:
- No Fedora, o `.bash_profile` padrão já faz `source ~/.bashrc`, então login shells e shells não-login (tmux, `bash` dentro de outro shell) ficam cobertos.
- Gerenciadores de versão (nvm, gouse, etc.) devem ser carregados **depois** de outras manipulações de PATH, pra garantir prioridade das versões gerenciadas.

## Detalhes de design

### `GO_CURRENT`
Variável usada para rastrear qual versão a função adicionou ao PATH. Sem ela, o PATH acumularia lixo a cada troca. Na primeira chamada está vazia → o `if` é pulado → comportamento correto (nada pra limpar ainda).

### Ordem no PATH
`go_bin` e `gopath/bin` são adicionados no **início** do PATH, garantindo que o `go` chamado seja sempre o da versão ativa, mesmo se houver outro Go instalado via `dnf` ou outro caminho.

### Versão padrão ao iniciar
Sem a linha `gouse go1.21.6 >/dev/null` no fim do `~/.gouse`, ao abrir um terminal novo o `go` não estará no PATH até rodar `gouse <versão>` manualmente.

Alternativa: persistir a última versão usada entre sessões.

Dentro da função `gouse`, adicionar antes do `echo` final:

```bash
echo "$version" > ~/.gouse_last
```

E no fim do `~/.gouse`, substituir o `gouse go1.21.6 >/dev/null` por:

```bash
if [ -f ~/.gouse_last ]; then
    gouse "$(cat ~/.gouse_last)" >/dev/null
else
    gouse go1.21.6 >/dev/null
fi
```

## Uso

```bash
gouse              # mostra versão atual e lista disponíveis
gouse go-latest    # troca pra latest
gouse go1.21.6     # volta pra 1.21.6
go version         # confirma versão ativa
```

Para ativar sem reabrir terminal após criar/editar:

```bash
source ~/.gouse
# ou
source ~/.bashrc
```
