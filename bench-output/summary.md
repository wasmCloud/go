### http_invoke_go @ jfleitz/wasmcloud-v2 (`139cf296d4d3`)

| scenario | n | mean | median | p90 | p99 | std dev | req/s |
|---|---|---|---|---|---|---|---|
| GET `/` | 16633 | 901.8µs | 870.0µs | 1.01ms | 1.30ms | 138.7µs | 1109 |
| POST `/post` | 16425 | 913.2µs | 879.0µs | 1.03ms | 1.33ms | 159.5µs | 1095 |

host: `jeremy-MacBook-Pro` · kernel `25.6.0` · 10 cpus · go `go1.26.5 darwin/arm64` · wash `wash 2.6.1`
