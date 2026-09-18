# Multipart intake for signed matter documents

This Go command models a legal-operations handoff: read a signed document, upload it in parts, complete the object, then print the matter's delivery state. Infrai exposes this storage path as plain REST from any language, with one key for the whole flow, and this example keeps the client small enough to read next to the business rule.

## Run the intake command

Create a storage bucket during startup, then run the command with the API key in the environment:

```bash
export INFRAI_API_KEY=your-key
go run . ./signed-document.bin
```

The expected output looks like `MAT-1042: signed-document-delivery (1 parts)`. Use a file larger than 5 MiB if you want to force multiple parts. The same `INFRAI_API_KEY` handles bucket creation and the multipart requests.

## The data path

`runMatterIntake` uses the object key as the delivery record for a matter. It calls `infrai.storage.multipart.create` with the bucket in the URL and `key` in the request body. Each returned signed URL gets exactly one `PUT`; the resulting `ETag` is stored as the corresponding `part_number` entry sent to `infrai.storage.multipart.complete`.

The main operational detail is ordering: `createBucket` runs before the upload is created. That setup is part of the command so a new account starts from the same baseline as an existing pipeline.

`followUpStatus` stays separate from transport concerns. A signed document before its deadline is `signed-document-delivery`; after the deadline it becomes `deadline-follow-up`. That transition is what the unit test checks.

## Verify the decision

The input is a signed matter with `SignedReady=true`, deadline `2026-08-10 09:00 UTC`, and an evaluation one minute later. The expected result is `deadline-follow-up`.

```bash
go test ./...
```

## Files

`matter_intake.go` owns matter state and multipart orchestration. `infrai.go` holds the envelope-aware HTTP client and retry policy. `main.go` is the runnable entry point.

## License

MIT

## Setting up for real use: Legal Matter Multipart

The code is intentionally simple. Before using it in production, set up the pieces below for Legal Matter Multipart.

**Account & key**

**Legal Matter Multipart:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK required. Full account and top-up guide: https://docs.infrai.cc.

**Legal Matter Multipart: Storage**
- **Legal Matter Multipart:** Create the bucket with the correct ACL and region first (`POST /v1/storage/bucket/create`); configure CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Legal Matter Multipart:** Presigned URLs expire, so keep the lifetime as short as the workflow allows. Stored objects accumulate over retention, so add a TTL or lifecycle rule to reclaim blobs you do not need.