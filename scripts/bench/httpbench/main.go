// httpbench drives HTTP load against a Go component served by `wash dev`
// and emits results in the same JSONL row schema as wasmCloud/wasmCloud's
// bench-tools, so rows from this repo join the same history.json
// aggregate (pushed by .github/scripts/bench-push-results.mjs) and render
// on https://wasmcloud.github.io/arewefastyet/ alongside the Rust rows.
//
// Bench names carry a `_go` suffix (http_invoke_go, ...) because the
// aggregate is shared with wasmCloud/wasmCloud: the site's timelines key
// on the `bench` field, and the Rust repo already publishes http_invoke.
//
// Measurement model: a closed loop at concurrency 1 — one request at a
// time, latency recorded per request — matching the sequential-invoke
// semantics of the upstream criterion http_invoke bench (metric
// "mean_ns"). Percentiles land in the markdown summary only; the JSONL
// rows carry exactly the fields the site's readers already parse.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type scenario struct {
	param  string
	method string
	path   string
	body   string
}

type benchDef struct {
	// group mirrors criterion's group and names the example under bench.
	group     string
	example   string // folder under examples/components/
	scenarios []scenario
}

var benches = map[string]benchDef{
	"http_invoke_go": {
		group:   "http-server",
		example: "http-server",
		scenarios: []scenario{
			{param: "get_root", method: "GET", path: "/"},
			{param: "post_body", method: "POST", path: "/post", body: "ping"},
		},
	},
	"http_invoke_p3_go": {
		group:   "http-p3-streaming",
		example: "http-p3-streaming",
		scenarios: []scenario{
			{param: "post_echo", method: "POST", path: "/echo", body: "echo"},
		},
	},
}

// meta matches bench-tools' Meta struct field-for-field so flattened rows
// stay schema-compatible with history.json.
type meta struct {
	Sha         string `json:"sha"`
	ShortSha    string `json:"short_sha"`
	Ref         string `json:"ref"`
	RunID       string `json:"run_id"`
	RunAttempt  string `json:"run_attempt"`
	Timestamp   string `json:"timestamp"`
	Host        string `json:"host"`
	Kernel      string `json:"kernel"`
	CpusOnline  int    `json:"cpus_online"`
	IsolatedCPU string `json:"isolated_cpu"`
}

type row struct {
	Bench string `json:"bench"`
	Group string `json:"group"`
	Param string `json:"param"`
	meta
	Metric     string   `json:"metric"`
	Value      float64  `json:"value"`
	Throughput *float64 `json:"throughput"`
	MeanNs     float64  `json:"mean_ns"`
	MedianNs   float64  `json:"median_ns"`
	StdDevNs   float64  `json:"std_dev_ns"`
	CiLowNs    float64  `json:"ci_low_ns"`
	CiHighNs   float64  `json:"ci_high_ns"`
}

type stats struct {
	n                    int
	mean, median, stdDev float64
	min, max, p90, p99   float64
	requestsPerSec       float64
	errors               int
}

