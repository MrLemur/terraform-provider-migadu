# Terraform Provider for Migadu

Terraform provider for managing [Migadu](https://www.migadu.com/) resources.

## Documentation

Provider/resource/data source usage docs are maintained in `docs/`:

- Provider: `docs/index.md`
- Resources: `docs/resources/`
- Data sources: `docs/data-sources/`

Examples are available in `examples/`.

## Requirements

- Terraform `>= 1.0`
- Go `>= 1.25` (for local development/builds)
- A Migadu account with API access

## Local Development

Build:

```bash
make build
```

Install locally:

```bash
make install
```

Format:

```bash
make fmt
```

## Testing

### Unit tests

Runs fast tests only (no real API calls):

```bash
go test -count=1 ./...
```

or:

```bash
make test
```

### Acceptance tests

Acceptance tests run against real Migadu infrastructure.

Required environment variables:

```bash
export TF_ACC=1
export MIGADU_USERNAME="admin@example.com"
export MIGADU_API_KEY="your-api-key"
export MIGADU_TEST_DOMAIN="example.com"
```

Run all acceptance tests:

```bash
go test ./internal/provider -run '^TestAcc' -count=1 -v
```

Run specific acceptance tests:

```bash
go test ./internal/provider -run 'TestAccMailboxResource_basic' -count=1 -v
go test ./internal/provider -run 'TestAccAliasResource_basic' -count=1 -v
```

Notes:

- These tests create/update/delete real resources in `MIGADU_TEST_DOMAIN`.
- Use a dedicated test domain/account.
- Keep `TF_ACC` unset for normal local/CI unit test runs.

## License

MIT. See `LICENSE`.
