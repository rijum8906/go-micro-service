# Relay

## Initialize Project

### Step 1: Create go work

```bash
go work init
```

### Step 2: Add packages

```bash
go work use ./services/user
go work use ./services/graphql-gateway
go work use ./packages/core
go work use ./packages/pb
```

### Step 3: Sync go work

```bash
go work sync

```

### Step 4: Install go dependencies

```bash
go mod download
```

### Step 5: Run setup targets

```bash
make setup SERVICE=user
make setup SERVICE=graphql-gateway
```
