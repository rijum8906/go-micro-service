# rcli

`rcli` is the Relay workspace CLI. The Cobra root command is currently named `my-cli`, but the source also invokes the binary as `rcli`, so the examples below use `rcli`.

## Command Tree

```text
rcli
├── project
│   ├── init (alias: setup)
│   │   └── db [--use-atlas]
│   └── sync
├── init
│   ├── project
│   └── service
├── dev
│   ├── start (aliases: run, up)
│   └── install
├── db
│   ├── init
│   ├── exec [--test]
│   │   ├── apply
│   │   ├── clean
│   │   └── truncate
│   └── migrate
│       ├── apply (alias: up)
│       ├── sql-apply
│       ├── schema
│       ├── clean
│       ├── status
│       ├── create --name <name> (alias: new)
│       ├── rehash
│       ├── rollback --count <n>
│       └── reset
└── test [--verbose] [--run <regexp>]
    ├── setup
    ├── run
    │   ├── unit
    │   ├── integration
    │   └── bench
    ├── cover
    └── up
```

## Root Command

`rcli` manages local Relay development workflows. It wires together project setup, local development helpers, database commands, and test helpers.

## `project`

These commands are intended to run from the repository root. `project` has a persistent pre-run check that errors unless the current directory is the Relay root.

### `rcli project init`

Initializes the workspace:

- runs `go work init`
- adds the root, all `packages/*`, and all `services/*` directories to the workspace
- runs `go work sync`
- runs `go mod download`
- copies each service `.env.example` to `.env` if the target does not already exist

Alias: `rcli project setup`

### `rcli project init db`

Initializes databases for every discovered service.

- runs `rcli db init` inside each service directory
- runs `rcli db exec apply` by default
- runs `rcli db migrate apply` instead when `--use-atlas` or `-a` is set

Flag:

- `--use-atlas`, `-a`: use Atlas migrations instead of raw schema SQL

### `rcli project sync`

Refreshes the workspace after new services or packages are added:

- re-adds workspace directories
- runs `go work sync`
- runs `go mod download`
- recopies missing `.env` files

## `init`

This namespace overlaps with `project init`, but is implemented separately.

### `rcli init project`

Initializes the project by creating a Go workspace, syncing modules, downloading dependencies, and copying service `.env.example` files to `.env`.

### `rcli init service`

Declared but not implemented yet.

## `dev`

### `rcli dev start`

Runs `docker compose up --build`.

Aliases: `rcli dev run`, `rcli dev up`

### `rcli dev install`

Installs local development dependencies:

- installs `sqlc` via `go install`
- installs `atlas` via the Atlas curl installer on Linux and macOS

## `db`

Most `db` commands expect to be run from an individual service directory because they load service-local `.env` files and/or read `db/schema.sql`.

### `rcli db init`

Creates the configured database if it does not already exist.

### `rcli db exec`

Runs direct SQL/schema operations without Atlas.

Persistent flag:

- `--test`, `-t`: connect to the hard-coded test database (`test_db` on port `5433`) instead of the service database

Subcommands:

- `rcli db exec apply`: executes `db/schema.sql` against the selected database
- `rcli db exec clean`: drops and recreates the `public` schema
- `rcli db exec truncate`: drops and recreates the `public` schema, then reapplies `db/schema.sql`

### `rcli db migrate`

Runs Atlas-based migration workflows.

Subcommands:

- `rcli db migrate apply`: runs `atlas migrate apply`
- `rcli db migrate up`: alias for `apply`
- `rcli db migrate sql-apply`: creates the configured app and dev databases if needed, then executes `db/schema.sql`
- `rcli db migrate schema`: runs `atlas schema apply --auto-approve`
- `rcli db migrate clean`: runs `atlas schema clean`
- `rcli db migrate status`: runs `atlas migrate status`
- `rcli db migrate create --name <name>`: runs `atlas migrate diff <name>`
- `rcli db migrate new --name <name>`: alias for `create`
- `rcli db migrate rehash`: runs `atlas migrate rehash`
- `rcli db migrate rollback --count <n>`: runs `atlas migrate down <n>`
- `rcli db migrate reset`: declared but not implemented yet

Required flags:

- `rcli db migrate create --name`, `-n`
- `rcli db migrate rollback --count`, `-c`

## `test`

Persistent flags apply to all `test` subcommands:

- `--verbose`, `-v`: pass verbose mode to `go test` helpers where supported
- `--run`, `-r`: pass a regexp filter to `go test` helpers where supported

### `rcli test setup`

Prepares the hard-coded test database (`test_db` on port `5433`) by applying `db/schema.sql`.

### `rcli test run`

Runs `go test ./... -race -cover -coverprofile=coverage.out -covermode=atomic`.

Nested subcommands:

- `rcli test run unit`: runs short tests with race detection
- `rcli test run integration`: runs integration-tagged tests
- `rcli test run bench`: runs benchmarks only

### `rcli test cover`

Generates `coverage.out` and opens the HTML coverage report using `go tool cover -html=coverage.out`.

### `rcli test up`

Runs `docker compose -f docker-compose.test.yml up --build`.
