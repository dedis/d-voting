# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Latest changes in each category go to the top

## [Unreleased]

### Added
- Changelog - please use it

### Changed
- for the Dockerfiles and docker-compose.yml, `DELA_NODE_URL` has been replaced with `DELA_PROXY_URL`,
 which is the more accurate name.
- the actions in package.json for the frontend changed. Both are somewhat development mode,
 as the webserver is not supposed to be used in production. 
  - `start`: starts in plain mode 
  - `start-https`: starts in HTTPS mode

### Deprecated
### Removed
### Fixed
- When fetching form and user updates, only do it when showing the activity
- Redirection when form doesn't exist and nicer error message
- File formatting and errors in comments
- Popup when voting and some voting translation fixes
- Fixed return error when voting

### Security
- Use `REACT_APP_RANDOMIZE_VOTE_ID === 'true'` to indicate randomizing vote ids
