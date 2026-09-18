# Raw Valkey client

`Open` returns the application-owned `valkey-go` client. Callers issue commands
directly with their request context:

```go
result := client.Do(ctx, client.B().Get().Key(key).Build())
```

The configured timeout bounds each connection attempt and handshake. On an
established pipeline, `valkey-go` detects an unresponsive connection after its
keepalive interval plus its I/O timeout, normally about twice the configured
value. A caller context controls the logical operation deadline, and blocking
commands need an explicit deadline. During a lazy connection or reconnect
handshake, early cancellation can return only when the caller's existing deadline
or the configured Valkey timeout closes the socket. Concurrent callers waiting on
that shared setup inherit the same bound. Streaming commands retain `valkey-go`'s
native cancellation behavior. `Client.Close` also retains the upstream shutdown
behavior and can wait up to its native one-second close allowance per blackholed
connection.
Valkey is not part of readiness. An initial connection error is degraded startup
when `Open` also returns a client, because that client reconnects on later calls.
