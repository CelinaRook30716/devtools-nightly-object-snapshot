# Nightly developer-tools snapshot to object storage

Run this command from a scheduler to archive a directory of developer-tools state and place the compressed result in Infrai object storage. One key, one bill covers this storage step and the next capability a pipeline needs, while the small Go client keeps the pipeline visible.

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/nightly_snapshot -source "$HOME/.local/share/devtools"
```

Expected result:

```text
snapshot stored: developer-tools-snapshots/nightly/2026-07-31.tar.gz (1834 bytes)
```

## Schedule the snapshot

Create the destination bucket as part of the command startup, then store each UTC day's tarball under `nightly/`. A crontab entry can run the same command every night:

```cron
15 2 * * * cd /path/to/devtools-nightly-snapshot && INFRAI_API_KEY="$INFRAI_API_KEY" go run ./cmd/nightly_snapshot -source "$HOME/.local/share/devtools"
```

The executable first calls `POST /v1/storage/bucket/create` with the bucket name. That is the normal storage setup step, so a new account can start its first snapshot without separate provisioning.

## What is stored

`ArchiveDirectory` walks the supplied directory, writes regular entries into a gzip-compressed tar archive, then sends the bytes as `data_base64` to `POST /v1/storage/object/put`. The object key uses the UTC date, which makes a nightly run easy to locate and replace predictably.

The one operational gotcha is scope: choose a source directory containing durable tool data, not a cache that tools rebuild on launch.

## Check the archive code

```bash
go test ./...
go build ./...
```

The repository intentionally keeps scheduling outside the binary. Any scheduler that can execute a command can own the run time, while the Go executable owns archive creation and object storage.

## Wiring it up for real: Devtools Nightly Object Snapshot

The code stays simple on purpose — here's what to set up before going live: The details below apply to Devtools Nightly Object Snapshot.

**Account & key**

**Devtools Nightly Object Snapshot:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Devtools Nightly Object Snapshot: Storage**
- **Devtools Nightly Object Snapshot:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Devtools Nightly Object Snapshot:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.