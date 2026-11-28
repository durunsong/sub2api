# Muse video adapter verification

## Behavior

- OpenAI and Composite groups accept Muse creation through both `POST /v1/videos` and `POST /videos`; Grok creation remains on `/videos/generations`. Composite creation must resolve to the OpenAI platform.
- A configured OpenAI API-key account with an explicit public-model mapping to `muse-video` forwards create/status/content to the configured provider. Public IDs use `muse_task_`; status and content reuse user, key and account ownership binding.
- Unsupported request fields, reference formats and video billing controls fail before a paid create. Every accepted image reference enters the moderation payload. Failed or ambiguous upstream creates are not replayed.
- A completed task can bill once using its create-time duration and resolution. At settlement, the current group price for `muse-video` and the result resolution is loaded from the group repository. A missing price fails the record task and releases the claim, allowing a later status poll after the price is restored. Muse never falls back to flat, channel or default video pricing.
- Status exposes only public task fields. Content accepts only a same-origin `/v1/media/{name}` result URL, disables redirects, omits credentials on the media download, and bounds the response to 128 MiB.

## Verification

- `go test -tags=unit ./internal/service ./internal/handler ./internal/server/routes -run 'TestMuse|TestGatewayRoutes(OpenAIVideosReachMuseHandler|NonGrokVideosAreRejectedAtPlatformGate)' -count=1`: passed all three packages on 2026-11-29.
- The `/v1/videos` route regression test failed with the previous route condition (`404` at `/v1/videos`) and passed after the fix.
- No provider request, production operation, migration, dependency install, commit or push was performed in this task.

## Integrated validation

- The parent task transferred the changes to the original v0.2.26 checkout and passed the broader focused tests, full backend build and vet. See `verification.md` for the final commands and frontend/browser evidence. The command above records the earlier video worktree check.
