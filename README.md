# GoQueue

Apply `db/schema.sql` before starting the app or upgrading an existing database.

```sh
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f db/schema.sql
go test -race ./...
go run .
```

## Crash-safe job output

The worker stores the job's persisted payload in `job_outputs` before marking
the job complete. The primary key is the persisted job ID. PostgreSQL writes the
output and enforces uniqueness in one statement. If the worker stops after
committing the output, startup recovery can replay the job and finish its status
update using the same ID. Failed simulations produce no output.

The guarantee covers this PostgreSQL output for a single job ID. An external
effect would require the receiving service to enforce a stable idempotency key.
Separate HTTP submissions currently create separate job IDs.
