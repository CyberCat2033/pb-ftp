# Changelog

All notable user-facing changes to pb-ftp are documented in this file.

## [1.0.4] - 2026-06-20

### Fixed

- Made the local update endpoint forward-compatible with future Android client metadata fields.
- Refreshed the PocketBook storage scan before restarting after a self-update so the applications list can pick up the replaced launcher more reliably.
- Gave the update response time to reach the Android client before pb-ftp exits for restart.

## [1.0.3] - 2026-06-20

### Added

- Added localized changelog publication for Android-side PocketBook server update prompts.

### Fixed

- Restarted the bundled FTP server automatically when PocketBook wakes after sleep and the server process has exited.
- Added a graceful FTP QUIT handshake before stopping the bundled FTP server on app exit.
