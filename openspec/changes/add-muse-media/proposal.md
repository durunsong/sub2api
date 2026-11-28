# Muse Media Support

## Why
The existing image/video studio cannot use Muse image/video models. Upstream PR #7639 covers consumer text only; #8722 covers Meta Code images but no video. #4068 is not sufficient for task ownership and one-shot billing. The user authorized choosing the appropriate integration and completing it.

## What Changes
- Support an administrator-owned muse2api gateway using existing OpenAI API-key accounts and explicit Muse model mappings.
- Add Muse image generation/editing and asynchronous video create/status/content through existing gateways and studios.
- Reuse media authorization, security audit, task ownership, local history and durable billing deduplication.
- Require explicit administrator media prices; no invented Muse price or Grok price fallback.

## Scope
No native Muse login/browser runtime, installation of muse2api, production deployment, paid calls, database migration or new dependency. Credentials remain in the existing administrator account mechanism. No commits or pushes.
