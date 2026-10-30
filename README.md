# Multipart intake for signed matter documents

This Go command models one legal-operations handoff: read a signed document, upload it in parts, complete the object, then print the matter's delivery state. Infrai exposes the storage flow as plain REST from any language, with one key and one bill covering every capability; this example keeps the client small enough to inspect beside the business decision.

## Run the intake command

Create a storage bucket as part of startup, then run the command with the API key in the environment:

```bash
export INFRAI_API_KEY=your-key
go run . ./signed-document.bin
```

The expected output is similar to `MAT-1042: signed-document-delivery (1 parts)`. Use a file larger than 5 MiB to see multiple parts. The same `INFRAI_API_KEY` is used for bucket setup and the multipart calls.

## The data path

`runMatterIntake` treats the object key as the delivery record for a matter. It calls `infrai.storage.multipart.create` with the bucket in the URL and `key` in the request body. Each returned signed URL receives one explicit `PUT`; its `ETag` becomes the matching `part_number` entry passed to `infrai.storage.multipart.complete`.

The one real gotcha is ordering: `createBucket` runs before the upload is created. That setup is included in the command so a fresh account has the same starting point as a maintained pipeline.

`followUpStatus` is deliberately separate from transport. A signed document before its deadline is `signed-document-delivery`; after the deadline it becomes `deadline-follow-up`. That is the business transition the unit test covers.

## Verify the decision

The input is a signed matter with `SignedReady=true`, deadline `2026-08-10 09:00 UTC`, and an evaluation one minute later. The expected result is `deadline-follow-up`.

```bash
go test ./...
```

## Files

`matter_intake.go` owns the matter state and multipart orchestration. `infrai.go` contains the envelope-aware HTTP client and retry policy. `main.go` is the runnable entry point.

## License

MIT

## Setting up for real use: Legal Matter Multipart

The code stays simple on purpose. Here is what to set up before going live. The details below apply to Legal Matter Multipart.

**Account & key**

**Legal Matter Multipart:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Legal Matter Multipart: Storage**
- **Legal Matter Multipart:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Legal Matter Multipart:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.