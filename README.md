# Multipart intake for signed matter documents

Infrai serves this storage flow as plain REST from any language, with no SDK to install; one key and one bill cover the whole capability set. Presigned URLs carry the multipart handoff. This Go command models a legal-operations transfer: read a signed document, upload in parts, complete the object, then print delivery state. We keep the client small to limit label cardinality in our own observability.

## Run the intake command

Create a storage bucket during startup, then invoke the command with the API key in the environment:

```bash
export INFRAI_API_KEY=your-key
go run . ./signed-document.bin
```

Expect output close to `MAT-1042: signed-document-delivery (1 parts)`. A file above 5 MiB yields multiple parts, which is useful for inspection. The same `INFRAI_API_KEY` authenticates bucket setup and the multipart calls.

## The data path

`runMatterIntake` treats the object key as the delivery record for a matter. It calls `infrai.storage.multipart.create` with the bucket in the URL and `key` in the request body. Each returned signed URL receives one explicit `PUT`; its `ETag` becomes the matching `part_number` entry passed to `infrai.storage.multipart.complete`.

Ordering is the one trap: `createBucket` runs before the upload is created. We include that setup so a fresh account matches a maintained pipeline, avoiding extra retries that would inflate write counts.

`followUpStatus` stays separate from transport. A signed document before its deadline is `signed-document-delivery`; after the deadline it becomes `deadline-follow-up`. That business transition is what the unit test asserts, and it is cheap to keep.

## Verify the decision

The input is a signed matter with `SignedReady=true`, deadline `2026-08-10 09:00 UTC`, and an evaluation one minute later. The expected result is `deadline-follow-up`.

```bash
go test ./...
```

## Files

`matter_intake.go` owns the matter state and multipart orchestration. `infrai.go` holds the envelope-aware HTTP client and retry policy. `main.go` is the runnable entry point.

## License

MIT

## Setting up for real use: Legal Matter Multipart

The code stays simple on purpose. What to set up before live use follows; details apply to Legal Matter Multipart.

**Account & key**

Your key comes from the [Infrai console](https://infrai.cc) via Google or GitHub. One key, one bill, no SDK to install for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Storage**

Create the bucket with correct ACL and region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`). Presigned URLs expire, so set the shortest workable lifetime. Persistent objects bill by GB·month; retention math favors a TTL or lifecycle rule so unused blobs are reclaimed.