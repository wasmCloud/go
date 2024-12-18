//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world component --out gen ./wit
package main

import (
	"test-harness/gen/wasi/logging/logging"
	"test-harness/gen/wasmcloud/postgres/query"
	"test-harness/gen/wasmcloud/postgres/types"

	"go.bytecodealliance.org/cm"
)

func init() {
	logging.Exports.Log = log
	query.Exports.Query = queryFunc
	query.Exports.QueryBatch = batch
}

func log(level logging.Level, context string, message string) {
	// Noop because we're testing
}

func queryFunc(q string, params cm.List[query.PgValue]) cm.Result[query.QueryErrorShape, cm.List[types.ResultRow], types.QueryError] {
	entry := []types.ResultRowEntry{
		{
			ColumnName: "id",
			Value:      types.PgValueInt8(1),
		},
		{
			ColumnName: "description",
			Value:      types.PgValueText("hello there"),
		},
		{
			ColumnName: "created_at",
			Value: types.PgValueTimestampTz(types.TimestampTz{
				Timestamp: types.Timestamp{
					Date: types.DateYmd(cm.Tuple3[int32, uint32, uint32]{
						F0: 2024,
						F1: 12,
						F2: 17,
					}),
					Time: types.Time{
						Hour: 10,
						Min:  10,
						Sec:  45,
					},
				},
			}),
		},
	}
	results := []query.ResultRow{
		query.ResultRow(cm.ToList(entry)),
	}
	return cm.OK[cm.Result[query.QueryErrorShape, cm.List[types.ResultRow], types.QueryError]](cm.ToList(results))
}
func batch(q string) cm.Result[query.QueryError, struct{}, query.QueryError] {
	// Noop because we don't use this in the test
	return cm.OK[cm.Result[query.QueryError, struct{}, query.QueryError]](struct{}{})
}

func main() {}
