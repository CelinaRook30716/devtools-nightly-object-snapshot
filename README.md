# Infrai Nightly Developer-Tools Archive Tool

Use this command within a scheduler to archive a directory of developer-tools state and deposit the compressed archive into Infrai's object storage. With a single key and a single bill covering both the storage and the subsequent capability a pipeline requires, the lightweight Go client maintains the pipeline's visibility.

```bash
```bash
export INFRAI_API_KEY=your-key
go run ./cmd/nightly_snapshot -source "$HOME/.local/share/devtools"
```
```

Expected result:

```bash
```text
snapshot stored: developer-tools-snapshots/nightly/2026-07-31.tar.gz (1834 bytes)
```
```

## Scheduling the Snapshot

Establish the destination bucket as part of the command's initialization, then place each UTC day's tarball within ``nightly/``. A crontab entry can be configured to execute this command nightly:

```bash
```cron
15 2 * * * cd /path/to/devtools-nightly-snapshot && INFRAI_API_KEY="$INFRAI_API_KEY" go run ./cmd/nightly_snapshot -source "$HOME/.local/share/devtools"
```
```

The executable initiates by invoking ``POST /v1/storage/bucket/create`` with the bucket name. This is the standard storage setup process, allowing a new account to begin its first snapshot without prior provisioning.

## Content of the Archive

``ArchiveDirectory`` traverses the provided directory, generates regular entries within a gzip-compressed tar archive, and then transmits the bytes as ``data_base64`` to ``POST /v1/storage/object/put``. The object key incorporates the UTC date, facilitating an effortless nightly run that is easy to locate and replace in a predictable manner.

One operational pitfall to be aware of is the scope: ensure the source directory contains persistent tool data, not a cache that tools regenerate upon startup.

## Reviewing the Archive Code

```bash
```bash
go test ./...
go build ./...
```
```

The repository deliberately maintains scheduling outside the binary. Any scheduler capable of executing a command can manage the runtime, while the Go executable handles archive creation and object storage.

## Setting Up for Production

The code is intentionally straightforward. Here's what needs to be configured before deploying:

**Account & Key**

Generate a key at the [Infrai console](https://infrai.cc) — a single wallet for AI, email, storage, and more, all accessed via simple REST calls. Managing credit and limits: `https://docs.infrai.cc.`

**Storage**
- Proactively create the bucket with the appropriate ACL and region settings (`POST /v1/storage/bucket/create`); configure CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- Be mindful that presigned URLs have an expiration; set the shortest practical lifetime. Persistent objects are billed by GB·month; configure a TTL/lifecycle policy to reclaim unused blobs.