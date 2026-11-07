// The query component: filters newline-delimited JSON sensor readings, and
// optionally re-encodes them as CSV, while they stream through.
//
// It is pure compute. It imports nothing but logging, holds no connection,
// and never sees where its input came from or where its output goes. The api
// component hands it the stream it got from the store, which is the point:
// the bytes go from the store component to this one without passing through
// the api, and nothing along the way holds more than a buffer of them.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"go.bytecodealliance.org/wasi/wasilog"

	exports "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/queryexports/export_wasmcloud_archive_query"
	query "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/queryexports/wasmcloud_archive_query"
	"github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/streamio"
	witTypes "go.bytecodealliance.org/pkg/wit/types"

	// Links the //go:wasmexport glue for wasmcloud:archive/query.
	_ "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/queryexports/wit_exports"

	// Puts the SDK's WIT, and the headless world this component is built
	// with, where componentize-go looks for them.
	_ "go.wasmcloud.dev/component"
)

// The longest line accepted. A longer one fails the run rather than being
// buffered without bound.
const maxLine = 64 << 10

// reading is one line of input, as the recorder writes it.
type reading struct {
	TS     string  `json:"ts"`
	Site   string  `json:"site"`
	Sensor string  `json:"sensor"`
	Metric string  `json:"metric"`
	Value  float64 `json:"value"`
}

func init() {
	exports.Exports.Run = run
}

func run(input *witTypes.StreamReader[uint8], inputDone *exports.InputDone, filter query.Filter) (*witTypes.StreamReader[uint8], *exports.OutputDone) {
	in := streamio.NewReader(input, inputDone)

	data, dataReader := query.MakeStreamU8()
	done, doneReader := query.MakeFutureResultSummaryString()
	out := streamio.NewWriter(data, done)

	// Return the output stream straight away and fill it from a goroutine,
	// so the caller can start sending results before the input is finished.
	go func() {
		summary, err := filterReadings(out, in, filter)
		in.Close()
		if err != nil && !errors.Is(err, streamio.ErrReaderGone) {
			wasilog.ContextLogger("run").Warn("query failed", "err", err,
				"scanned", summary.Scanned, "matched", summary.Matched)
		}
		out.CloseWith(summary, err)
	}()

	return dataReader, doneReader
}

func filterReadings(dst io.Writer, src io.Reader, filter query.Filter) (query.Summary, error) {
	var summary query.Summary
	w := bufio.NewWriterSize(dst, 64<<10)

	csv := filter.Format == query.FormatCsv
	if csv {
		w.WriteString("ts,site,sensor,metric,value\n")
	}

	lines := bufio.NewScanner(src)
	lines.Buffer(make([]byte, 0, 4<<10), maxLine)

	limit := uint64(0)
	if filter.Limit.IsSome() {
		limit = uint64(filter.Limit.Some())
	}

	row := make([]byte, 0, 256)
	for lines.Scan() {
		line := lines.Bytes()
		if len(line) == 0 {
			continue
		}
		summary.Scanned++

		var r reading
		if err := json.Unmarshal(line, &r); err != nil {
			summary.Malformed++
			continue
		}
		if !matches(&r, filter) {
			continue
		}

		if csv {
			row = row[:0]
			row = append(row, r.TS...)
			row = append(row, ',')
			row = append(row, r.Site...)
			row = append(row, ',')
			row = append(row, r.Sensor...)
			row = append(row, ',')
			row = append(row, r.Metric...)
			row = append(row, ',')
			row = strconv.AppendFloat(row, r.Value, 'f', -1, 64)
			row = append(row, '\n')
			w.Write(row)
		} else {
			w.Write(line)
			w.WriteByte('\n')
		}
		summary.Matched++

		if limit > 0 && summary.Matched >= limit {
			// Done early. Returning drops the input stream, which tells the
			// store to stop reading chunks nobody is going to look at.
			return summary, w.Flush()
		}
	}
	if err := lines.Err(); err != nil {
		w.Flush()
		return summary, fmt.Errorf("reading input after %d lines: %w", summary.Scanned, err)
	}
	return summary, w.Flush()
}

func matches(r *reading, f query.Filter) bool {
	if f.Site.IsSome() && r.Site != f.Site.Some() {
		return false
	}
	if f.Sensor.IsSome() && r.Sensor != f.Sensor.Some() {
		return false
	}
	if f.MinValue.IsSome() && r.Value < f.MinValue.Some() {
		return false
	}
	return true
}

func main() {}
