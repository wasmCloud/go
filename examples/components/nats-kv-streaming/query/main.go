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
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"go.bytecodealliance.org/wasi/wasilog"

	exports "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/queryexports/exports/export_wasmcloud_archive_query"
	query "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/queryexports/imports/wasmcloud_archive_query"
	"github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/streamio"
	witTypes "go.bytecodealliance.org/pkg/wit/types"

	// Links the //go:wasmexport glue for wasmcloud:archive/query.
	_ "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/queryexports/exports/wit_exports"

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

	emit := func(line []byte, _ *reading) error {
		if _, err := w.Write(line); err != nil {
			return err
		}
		return w.WriteByte('\n')
	}
	if filter.Format == query.Format_Csv {
		// encoding/csv quotes a field that holds a comma, quote or newline,
		// which a reading uploaded from elsewhere may well contain.
		out := csv.NewWriter(w)
		emit = func(_ []byte, r *reading) error {
			if err := out.Write([]string{r.TS, r.Site, r.Sensor, r.Metric, strconv.FormatFloat(r.Value, 'f', -1, 64)}); err != nil {
				return err
			}
			out.Flush()
			return out.Error()
		}
		if err := out.Write([]string{"ts", "site", "sensor", "metric", "value"}); err != nil {
			return summary, err
		}
	}

	lines := bufio.NewScanner(src)
	lines.Buffer(make([]byte, 0, 4<<10), maxLine)

	// Stop once the limit is reached, including a limit of zero. Returning
	// drops the input stream, which tells the store to stop reading chunks
	// nobody is going to look at.
	limited := filter.Limit.IsSome()
	limit := uint64(0)
	if limited {
		limit = uint64(filter.Limit.Some())
	}

	for (!limited || summary.Matched < limit) && lines.Scan() {
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

		// A failed write means the reader of the output has gone. Stop
		// rather than scan, and make the store fetch, the rest of the input.
		if err := emit(line, &r); err != nil {
			return summary, err
		}
		summary.Matched++
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
