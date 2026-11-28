# Integrated verification

Date: 2026-11-29. Final implementation is in the original `personal_code/sub2api` checkout on top of `97b187b90` (Fork sync to v0.2.26). Changes were transferred with patches after the original checkout became clean. The concurrent upstream image-safety hook, release version and Fork customizations remain intact. The user subsequently authorized committing and pushing after review and verification.

## Backend

From `backend/`, all commands exited 0:

```sh
GOMAXPROCS=2 GOGC=50 go test -p 1 -tags=unit ./internal/service ./internal/handler ./internal/server/routes -run 'Muse|CompatibleImages|ImageStudio|VideoStudio|GrokMedia|GrokVideo|Seedance|VideoBilling|ImageCountBilling|OpenAIImages|OpenAIRecordUsage' -count=1
GOMAXPROCS=2 GOGC=50 go build -p 1 ./...
GOMAXPROCS=2 GOGC=50 go vet -p 1 ./...
```

All three selected test packages passed after the final code edit. Coverage includes Muse API-key mapping and pricing gates, task ownership, repeated completion claims, no paid create replay, URL restrictions, public content without credentials, explicit current video pricing and recovery after restoration, image dimensions from bytes, mixed/duplicate multipart rejection, raw reference rejection, binary-label image moderation, alias ambiguity, studio payloads and existing media behavior. This is a focused suite, not every backend unit or database integration test.

Expected failing tests were observed before fixes for multipart hidden references, echoed image dimensions, GPT-shaped aliases, mixed Muse/GPT aliases, raw base64 references, binary-label upload moderation and the `/v1/videos` routing condition. Later integrated tests passed.

## Frontend

From `frontend/`, all commands exited 0:

```sh
pnpm test:run src/features/video-studio src/views/admin/__tests__/groupsVideoModelPricing.spec.ts src/views/admin/__tests__/groupsImagePricing.spec.ts
pnpm typecheck
pnpm exec eslint e2e/video-studio.cjs src/features/video-studio/components/VideoComposer.vue src/features/video-studio/components/VideoSettings.vue src/features/video-studio/components/VideoStudioSettings.vue src/features/video-studio/parameters.spec.ts src/features/video-studio/parameters.ts src/features/video-studio/types.ts src/features/video-studio/useVideoStudio.ts src/i18n/locales/en/videoStudio.ts src/i18n/locales/zh/videoStudio.ts src/views/admin/__tests__/groupsImagePricing.spec.ts src/views/admin/__tests__/groupsVideoModelPricing.spec.ts src/views/admin/groupsImagePricing.ts src/views/admin/groupsVideoModelPricing.ts
pnpm build
```

36 tests in 10 files passed. The build's 3 locale checks passed; Vite compiled 1199 modules. Existing non-blocking warnings remain for old Browserslist data, mixed static/dynamic imports and chunks over 500 kB.

The existing Playwright fixture `node e2e/video-studio.cjs` passed against the integrated dev server at `http://127.0.0.1:5187`, once with `STUDIO_PROTOCOL=muse` and once with the default Grok protocol. The bundled Playwright runtime was supplied through `NODE_PATH`; no dependency was installed. Both runs verified one create, resumed polling after refresh, reference upload, decoded video metadata, IndexedDB blob, assets/history, independent admin settings, desktop/mobile overflow and light/dark views. Screenshots were inspected under `/tmp/sub2api-muse-video-ui-integrated` and `/tmp/sub2api-grok-video-ui-integrated`.

## Review and limits

The additional pre-push review identified four issues and added regression coverage:

- P1: GPT-shaped account aliases with binary-label multipart PNG uploads omitted the image from moderation. MIME detection now occurs in the shared multipart parser before moderation, independently of the eventual account mapping.
- P2: Composite image catalogs skipped all real-platform accounts. Capability resolution now follows the existing images route decision before evaluating the target platform's account mappings.
- P2: Recognizable video aliases ignored an administrator's explicit Muse protocol. A configured nonempty protocol now takes precedence over name inference.
- P2: Composite image aliases rewritten to another alias failed validation before account mapping. The trusted studio canonical model now also applies when both the route's original public name and rewritten model match the server-selected request. JSON generation and multipart editing tests exercise parsing and forwarding, and reject mismatched public-name context.

Each issue was reproduced by a failing regression test before the corresponding production fix. No real provider was contacted.

Independent image/studio and video reviews found issues that were fixed and re-reviewed without remaining confirmed P1/P2 findings. The multipart MIME check was additionally fixed and covered by the final integrated regression suite. Initial review history is retained in `images-studio-review.md`; the video implementation report is `video-report.md`.

`git diff --check` passed. Changed/new text files passed strict UTF-8 decoding and checks for BOM and replacement characters. No dependency manifests, lockfiles, release version, migrations, credentials or production settings were changed.

No real Muse provider request or paid generation was performed. Deployment still requires a reachable administrator-managed Muse2API service, explicit mappings and prices, and an authorized live smoke test. Unverifiable image outputs are not charged to the user or automatically retried; the operator bears any provider cost. Muse2API retains its own prompts/tasks/media independently of the local studio privacy setting.
