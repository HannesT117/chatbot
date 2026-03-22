# Commands

## Go server

### Run dev

```sh
go run ./cmd/server/
```

### Test

```sh
go test ./...
```

### Test (all, including live LLM integration)

```sh
CHATBOT_LIVE_TESTS=1 go test ./...
```

### Typecheck / build check

```sh
go build ./...
```

### Vet

```sh
go vet ./...
```

## Next.js frontend (web/)

### Run dev

```sh
cd web && npm run dev
```

### Test

```sh
cd web && npm test
```

### Typecheck

```sh
cd web && npx tsc --noEmit
```

### Lint

```sh
cd web && npm run lint
```
