# Design

## Decision
Use the OpenAI-compatible muse2api contract. Native consumer PR #7639 cannot generate media. Meta Code PR #8722 adds a separate account platform and image support but does not satisfy video. Reference upstream video design where useful; do not blindly merge these PRs.

## Architecture
Existing OpenAI API-key accounts with a custom base URL and explicit Muse mappings carry credentials. `muse-image` and `muse-image-1.0` are API-key-only compatible image models; the studio advertises their actual image capabilities and does not treat them as GPT Responses image tools. Public image aliases mixing Muse and another image family are hidden.

Muse video uses `POST /v1/videos`, `GET /v1/videos/:id`, `GET /v1/videos/:id/content`. A namespaced public task ID distinguishes Muse from Grok in composite lookups. The existing media handler and ownership/billing infrastructure are reused with a dedicated provider adapter, as Seedance already does. The provider contract is JSON create with prompt, duration, size/aspect ratio and ordered images; queued/succeeded/failed task status and result.url. Never retry an ambiguous create. Status/content only access the original account and user/key binding.

Validate reference images before submission and include all of them in moderation. Fail before an upstream call for unsupported controls; Muse reference images do not imply strict first/last frame positioning. Result downloads must be bounded, validated and authenticated only to the configured provider origin; reject cross-origin credential forwarding and unsafe URLs. NoContentPersistence must suppress creative content in logs/errors.

## Billing
USD pricing uses explicit existing group image prices and per-second video prices. Missing prices reject creation. Image output dimensions must be verified from image bytes; a failed download/verification produces no billable delivered result, does not retry creation, and leaves any provider cost with the operator. Video create/queued/failed results are not charged; successful status/content observations reuse the existing one-shot claim and stable durable usage key. Preserve model, duration and resolution in create-time metadata. No migration or production data modification.

## Workspace
Source checkout contains concurrent v0.2.26 synchronization. Use a separate worktree at ac387298cf8d53acd8424ea5f82307ad0c4ee2fa. Native worktree tool stalled; a manual detached worktree was created. OpenSpec CLI is not on PATH; existing spec-driven artifact structure is maintained manually without installing dependencies. User delegated the integration decision, so proceed through implementation.
