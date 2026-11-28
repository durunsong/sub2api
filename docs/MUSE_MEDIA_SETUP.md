# Muse Images and Video

## Scope

This adapter connects an administrator-managed Muse2API gateway through an existing OpenAI API-key account. It does not implement Muse browser login or a native Meta account platform.

The verified reference contract is `czg86389-hub/muse2api` at commit `59c413a85eecfb2e35da5d0895db20503dae73df`. Upstream Sub2API PR #7639 is text-only; #8722 adds Meta Code images but no video. Neither is merged wholesale into this Fork.

## Account and Group

1. Configure a working Muse2API instance separately. Set an externally reachable media base URL on that provider. Keep the gateway under administrator control.
2. Add an **OpenAI / API Key** account with its custom base URL, for example `https://muse.example/v1`, and its gateway key. Never use a panel JWT or Muse browser cookie as the API key.
3. Configure explicit model mappings: `muse-image -> muse-image` and `muse-video -> muse-video`. `muse-image-1.0` is also recognized as a compatible image ID; map it to the model actually supported by the gateway. Model aliases must be mapped on the account.
4. Bind the account to an OpenAI group and enable the group's media-generation permission. Composite groups can route the models to OpenAI accounts using the existing route configuration.
5. For images, enable image-count billing and explicitly configure **all three image prices (1K/2K/4K)**. The actual returned dimensions can differ from the requested dimensions; no unconfigured tier may fall back to a guessed price. Zero is an explicit free price.
6. For video, configure the `muse-video` model's 480p/720p/1080p prices, in **USD per second**, for the resolutions you offer. Missing pricing rejects generation before contacting the provider. Existing group pricing and rate multipliers still apply.

No new database migration or platform is required.

## Workbench

Open `/image-studio` and switch between the image and video tabs. Administrators configure visibility at `/admin/image-studio` on the corresponding tab. Account/group authorization remains authoritative; workbench settings do not create provider capabilities.

- Images: one output per request, text-to-image or reference-image edits; automatic, square, landscape and portrait sizes. Native masks, GPT quality controls, streaming and asynchronous image generation are rejected. Guided edits use the existing reference-image path.
- Video: text-to-video or ordered reference images; 4-15 seconds in the workbench; 480p/720p/1080p; 16:9 or 9:16. Muse reference images are not strict first/last frames. The direct API accepts 1-15 seconds.
- For a public video alias, select the Muse protocol in its workbench override and keep the account mapping pointed at `muse-video`.
- Refresh resumes the original task. It never automatically submits another paid generation.
- Do not share a public image alias across Muse and GPT accounts in the same group. Ambiguous Muse aliases are hidden because their parameter capabilities differ.

Provider size, quality and duration instructions are requests, not a guarantee of exact generated output. The adapter does not add provider capabilities.

## API and Settlement

- `POST /v1/images/generations`: synchronous OpenAI-compatible image response.
- `POST /v1/images/edits`: multipart `image`/`image[]`, or JSON `images: [{"image_url": "..."}]`. References remain in the moderation input.
- `POST /v1/videos`: JSON with `model`, `prompt`, `duration`, `resolution`, `size` (aspect ratio) and optional ordered `images: [{"url": "..."}]`.
- `GET /v1/videos/muse_task_...` and `/content`: bound to the original user, API key and account. Status responses expose only normalized task metadata; provider result URLs are replaced by the authenticated content endpoint.
- Videos settle after successful completion through the existing deferred, deduplicated billing path, using the explicit Muse price available at settlement. Queued and failed tasks are not successful outputs. If an administrator removes that price while a task is running, settlement remains pending instead of using Grok/default prices; restore the price and query the original task again. Do not manually resubmit an ambiguous failed create without checking the provider first.
- Media downloads are bounded, reject unsafe origins/paths and redirects, and do not send API credentials to arbitrary result URLs.
- Image billing uses dimensions verified from the returned image bytes, not the provider's echoed request size. If the output cannot be downloaded or verified, the request fails without automatic retry or user settlement; any provider generation cost is borne by the operator. A reachable provider media URL is therefore required even when a provider unexpectedly omits requested base64 data.

## Privacy and Validation

Sub2API workbench creative content stays in the user's browser and follows `NoContentPersistence` on the internal relay. **Muse2API itself stores tasks, prompts and generated media.** Configure retention and access controls on that external service; local workbench privacy does not imply upstream zero retention.

Automated checks use fake upstream responses and browser fixtures. No real Muse credentials, paid generations, production deployment or data migration are part of this change. A real deployment still needs administrator configuration and an authorized provider smoke test.
