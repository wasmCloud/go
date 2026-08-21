package postgres

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	prepared "go.wasmcloud.dev/component/imports/wasmcloud_postgres_0_2_0_prepared"
	query "go.wasmcloud.dev/component/imports/wasmcloud_postgres_0_2_0_query"
	types "go.wasmcloud.dev/component/imports/wasmcloud_postgres_0_2_0_types"
)

// rowsReadChunk is how many rows a Rows iterator requests from the host per
// stream read.
const rowsReadChunk = 64

// Query runs a parameterized statement against the database. Parameters are
// positional, referenced as $1, $2, ... in the SQL text.
//
// The result rows stream back incrementally: iterate with [Rows.Next], read
// the current row with [Rows.Row], and check [Rows.Err] after Next returns
// false. Always [Rows.Close] the result (defer is fine) so the host-side
// stream is released.
func Query(sql string, params ...Value) (*Rows, error) {
	res := query.Query(sql, params)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	ok := res.Ok()
	return &Rows{
		columns: ok.F0,
		stream:  ok.F1,
		done:    ok.F2,
	}, nil
}

// Rows is a streaming query result. The zero value is not usable; obtain one
// from [Query].
type Rows struct {
	columns []string
	stream  *witTypes.StreamReader[[]types.PgValue]
	done    *witTypes.FutureReader[witTypes.Result[witTypes.Unit, types.Error]]

	buf []([]types.PgValue)
	cur []types.PgValue
	err error
	eof bool
}

// Columns returns the result column names, in order. Position i names the
// value at position i of every row.
func (r *Rows) Columns() []string {
	return r.columns
}

// Next advances to the next row, blocking until one is available. It returns
// false when the result set is exhausted or a late error occurred; check
// [Rows.Err] afterwards.
func (r *Rows) Next() bool {
	if len(r.buf) == 0 && !r.eof {
		chunk := make([]([]types.PgValue), rowsReadChunk)
		n := r.stream.Read(chunk)
		if n == 0 {
			r.eof = true
			r.finish()
		} else {
			r.buf = chunk[:n]
		}
	}
	if len(r.buf) == 0 {
		return false
	}
	r.cur = r.buf[0]
	r.buf = r.buf[1:]
	return true
}

// Row returns the row most recently advanced to by [Rows.Next]. Values are
// positional, matching [Rows.Columns].
func (r *Rows) Row() []Value {
	return r.cur
}

// Err returns the first error encountered while streaming rows, including
// failures reported after the column header was sent (e.g. partway through a
// large result). It is only complete after [Rows.Next] has returned false.
func (r *Rows) Err() error {
	return r.err
}

// Close releases the host-side stream without necessarily draining it. It is
// safe to call multiple times.
func (r *Rows) Close() {
	if r.stream != nil {
		r.stream.Drop()
		r.stream = nil
	}
	if r.done != nil {
		r.done.Drop()
		r.done = nil
	}
	r.buf = nil
	r.eof = true
}

// finish reads the completion future once the row stream is exhausted, so
// errors that occur mid-stream are surfaced via Err.
func (r *Rows) finish() {
	if r.done == nil {
		return
	}
	result := r.done.Read()
	r.done = nil
	if result.IsErr() {
		r.err = convertError(result.Err())
	}
}

// QueryBatch runs a batch query, which may contain multiple statements
// (common in migrations). Parameters are not allowed: never build the query
// text from user-provided or untrusted data.
func QueryBatch(sql string) error {
	res := query.QueryBatch(sql)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// PreparedStatement is a statement prepared once and executable repeatedly.
type PreparedStatement struct {
	token string
}

// Prepare prepares a parameterized statement ($1, $2, ...) against the
// database and returns a handle for later execution.
func Prepare(sql string) (*PreparedStatement, error) {
	res := prepared.Prepare(sql)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &PreparedStatement{token: res.Ok()}, nil
}

// Token returns the host-issued opaque token for the prepared statement.
func (p *PreparedStatement) Token() string {
	return p.token
}

// Exec executes the prepared statement with the given parameters and returns
// the number of rows affected.
func (p *PreparedStatement) Exec(params ...Value) (uint64, error) {
	res := prepared.Exec(p.token, params)
	if res.IsErr() {
		return 0, convertError(res.Err())
	}
	return res.Ok(), nil
}
