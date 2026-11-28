## ADDED Requirements

### Requirement: Muse image support
The gateway SHALL accept configured Muse image models only through custom OpenAI API-key accounts, preserving image authorization, moderation, billing and local-only studio history.

#### Scenario: Generate or edit
- WHEN an authorized user chooses a configured Muse image model with an explicit price
- THEN the image studio sends the supported generation/edit request and displays the returned image
- AND no GPT Responses image-tool conversion is used.

### Requirement: Muse video lifecycle
The system SHALL create, query and download Muse videos through the existing video studio and API.

#### Scenario: Refresh and retrieve
- WHEN creation returns a task ID and the page is refreshed
- THEN only that existing task is queried and downloaded through the original bound account
- AND another user or API key cannot retrieve that task.

#### Scenario: Billing and failure
- WHEN a task succeeds and is polled or downloaded repeatedly
- THEN settlement occurs once using the configured USD/second rate and request duration
- AND queued, failed, rejected or ambiguous creation does not trigger automatic paid resubmission.

### Requirement: Safe integration
The system SHALL reject missing prices and unsupported parameters before upstream creation, moderate all image inputs and keep creative content out of server persistence. Existing Grok, GPT, Gemini and Seedance behavior SHALL remain covered by regression checks.
