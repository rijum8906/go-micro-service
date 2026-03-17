# Relay

a microservice

## Initialize Project

### Sync go work

#### Step 1: Create go work

```bash
go work init
```

#### Step 2: Add packages

```bash
go work use ./services/*
go work use ./packages/common
```

#### Step 3: Sync go work

```bash
go work sync
```
