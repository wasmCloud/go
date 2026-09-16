# Building

Every version you need to build a component against a released
`go.wasmcloud.dev/component`, and how to build with no network at all.

The versions below are verified against
[`component/v0.1.5`](https://github.com/wasmCloud/go/releases/tag/component%2Fv0.1.5).
They live in [`airgap/versions.env`](./airgap/versions.env), which is the file
the tooling reads; this document is the prose. `make airgap-versions` re-derives
every value from its real upstream source and fails if the two disagree.

## The short version

| What | Version |
|---|---|
| Go | 1.27.1 (`go.bytecodealliance.org/pkg` declares `go 1.27.1`) |
| `go.wasmcloud.dev/component` | v0.1.5 |
| `go.bytecodealliance.org/pkg` | v0.2.4-0.20260911130647-2495ff7eca86 |
| componentize-go | v0.4.3 (module and toolchain) |
| wit-bindgen-go | 0.61.1 (rev `4f9a02d7`) |
| wit-component / wit-parser | 0.258.0 |
| Patched Go (**async worlds only**) | `dicej/go` `go1.27.1-wasi-on-idle` |
| `wash` | v2.9.0 |

Nothing else is required. If a build reaches for something not on this list,
that is a bug — please open an issue.

## Always pin a componentize-go *tag*

`github.com/bytecodealliance/componentize-go`, the module you get via the `tool`
directive, is a ~130-line downloader. Its `main.go` hardcodes a release string:

```go
release := "v0.4.3"
```

and fetches *that* Rust binary from GitHub Releases into a user cache. The real
`wit-bindgen` and `wit-component` versions live in that release's `Cargo.toml`,
not anywhere in the Go module graph.

**Tagged releases name themselves**, so the module and the toolchain agree:
v0.4.1 fetches v0.4.1, v0.4.2 fetches v0.4.2, v0.4.3 fetches v0.4.3. Pin a tag
and there is nothing to think about. That is what this repo does.

**A pseudo-version breaks that, silently.** A pseudo-version names an untagged
commit on `main`, and `release :=` is only bumped by hand in the "prep for
release" commit just before tagging — so a snapshot of `main` still names the
*previous* release. Depend on `v0.4.4-0.2026...` and you get the v0.4.3
toolchain, while the module's own `Cargo.toml` advertises versions that never
run. This repo was in exactly that state before: it pinned
`v0.4.2-0.20260827144128-20f3b0c2a412`, a pre-release of v0.4.2, and therefore
built everything with v0.4.1 and wit-bindgen 0.59.0.

So: **read toolchain versions from the release named by `release :=`, not from
the module you depend on**, and do not bump this dependency to a `main`
pseudo-version expecting to get `main`'s wit-bindgen. You will not.

To confirm what a given module version will fetch:

```shell
grep 'release :=' "$(go env GOMODCACHE)/github.com/bytecodealliance/componentize-go@<version>/main.go"
```

`make airgap-versions` checks this on every CI run: it reads `release :=` out of
the pinned module and fails if it does not match the module's own version.

## The patched Go

**Upstream Go cannot build the async examples. At all.** There is no flag and no
slow path. Async WASI P3 support depends on `runtime.wasiOnIdle`
([golang/go#76775](https://github.com/golang/go/pull/76775)), which is still
open, so it exists only in a patched fork.

Eight of the thirteen examples need it — `http-p3-streaming`, both
`http-local-routing` modules, plus all five `nats-*` modules. The nats ones are not opting in: every function on
`wasmcloud:nats@0.1.0` is async, so importing it forces the P3 world.

componentize-go handles this for you. It resolves the target world, and if the
world needs async it probes your Go by reading
`<goroot>/src/runtime/lock_wasip1.go` for `wasiOnIdle`. If the patch is missing,
it downloads a patched toolchain and uses that instead, leaving your own Go
alone. The sync examples build with stock Go and never trigger any of this.

Two patched builds exist:

| Patched build | Pairs with | How you get it |
|---|---|---|
| `go1.27.1-wasi-on-idle` | Go 1.27 | **Automatic** — componentize-go v0.4.3 fetches it |
| `go1.25.5-wasi-on-idle-v2` | Go 1.25 | Fetched by componentize-go v0.4.2 and earlier; no longer used here |

Go 1.25 is no longer an option for this repo: `go.bytecodealliance.org/pkg`
now declares `go 1.27.1`, so every module that imports it needs a 1.27.1
toolchain regardless of which world it builds.

Two things make async builds fail with `downloaded Go does not support async`.
Both come down to Go's automatic toolchain switching, and both are fixed on your
side:

**Install Go 1.27.1 or newer; do not rely on `GOTOOLCHAIN=auto`.** If your
`go` is older, `go tool componentize-go` switches to a downloaded 1.27.1 and
exports `GOROOT` pointing at it. componentize-go v0.4.3 probes for the patch by
running `go env GOROOT`, and the inherited variable makes the patched toolchain
report the *stock* root, so the probe fails. The `--go` flag goes through the
same probe and does not help. Sync builds are unaffected.

**Upgrading from componentize-go v0.4.2 or earlier: delete the cached patched
Go.** The cache is version-blind (see below), so a machine that built an async
world before keeps its `go1.25.5` tree and never fetches `go1.27.1`. That old
toolchain sees the modules' `go 1.27.1` directive, re-execs into stock 1.27.1,
and fails the same probe. Remove `componentize-go/v2/go-<os>-<arch>-bootstrap`
from your user cache directory (`~/Library/Caches` on macOS, `$XDG_CACHE_HOME`
or `~/.cache` on Linux) and the next async build downloads the right one.

### Using a patched Go you downloaded yourself

Tarballs are published at:

```
https://github.com/dicej/go/releases/download/<tag>/go-<os>-<arch>-bootstrap.tbz
```

Each unpacks to a complete Go tree (`bin/go`, `src/`, `VERSION`). Two ways to
point componentize-go at one:

**Preferred — the `--go` flag.** componentize-go accepts your path directly and
skips the download entirely, as long as the tree really carries the patch:

```shell
go tool componentize-go build --go /opt/go-linux-amd64-bootstrap/bin/go
```

The path must end in `bin/go`: the patch probe walks two directories up to find
`src/runtime/lock_wasip1.go`. Under `wash build`, put the flag in the example's
`.wash/config.yaml` under `build.command`.

**Alternative — seed the cache.** Unpack the tree to
`$XDG_CACHE_HOME/componentize-go/v2/go-<os>-<arch>-bootstrap` and it will be
picked up. This works, and is what the air-gapped image does, but it is
version-blind: the path does not encode the Go version and the lookup only
checks whether the binary exists, so a stale toolchain is reused forever.

> **Do not trust `go version` on a patched toolchain.** `GOTOOLCHAIN=auto` makes
> it re-exec into your stock Go and report that version instead. Check the
> `VERSION` file, or the binary's mtime.

There is no environment variable that overrides the *download URL*. The internal
function takes one, but it is only ever called with defaults, so `--go` is the
supported escape hatch.

## Building offline

A build reaches the network in six places. All six have a fix:

| Egress | Trigger | Fix |
|---|---|---|
| `proxy.golang.org` | Resolving module deps | Pre-populated `GOMODCACHE`, `GOFLAGS=-mod=mod`, `GOPROXY=off` |
| `sum.golang.org` | Writing a `go.sum` entry | `GOSUMDB=off` |
| Go toolchain download | `GOTOOLCHAIN=auto` with a newer `go` directive | `GOTOOLCHAIN=local` and a matching toolchain installed |
| componentize-go releases | The wrapper fetching its Rust binary | Seed `$XDG_CACHE_HOME/componentize-go/bin` and `version.txt` |
| `dicej/go` releases | Async worlds needing the patched Go | Seed `$XDG_CACHE_HOME/componentize-go/v2/go-<os>-<arch>-bootstrap` |
| OCI registry | `wash build` resolving WIT deps | `wash build --skip-fetch`; WIT is vendored in-tree |

`GOPROXY=off` does not imply the second one. The checksum database is a separate
call, and any module without a `go.sum` entry triggers it — which includes
anything generated from `templates/`, since templates ship without a `go.sum`
and are tidied after `wash new`. Populating the module cache is the point at
which checksums are genuinely verified; `GOSUMDB=off` afterwards skips a check
that has already happened rather than skipping it altogether.

The two componentize-go caches share one root, so seeding
`$XDG_CACHE_HOME/componentize-go` covers both. Set `XDG_CACHE_HOME` explicitly
rather than relying on `$HOME`, so the path is predictable.

`version.txt` must contain the release string the wrapper expects (`v0.4.3`
here) or it will re-download on the first build, air gap or not.

## Verifying it

The repo ships a container that has exactly these versions and no network, and
builds every example in it:

```shell
make airgap-test     # build every example with --network none
make airgap-shell    # same image, interactive, for debugging
make airgap-versions # re-derive every version and check this document
```

This runs in CI on every change to the SDK, the examples, or the templates. A
build that needs something not pinned here fails there.

By default the container builds against the **released** SDK, exactly as the
checked-in `go.mod` files declare — the situation a user cloning this repo is
in. To test the working tree instead:

```shell
make airgap-test AIRGAP_USE_LOCAL_SDK=1
```

## Keeping this current

Cutting an SDK release means bumping the examples to it. All eleven example
modules and `templates/http-hello-world` must require the same version, and that
version must be a real tag — not a pseudo-version, and not an older release that
happens to resolve. Under minimal version selection they resolve to exactly what
they declare, so a stale pin means the examples are built against SDK code
nobody is shipping.

Then update `airgap/versions.env` and run `make airgap-versions`.

Bumping componentize-go also moves wit-bindgen, because the wit-bindgen version
is baked into the componentize-go release rather than chosen separately. Two
examples (`http-keyvalue-crud` and `http-otel`) declare their own extra world
and commit the resulting bindings, so those need refreshing to match:

```shell
./examples/components/regenerate_bindings.sh
```

`make airgap-versions` fails if the committed `// Generated by wit-bindgen X`
headers fall behind the pinned toolchain, naming the files to regenerate.

### Known follow-up

The `dicej/go` tarballs have no checksums recorded here. They should be pinned
by SHA256 and verified during the image build, so a re-cut release surfaces as a
build failure rather than a mystery.
