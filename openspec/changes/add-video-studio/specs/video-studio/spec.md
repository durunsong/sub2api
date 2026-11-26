## ADDED Requirements

### Requirement: Independent tabs
The system SHALL provide image/video tabs on the existing user and admin routes without changing image generation or image data.
#### Scenario: Switch tabs while generating
- **WHEN** a user opens the video tab
- **THEN** the image draft and running tasks survive, while hidden global controls do not consume video input.

### Requirement: Local sessions and resumable tasks
The system SHALL store creative content only in user-isolated IndexedDB.
#### Scenario: Reload
- **WHEN** a pending record has an upstream ID
- **THEN** only polling resumes, never a paid create request.
#### Scenario: Delete while polling
- **WHEN** a deleted record receives a late response
- **THEN** the conditional update does not recreate it.
#### Scenario: Local quota failure
- **WHEN** saving output fails
- **THEN** the result remains available for download and explicit resaving.

### Requirement: Authorization and privacy
The system SHALL reuse existing group authorization, task ownership and billing, without persisting creative payloads in audit/Ops/errors.
#### Scenario: Unauthorized model or task
- **WHEN** a caller selects an unauthorized model/group or another user's task
- **THEN** the server denies access.

### Requirement: Independent admin settings
The system SHALL independently manage video availability and model labels.
#### Scenario: Filtered bulk editing
- **WHEN** an admin toggles filtered models
- **THEN** only valid filtered rows change and retired overrides can be preserved unchanged or deleted.