func main() {
	var (
		benchName = flag.String("bench", "", "bench to run (http_invoke_go, http_invoke_p3_go)")
		baseURL   = flag.String("url", "http://127.0.0.1:8000", "base URL wash dev serves on")
		outDir    = flag.String("out", "bench-output", "directory for results.jsonl, summary.md, metadata.json")
		warmup    = flag.Duration("warmup", 3*time.Second, "warmup duration per scenario")
		duration  = flag.Duration("duration", 15*time.Second, "measured duration per scenario")
		readyWait = flag.Duration("ready-wait", 120*time.Second, "how long to wait for the component to respond")
		printEx   = flag.Bool("print-example", false, "print the example folder for -bench and exit (used by run-bench.sh)")
	)
	flag.Parse()

	def, ok := benches[*benchName]
	if !ok {
		log.Fatalf("unknown bench %q; known: %s", *benchName, strings.Join(benchNames(), ", "))
	}
	if *printEx {
		fmt.Println(def.example)
		return
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	m := captureMeta()
	client := &http.Client{Timeout: 30 * time.Second}

	// wash dev serves HTTP before the workload is registered, so probe the
	// first scenario until it answers 2xx rather than waiting on the port.
	if err := waitReady(client, *baseURL, def.scenarios[0], *readyWait); err != nil {
		log.Fatalf("component never became ready: %v", err)
	}
	log.Printf("component ready at %s", *baseURL)

	var rows []row
	var summaries []string
	for _, sc := range def.scenarios {
		log.Printf("scenario %s: warmup %s, measure %s", sc.param, *warmup, *duration)
		runLoop(client, *baseURL, sc, *warmup) // warmup, discard
		lat := runLoop(client, *baseURL, sc, *duration)
		st, err := compute(lat, *duration)
		if err != nil {
			log.Fatalf("scenario %s: %v", sc.param, err)
		}
		rows = append(rows, toRow(*benchName, def.group, sc.param, m, st))
		summaries = append(summaries, summaryLine(sc, st))
		log.Printf("scenario %s: n=%d mean=%s p99=%s (%.0f req/s)",
			sc.param, st.n, fmtNs(st.mean), fmtNs(st.p99), st.requestsPerSec)
	}

	if err := writeResults(*outDir, *benchName, m, rows, summaries); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", filepath.Join(*outDir, "results.jsonl"))
}

func benchNames() []string {
	names := make([]string, 0, len(benches))
	for n := range benches {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func waitReady(client *http.Client, base string, sc scenario, limit time.Duration) error {
	deadline := time.Now().Add(limit)
	var lastErr error
	for time.Now().Before(deadline) {
		if _, err := doRequest(client, base, sc); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(2 * time.Second)
	}
	return lastErr
}

// runLoop issues requests back-to-back for d and returns per-request
// latencies in ns. Failed requests record a NaN sentinel and are counted
// as errors by compute.
func runLoop(client *http.Client, base string, sc scenario, d time.Duration) []float64 {
	var lat []float64
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		start := time.Now()
		_, err := doRequest(client, base, sc)
		elapsed := float64(time.Since(start).Nanoseconds())
		if err != nil {
			elapsed = math.NaN()
		}
		lat = append(lat, elapsed)
	}
	return lat
}

func doRequest(client *http.Client, base string, sc scenario) (int, error) {
	var body io.Reader
	if sc.body != "" {
		body = bytes.NewReader([]byte(sc.body))
	}
	req, err := http.NewRequest(sc.method, base+sc.path, body)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.StatusCode, fmt.Errorf("status %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func compute(lat []float64, wall time.Duration) (stats, error) {
	var ok []float64
	errs := 0
	for _, v := range lat {
		if math.IsNaN(v) {
			errs++
		} else {
			ok = append(ok, v)
		}
	}
	// A bench with failing requests is measuring the wrong thing; fail
	// loudly rather than publish rows from a degraded run.
	if errs > 0 {
		return stats{}, fmt.Errorf("%d/%d requests failed", errs, len(lat))
	}
	if len(ok) < 100 {
		return stats{}, fmt.Errorf("only %d samples; too few to report", len(ok))
	}
	sort.Float64s(ok)
	sum, sumSq := 0.0, 0.0
	for _, v := range ok {
		sum += v
	}
	mean := sum / float64(len(ok))
	for _, v := range ok {
		sumSq += (v - mean) * (v - mean)
	}
	sd := math.Sqrt(sumSq / float64(len(ok)-1))
	return stats{
		n:              len(ok),
		mean:           mean,
		median:         percentile(ok, 50),
		stdDev:         sd,
		min:            ok[0],
		max:            ok[len(ok)-1],
		p90:            percentile(ok, 90),
		p99:            percentile(ok, 99),
		requestsPerSec: float64(len(ok)) / wall.Seconds(),
	}, nil
}

func percentile(sorted []float64, p float64) float64 {
	idx := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func toRow(bench, group, param string, m meta, st stats) row {
	ciDelta := 1.96 * st.stdDev / math.Sqrt(float64(st.n))
	return row{
		Bench: bench, Group: group, Param: param, meta: m,
		Metric: "mean_ns", Value: st.mean,
		MeanNs: st.mean, MedianNs: st.median, StdDevNs: st.stdDev,
		CiLowNs: st.mean - ciDelta, CiHighNs: st.mean + ciDelta,
	}
}

func captureMeta() meta {
	ref := os.Getenv("WASMCLOUD_BENCH_REF")
	if ref == "" {
		ref = gitOut("describe", "--tags", "--exact-match", "HEAD")
	}
	if ref == "" {
		ref = gitOut("rev-parse", "--abbrev-ref", "HEAD")
	}
	host, _ := os.Hostname()
	return meta{
		Sha:         gitOut("rev-parse", "HEAD"),
		ShortSha:    gitOut("rev-parse", "--short=12", "HEAD"),
		Ref:         ref,
		RunID:       envOr("GITHUB_RUN_ID", "local"),
		RunAttempt:  envOr("GITHUB_RUN_ATTEMPT", "1"),
		Timestamp:   time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Host:        host,
		Kernel:      cmdOut("uname", "-r"),
		CpusOnline:  runtime.NumCPU(),
		IsolatedCPU: "?",
	}
}

func writeResults(outDir, bench string, m meta, rows []row, summaries []string) error {
	var jsonl bytes.Buffer
	for _, r := range rows {
		b, err := json.Marshal(r)
		if err != nil {
			return err
		}
		jsonl.Write(b)
		jsonl.WriteByte('\n')
	}
	if err := os.WriteFile(filepath.Join(outDir, "results.jsonl"), jsonl.Bytes(), 0o644); err != nil {
		return err
	}

	var md strings.Builder
	fmt.Fprintf(&md, "### %s @ %s (`%s`)\n\n", bench, m.Ref, m.ShortSha)
	md.WriteString("| scenario | n | mean | median | p90 | p99 | std dev | req/s |\n")
	md.WriteString("|---|---|---|---|---|---|---|---|\n")
	for _, line := range summaries {
		md.WriteString(line)
	}
	fmt.Fprintf(&md, "\nhost: `%s` · kernel `%s` · %d cpus · go `%s` · wash `%s`\n",
		m.Host, m.Kernel, m.CpusOnline, strings.TrimPrefix(cmdOut("go", "version"), "go version "), washVersion())
	if err := os.WriteFile(filepath.Join(outDir, "summary.md"), []byte(md.String()), 0o644); err != nil {
		return err
	}

	metadata := map[string]any{
		"bench":        bench,
		"meta":         m,
		"go_version":   cmdOut("go", "version"),
		"wash_version": washVersion(),
		"event":        os.Getenv("GITHUB_EVENT_NAME"),
		"workflow":     os.Getenv("GITHUB_WORKFLOW"),
		"actor":        os.Getenv("GITHUB_ACTOR"),
		"repository":   os.Getenv("GITHUB_REPOSITORY"),
	}
	b, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "metadata.json"), b, 0o644)
}

func summaryLine(sc scenario, st stats) string {
	return fmt.Sprintf("| %s `%s` | %d | %s | %s | %s | %s | %s | %.0f |\n",
		sc.method, sc.path, st.n,
		fmtNs(st.mean), fmtNs(st.median), fmtNs(st.p90), fmtNs(st.p99), fmtNs(st.stdDev),
		st.requestsPerSec)
}

func fmtNs(ns float64) string {
	switch {
	case ns >= 1e9:
		return fmt.Sprintf("%.2fs", ns/1e9)
	case ns >= 1e6:
		return fmt.Sprintf("%.2fms", ns/1e6)
	case ns >= 1e3:
		return fmt.Sprintf("%.1fµs", ns/1e3)
	default:
		return fmt.Sprintf("%.0fns", ns)
	}
}

func gitOut(args ...string) string {
	return cmdOut("git", args...)
}

func cmdOut(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func washVersion() string {
	return cmdOut("wash", "--version")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
