# Security Policy

This policy covers the wasmCloud/go repository.
The full process, including how reports are triaged, fixed and disclosed,
is the [wasmCloud Security Process and Policy](https://github.com/wasmCloud/wasmCloud/blob/main/SECURITY.md).
How this repository is threatened and defended is in [SECURITY_ASSESSMENT.md](SECURITY_ASSESSMENT.md).

## Report A Vulnerability

CRA stewardship: This project is supported under the Linux Foundation CRA stewardship framework, as described at https://www.linuxfoundation.org/security. Security vulnerabilities should be reported through the mechanisms described below, which we will coordinate with our CRA steward. For actively exploited vulnerabilities and severe incidents that may require CRA escalation, please use the project's emergency security reporting mechanisms as appropriate.

Don't report security vulnerabilities as public issues, discussions or pull requests. Report them
privately in either of these ways:

- Email the private security list at **security@wasmcloud.com** with the details.
- Use [GitHub private vulnerability reporting](https://github.com/wasmCloud/go/security/advisories/new)
  from this repository's **Security** tab.

Reports go to the wasmCloud security team, a subset of the org maintainers. Each report is
acknowledged within **3 business days**.

## Published vulnerabilities

- Fixed vulnerabilities are published as
  [GitHub Security Advisories](https://github.com/wasmCloud/go/security/advisories) in this repository,
  and reach the [Go vulnerability database](https://pkg.go.dev/vuln/) from there.
- Known vulnerabilities that don't affect the SDK are recorded,
  with the reason, in the OpenVEX documents in [vex/](vex/).

## Verifying a release

### The SDK module

`go` verifies every module it downloads against the
[Go checksum database](https://sum.golang.org),
so `go get go.wasmcloud.dev/component@vX.Y.Z` fails if the module isn't the one first published at that version.
Don't turn that check off for `go.wasmcloud.dev`:
leave it out of `GONOSUMDB`, `GOPRIVATE` and `GOINSECURE`, and don't set `GOSUMDB=off`.
To re-check modules already in your module cache, run `go mod verify`.

### The release and its author

Each `component/vX.Y.Z` tag is pushed by
[release-tag.yaml](.github/workflows/release-tag.yaml)
as the `wasmcloud-automation-app[bot]` GitHub App, never by a person.
Releases are immutable, so GitHub signs a release attestation when each release is published,
binding the tag to its commit.
Verify it, and that it was published in this repository, with:

```shell
gh release verify component/vX.Y.Z --repo wasmCloud/go
```

Releases published before 2026-11-05 predate immutable releases and have no attestation;
for those, rely on the checksum database.

### The go-airgap image

Each `ghcr.io/wasmcloud/go-airgap` image carries a signed SLSA build provenance attestation.
Check both the image and the workflow that built it:

```shell
gh attestation verify oci://ghcr.io/wasmcloud/go-airgap:<tag> \
  --repo wasmCloud/go \
  --signer-workflow wasmCloud/go/.github/workflows/airgapped.yaml
```

## Secrets and credentials

This repository follows the
[stored secrets](https://github.com/wasmCloud/wasmCloud/blob/main/SECURITY.md#stored-secrets)
and rotation policy of the wasmCloud project.
In this repository that means:

- **What is stored:** the wasmCloud bot GitHub App's ID and private key,
  as the repository secrets `BOT_APP_ID` and `BOT_APP_PRIVATE_KEY`.
  [release-train.yaml](.github/workflows/release-train.yaml),
  [release-tag.yaml](.github/workflows/release-tag.yaml)
  and [renovate.yaml](.github/workflows/renovate.yaml) use them to mint short-lived app tokens.
  Every other workflow uses only its job's `GITHUB_TOKEN` or OIDC.
- **Storage:** secrets live only in GitHub Actions secrets,
  never in the repository, images, artifacts or logs.
  Secret scanning and push protection block a push that contains a detected secret.
- **Access:** only repository admins (`@wasmCloud/go-maintainers`) and org maintainers
  can create, update or delete them.
  Anyone with write access can run a workflow that reads them,
  so write access is granted only as described in [MAINTAINERS.md](MAINTAINERS.md).
  Pull requests, including those from forks, never run a workflow that reads them.
- **Rotation:** a secret is rotated immediately if it may have been exposed,
  when someone who held it or administers it loses that access,
  and at least once a year.
  To rotate the app key, generate a new private key in the app's settings,
  update `BOT_APP_PRIVATE_KEY`, run the Renovate workflow to confirm it works,
  then delete the old key.
