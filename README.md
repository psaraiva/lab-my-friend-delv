# lab-my-friend-delv

[![project](https://img.shields.io/badge/github-psaraiva%2Flab--my--friend--delv-blue)](https://github.com/psaraiva/lab-my-friend-delv)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[![Go Report Card](https://goreportcard.com/badge/github.com/psaraiva/lab-my-friend-delv?style=flat)](https://goreportcard.com/report/github.com/psaraiva/lab-my-friend-delv)
[![Language: Português](https://img.shields.io/badge/Language-Portugu%C3%AAs-green?style=flat-square)](./README.pt-BR.md)

A Go playground for exploring **Delve (`dlv`)** — the Go debugger.

The scenario to explore is a simple distributed system with a CLI, two HTTP services, weighted business logic, inter-service HTTP calls, and concurrency via mutexes. This variety of patterns provides a rich environment to practice Delve's features.

The system is composed of three independent applications that communicate via HTTP:

```
App01 (CLI)  -->  App02 (Roller :9022)  -->  App03 (Manager :9023)
```

---

## Applications

### App01 — CLI Dice Roller

A command-line interface built with **Cobra**. Accepts a dice name (1–25 alphanumeric characters), calls App02, and displays the result in a friendly format.

```bash
cd app01
go run main.go <dice-name>

# Example
go run main.go dragao
```

**Output:**
```
----------------------------------------
  Dice   : dragao (D16)
  Sides  : 16
  Result --> 4
----------------------------------------
```

---

### App02 — Dice Rolling Server (`:9022`)

An HTTP server built with **Gin**. Exposes a single endpoint that rolls a named die using weighted probability. It queries App03 to retrieve the die configuration before rolling.

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/roll/:name` | Rolls the named die and returns the result |

**Weighted probability**

| Die | Notable weights |
|-----|----------------|
| D6  | Faces 2 and 5: 20% &bull; others: 15% |
| D12 | All faces: ~8.33% (uniform) |
| D16 | Face 4: 8.25% &bull; Face 10: 7.25% &bull; Face 11: 4.25% &bull; Face 16: 5.25% &bull; others: 6.25% |

Weights are fully parameterized in the `diceConfigs` map in [app02/main.go](app02/main.go).

---

### App03 — Dice Manager (`:9023`)

An HTTP server built with **Gin** that manages the dice registry (in-memory). Provides full CRUD for dice records.

| Method | Route | Description |
|--------|-------|-------------|
| `POST` | `/dices` | Register a new die `{ name, sides }` |
| `GET` | `/dices` | List all registered dice |
| `GET` | `/dices/:name` | Get a die by name |
| `DELETE` | `/dices/:name` | Remove a die by name |

**Validation rules:**
- `name`: 1–25 alphanumeric characters (`a-Z`, `0-9`)
- `sides`: must be `6`, `12`, or `16`

---

## Getting Started

### 1. Start App03 (dice manager)

```bash
cd app03
go run main.go
```

### 2. Start App02 (dice roller)

```bash
cd app02
go run main.go
```

### 3. Register a die

```bash
curl -X POST http://localhost:9023/dices \
  -H "Content-Type: application/json" \
  -d '{"name":"dragao","sides":16}'
```

### 4. Roll it via CLI

```bash
cd app01
go run main.go dragao
```

---

## Debugging with Delve (dlv)

The project name is a nod to **[Delve](https://github.com/go-delve/delve)** — the Go debugger. `dlv` integrates seamlessly with VS Code and makes debugging Go services straightforward.

### Install Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Verify the installation:

```bash
dlv version
```

### Run manually with dlv

```bash
# Debug App03
cd app03
dlv debug main.go

# Debug App02
cd app02
dlv debug main.go

# Debug App01 passing an argument
cd app01
dlv debug main.go -- dragao
```

---

## VS Code Debug Configurations

Each application has its own [`.vscode/launch.json`](.vscode/launch.json) with a ready-to-use debug configuration. Open the **Run and Debug** panel (`Ctrl+Shift+D`) inside the app's folder and select the target:

| Configuration | Description |
|---------------|-------------|
| `App01 - CLI` | Runs App01 with the argument `dado20lados` |
| `App02 - Lançador de Dados` | Starts App02 on port `9022` |
| `App03 - Gerenciador de Dados` | Starts App03 on port `9023` |

> **Tip:** Start App03 first, then App02, and finally App01. The debug panel allows running multiple configurations simultaneously.

---

## REST Client (Huachao Mao)

Each server application has its own [`.cli-rest/app.http`](.cli-rest/app.http) file with ready-to-use HTTP requests covering **all endpoints** of App02 and App03, written in the format of the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) VS Code extension by **Huachao Mao**.

### Install the extension

Search for `REST Client` in the VS Code Extensions panel, or install via the command line:

```bash
code --install-extension humao.rest-client
```

### How to use

Open `app02/.cli-rest/app.http` or `app03/.cli-rest/app.http` in VS Code. A **Send Request** link will appear above each `###` block. Click it to execute the request and see the response in a side panel.

**Included requests:**

*App03:*
- Register D6, D12, and D16 dice
- List all dice
- Get a die by name
- Delete a die by name

*App02:*
- Roll a D6, D12, and D16 die by name

---

## Project Structure

```
lab-my-friend-delv/
├── app01/
│   ├── .vscode/
│   │   └── launch.json     # Delve debug configuration
│   ├── main.go             # CLI entry point
│   └── go.mod
├── app02/
│   ├── .cli-rest/
│   │   └── app.http        # REST Client requests (App02)
│   ├── .vscode/
│   │   └── launch.json     # Delve debug configuration
│   ├── main.go             # Rolling server + weighted probability
│   └── go.mod
└── app03/
    ├── .cli-rest/
    │   └── app.http        # REST Client requests (App03)
    ├── .vscode/
    │   └── launch.json     # Delve debug configuration
    ├── main.go             # Dice manager server
    └── go.mod
```

---

## Requirements

- Go `1.25.7`+
- [Delve](https://github.com/go-delve/delve) (`dlv`) for debugging
- VS Code with the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension (recomendado)
