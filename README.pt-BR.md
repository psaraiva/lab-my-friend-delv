# lab-my-friend-delv

[![project](https://img.shields.io/badge/github-psaraiva%2Flab--my--friend--delv-blue)](https://github.com/psaraiva/lab-my-friend-delv)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[![Go Report Card](https://goreportcard.com/badge/github.com/psaraiva/lab-my-friend-delv?style=flat)](https://goreportcard.com/report/github.com/psaraiva/lab-my-friend-delv)
[![Language: English](https://img.shields.io/badge/Idioma-English-blue?style=flat-square)](./README.md)


Um playground Go para explorar o **Delve (`dlv`)** — o debugger oficial do Go.

O cenário a ser explorado é sistema distribuído simples com uma CLI, dois serviços HTTP, lógica de negócio com pesos, chamadas HTTP entre serviços e concorrência com mutex. Essa variedade de padrões oferece um ambiente rico para praticar os recursos do Delve.

O sistema é composto por três aplicações independentes que se comunicam via HTTP:

```
App01 (CLI)  -->  App02 (Lançador :9022)  -->  App03 (Gerenciador :9023)
```

---

## Aplicações

### App01 — CLI Lançador de Dados

Interface de linha de comando construída com **Cobra**. Recebe o nome de um dado (1 a 25 caracteres alfanuméricos), chama o App02 e exibe o resultado de forma amigável.

```bash
cd app01
go run main.go <nome-do-dado>

# Exemplo
go run main.go dragao
```

**Saída:**
```
----------------------------------------
  Dado   : dragao (D16)
  Faces  : 16
  Resultado --> 4
----------------------------------------
```

---

### App02 — Servidor de Lançamento (`:9022`)

Servidor HTTP construído com **Gin**. Expõe um único endpoint que lança um dado pelo nome usando probabilidade ponderada. Consulta o App03 para obter a configuração do dado antes de lançar.

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/roll/:name` | Lança o dado pelo nome e retorna o resultado |

**Probabilidade ponderada**

| Dado | Pesos em destaque |
|------|-------------------|
| D6  | Faces 2 e 5: 20% &bull; demais: 15% |
| D12 | Todas as faces: ~8,33% (uniforme) |
| D16 | Face 4: 8,25% &bull; Face 10: 7,25% &bull; Face 11: 4,25% &bull; Face 16: 5,25% &bull; demais: 6,25% |

Os pesos são totalmente parametrizados no mapa `diceConfigs` em [app02/main.go](app02/main.go).

---

### App03 — Gerenciador de Dados (`:9023`)

Servidor HTTP construído com **Gin** que gerencia o cadastro de dados (em memória). Oferece CRUD completo para os registros de dados.

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/dices` | Cadastra um novo dado `{ name, sides }` |
| `GET` | `/dices` | Lista todos os dados cadastrados |
| `GET` | `/dices/:name` | Retorna um dado pelo nome |
| `DELETE` | `/dices/:name` | Remove um dado pelo nome |

**Regras de validação:**
- `name`: 1 a 25 caracteres alfanuméricos (`a-Z`, `0-9`)
- `sides`: deve ser `6`, `12` ou `16`

---

## Como Executar

### 1. Iniciar o App03 (gerenciador)

```bash
cd app03
go run main.go
```

### 2. Iniciar o App02 (lançador)

```bash
cd app02
go run main.go
```

### 3. Cadastrar um dado

```bash
curl -X POST http://localhost:9023/dices \
  -H "Content-Type: application/json" \
  -d '{"name":"dragao","sides":16}'
```

### 4. Lançar via CLI

```bash
cd app01
go run main.go dragao
```

---

## Depuração com Delve (dlv)

O nome do projeto é uma homenagem ao **[Delve](https://github.com/go-delve/delve)** — o debugger oficial do Go. O `dlv` integra nativamente com o VS Code e torna a depuração de serviços Go simples e eficiente.

### Instalar o Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Verifique a instalação:

```bash
dlv version
```

### Executar manualmente com dlv

```bash
# Depurar o App03
cd app03
dlv debug main.go

# Depurar o App02
cd app02
dlv debug main.go

# Depurar o App01 passando argumento
cd app01
dlv debug main.go -- dragao
```

---

## Configurações de Debug no VS Code

Cada aplicação possui seu próprio [`.vscode/launch.json`](.vscode/launch.json) com uma configuração de depuração pronta para uso. Abra o painel **Run and Debug** (`Ctrl+Shift+D`) dentro da pasta da aplicação e selecione o alvo desejado:

| Configuração | Descrição |
|--------------|-----------|
| `App01 - CLI` | Executa o App01 com o argumento `dado20lados` |
| `App02 - Lançador de Dados` | Inicia o App02 na porta `9022` |
| `App03 - Gerenciador de Dados` | Inicia o App03 na porta `9023` |

> **Dica:** Inicie o App03 primeiro, depois o App02 e por último o App01. O painel de depuração permite executar múltiplas configurações simultaneamente.

---

## REST Client (Huachao Mao)

Cada aplicação servidor possui seu próprio [`.cli-rest/app.http`](.cli-rest/app.http) com requisições HTTP prontas para uso cobrindo **todos os endpoints** do App02 e App03, escritas no formato da extensão [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) para VS Code, desenvolvida por **Huachao Mao**.

### Instalar a extensão

Pesquise por `REST Client` no painel de extensões do VS Code, ou instale via linha de comando:

```bash
code --install-extension humao.rest-client
```

### Como usar

Abra o arquivo `app02/.cli-rest/app.http` ou `app03/.cli-rest/app.http` no VS Code. Um link **Send Request** aparecerá acima de cada bloco `###`. Clique para executar a requisição e ver a resposta em um painel lateral.

**Requisições incluídas:**

*App03:*
- Cadastrar dados D6, D12 e D16
- Listar todos os dados
- Buscar dado pelo nome
- Remover dado pelo nome

*App02:*
- Lançar dado D6, D12 e D16 pelo nome

---

## Estrutura do Projeto

```
my-friend-delv/
├── app01/
│   ├── .vscode/
│   │   └── launch.json     # Configuração de debug com Delve
│   ├── main.go             # Ponto de entrada da CLI
│   └── go.mod
├── app02/
│   ├── .cli-rest/
│   │   └── app.http        # Requisições REST Client (App02)
│   ├── .vscode/
│   │   └── launch.json     # Configuração de debug com Delve
│   ├── main.go             # Servidor de lançamento + probabilidade ponderada
│   └── go.mod
└── app03/
    ├── .cli-rest/
    │   └── app.http        # Requisições REST Client (App03)
    ├── .vscode/
    │   └── launch.json     # Configuração de debug com Delve
    ├── main.go             # Servidor gerenciador de dados
    └── go.mod
```

---

## Requisitos

- Go `1.25.7`+
- [Delve](https://github.com/go-delve/delve) (`dlv`) para depuração
- VS Code com a extensão [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) (recomendado)
