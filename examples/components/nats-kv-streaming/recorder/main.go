// The recorder: the workload's long-running service.
//
// A component runs once per call and is gone. A service is started with the
// workload, runs for as long as it does, and can call any component in the
// same workload. That makes it the place for scheduled work. Every interval,
// the recorder generates a fresh export of sensor readings — 100,000 records,
// about 10 MB of newline-delimited JSON — and streams it into the store under
// one name, replacing the previous export.
//
// The export is never in memory. Records are written straight into the stream
// the store is reading, through a 64 KiB buffer, so the service's footprint is
// the same at 10 MB as it would be at 10 GB.
//
// It is configured through the workload's environment:
//
//	RECORDER_OBJECT    name to store the export under    (default "readings")
//	RECORDER_RECORDS   readings per export               (default 100000)
//	RECORDER_INTERVAL  time between exports              (default 5m)
package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.bytecodealliance.org/wasi/wasilog"
	"go.wasmcloud.dev/component/sleep"

	store "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/client/imports/wasmcloud_archive_store"
	"github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/streamio"
	runexports "github.com/wasmCloud/go/examples/components/nats-kv-streaming/recorder/exports/export_wasi_cli_run"

	// Links the //go:wasmexport glue for wasi:cli/run@0.3.0.
	_ "github.com/wasmCloud/go/examples/components/nats-kv-streaming/recorder/exports/wit_exports"

	// Puts the SDK's WIT, and the headless world this component is built
	// with, where componentize-go looks for them.
	_ "go.wasmcloud.dev/component"
)

const (
	sites          = 8
	sensorsPerSite = 64

	// After a failed export, try again sooner than a full interval: the usual
	// cause is the bucket not existing yet, which an operator fixes in
	// seconds.
	retryAfter = 10 * time.Second
)

func init() {
	runexports.Exports.Run = run
}

func run() witTypes.Result[witTypes.Unit, witTypes.Unit] {
	log := wasilog.ContextLogger("recorder")

	name := envString("RECORDER_OBJECT", "readings")
	records := envInt("RECORDER_RECORDS", 100_000)
	interval := envDuration("RECORDER_INTERVAL", 5*time.Minute)
	log.Info("recorder started", "object", name, "records", records, "interval", interval.String())

	for generation := uint64(1); ; generation++ {
		started := time.Now()
		info, err := export(name, records, generation)
		if err != nil {
			log.Error("export failed", "object", name, "err", err, "retry_in", retryAfter.String())
			sleep.Sleep(retryAfter)
			continue
		}
		log.Info("export stored",
			"object", info.Name, "records", records, "bytes", info.Size,
			"chunks", info.Chunks, "sha256", info.Sha256,
			"elapsed", time.Since(started).Round(time.Millisecond).String())

		// time.Sleep would park the goroutine on a Go runtime timer, which
		// traps inside an async export. sleep.Sleep waits on the host clock.
		sleep.Sleep(interval)
	}
}

// export generates one export and streams it into the store.
func export(name string, records int, generation uint64) (store.ObjectInfo, error) {
	data, dataReader := store.MakeStreamU8()
	done, doneReader := store.MakeFutureResultUnitString()
	out := streamio.NewWriter(data, done)

	// The generator and the Put run at the same time: Put reads what the
	// goroutine writes, and each write waits for the read before it, so
	// nothing accumulates on either side.
	go func() {
		buf := bufio.NewWriterSize(out, 64<<10)
		err := generate(buf, records, generation, time.Now().UTC())
		if err == nil {
			err = buf.Flush()
		}
		streamio.Close(out, err)
	}()

	res := store.Put(name, "application/x-ndjson", dataReader, doneReader)
	if res.IsErr() {
		return store.ObjectInfo{}, storeError(res.Err())
	}
	return res.Ok(), nil
}

// generate writes records readings, spread over the hour before now, one JSON
// object per line:
//
//	{"ts":"2026-11-07T19:00:00.036Z","site":"plant-03","sensor":"plant-03-s041","metric":"temp_c","value":48.27}
//
// Each sensor has its own baseline and a daily swing. A few sensors in each
// export run hot, so there is always something above 95 °C to query for.
func generate(w io.Writer, records int, generation uint64, now time.Time) error {
	rng := rand.New(rand.NewPCG(generation, uint64(now.UnixNano())))
	sensors := sites * sensorsPerSite

	baseline := make([]float64, sensors)
	for i := range baseline {
		baseline[i] = 25 + rng.Float64()*35
	}
	for range 3 {
		baseline[rng.IntN(sensors)] = 92 + rng.Float64()*6
	}

	window := time.Hour
	step := window / time.Duration(records)
	line := make([]byte, 0, 160)
	for i := range records {
		s := i % sensors
		ts := now.Add(-window + time.Duration(i)*step)
		hours := float64(ts.UnixMilli()) / float64(time.Hour/time.Millisecond)
		value := baseline[s] + 4*math.Sin(2*math.Pi*hours/24+float64(s)) + rng.NormFloat64()*0.8

		site := s/sensorsPerSite + 1
		line = line[:0]
		line = append(line, `{"ts":"`...)
		line = ts.AppendFormat(line, "2006-01-02T15:04:05.000Z07:00")
		line = fmt.Appendf(line, `","site":"plant-%02d","sensor":"plant-%02d-s%03d","metric":"temp_c","value":`, site, site, s%sensorsPerSite)
		line = strconv.AppendFloat(line, math.Round(value*100)/100, 'f', -1, 64)
		line = append(line, "}\n"...)
		if _, err := w.Write(line); err != nil {
			return err
		}
	}
	return nil
}

func storeError(e store.Error) error {
	switch e.Tag() {
	case store.Error_NotFound:
		return fmt.Errorf("store: not found")
	case store.Error_InvalidName:
		return fmt.Errorf("store: invalid name %q", e.InvalidName())
	case store.Error_Aborted:
		return fmt.Errorf("store: aborted: %s", e.Aborted())
	case store.Error_Storage:
		return fmt.Errorf("store: %s", e.Storage())
	default:
		return fmt.Errorf("store: unknown error (tag %d)", e.Tag())
	}
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil && v > 0 {
		return v
	}
	return fallback
}

func main() {}
