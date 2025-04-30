# Oracle-DB-Connection-Validation

Add a Ping for a connection to validate and monitor connection/network health.

# intro

The connection pool is managed automatically by the `sql.DB` type from the standard library. 
When you call `sql.Open`, it does **not** immediately create any connections; instead, it prepares a database handle that internally manages a pool of connections as needed[1][6]. 

Whenever you perform operations like `db.Query`, `db.Exec`, or `db.PingContext`, the `sql.DB` pool will either reuse an existing idle connection or open a new one if allowed by the pool limits[1][2][6]. The pool is thread-safe and designed for concurrent use by multiple goroutines[1][2].

You can further configure the pool’s behavior using methods such as:

- `db.SetMaxOpenConns(n)` to set the maximum number of open connections.
- `db.SetMaxIdleConns(n)` to set the maximum number of idle connections.
- `db.SetConnMaxLifetime(d)` and `db.SetConnMaxIdleTime(d)` to control connection lifetimes[5][6].

**Summary Table: Where is the pool managed?**

| Code Element      | Pool Management Location         |
|-------------------|---------------------------------|
| `sql.DB` handle   | Manages the pool internally     |
| `sql.Open`        | Prepares the pool, no connections yet |
| `db.PingContext`  | Uses a connection from the pool |
| Pool configuration| Methods on `*sql.DB`            |

You do **not** need to manually create or manage the pool; the Go standard library does this for you as long as you use the `*sql.DB` handle for your database operations[1][6].

## Example: Connection Validation

```go
func (conn *Connection) IsValid() bool {
	const validationTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), validationTimeout)
	defer cancel()

	start := time.Now()
	err := conn.Ping(ctx)
	metrics.OraclePingResponseTime.Observe(time.Since(start).Seconds())
	if err != nil {
		// conn.tracer.Print("IsValid ping failed:", err)
		metrics.OraclePingErrCounter.Inc()
		return false
	}
	return true
}

```

## Metrics

```go

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OraclePingErrCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oracle_db_connection_validation_ping_err",
	})

	OraclePingResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "oracle_db_connection_validation_ping_second",
		Buckets: []float64{0.0001, 0.0002, 0.0005, 0.001, 0.002, 0.005, 0.010, 0.020, 0.050, 0.1, 0.2, 0.5, 1, 2, 5}, // 16 buckets
	})
)
```

## Environment

```sh
export MY_ORACLE_DSN = "oracle://ORACLE_USER:ORACLE_PASSWORD@ORACLE_SERVER/ORACLE_SERVICE_NAME?TRACE FILE=trace.log"
```



# Citations
1.  https://go.dev/doc/database/manage-connections  
2.  https://koho.dev/understanding-go-and-databases-at-scale-connection-pooling-f301e56fa73  
3.  https://shahidyousuf.com/database-connection-pooling-example-in-golang  
4.  https://www.reddit.com/r/golang/comments/trvltc/best_practice_for_handling_db_connection/  
5.  https://app.studyraid.com/en/read/11522/361998/connection-pool-tuning  
6.  https://go.dev/doc/database/  
7.  https://hexacluster.ai/postgresql/postgresql-client-side-connection-pooling-in-golang-using-pgxpool/  
8.  https://stackoverflow.com/questions/75326111/understanding-sql-connection-pool-work-with-go  
