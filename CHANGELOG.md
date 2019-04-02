# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Custom copy/paste fields (ex: `kubectl config set-cluster [...]`)
- More prometheus exports
- UX improvements

## [2.6.0] - 2019-04-02
### Added
- Prometheus endpoint (:9090 by default)

  Current available metrics:
  - loginapp_request_total{"code", "method"}: Counter vector
  - loginapp_request_duration{"code", "method"}: Gauge vector

  More metrics will be added in feature minor upgrades

- Support extra auth code options.

  loginapp now supports extra auth code options, to ensure
  compatibility with IdP like ADFS (see issue #16)
  Configuration option:
  ```
  [...]
  oidc:
    extra_auth_code_opts:
      resource: XXXXX
  [...]
  ```

- Changelog :)

### Changed
- Code refactor: to easily add prometheus metrics, we had to
  split code. Current logic is:
  - `cli.go`: CLI related code
  - `config.go`: app configuration
  - `handlers.go`: main router handlers
  - `logging.go`: logging related code
  - `main.go`: app entrypoint
  - `prometheus.go`: prometheus metrics setup
  - `routes.go`: main router setup
  - `server.go`: server related code
  - `templates.go`: html templates
  - `util.go`: another util garbage file...

  Also update checks and go fmt

- Improve user requests logging: add return code

## [2.5.0] - 2018-12-21
### Added
- Show a kubectl based version of configuration update (see PR #12) from (@aveyrenc)[https://github.com/aveyrenc]

## [2.4.1] - 2018-11-05
### Fixed
- License field issues

## [2.4.0] - 2018-10-28
### Changed
- Evacuate html templates from binary.
  It allows users to override default assets.

## [2.3.0] - 2018-10-24
### Added
- Customizable username claim. You can now change output username claim
  in loginapp configuration.

  Default is set to "name". "name" and "email" are
  common claims, the full list of supported claims
  are available at 'well-known' URL of your issuer
  (ex: https://dex.example.com/.well-known/openid-configuration)

- Golang syntax improvements

### Changed
- Configuration checks methods: simply code to easily include
  new configuration option and associated check function

## [2.2.0] - 2018-10-11
### Added
- Debug output

### Changed
- Split CLI setup from main

### Fixed
- `offline_as_scope` option (see PR #4) from [@robbiemcmichael](https://github.com/robbiemcmichael)
- Log typos

## [2.1.0] - 2018-08-15
### Added
- HTML frontend click/copy feature
- Ability to use your own assets
- Skip main page option (`web_output.skip_main_page`)
- Document new opts in README, add a dev doc
- Configuration checks

### Changed
- Code refactoring

### Fixed
- Multiple client_id in html render when using cross_client feature
  (related to option `web_output.main_client_id`).

## [2.0.2] - 2018-07-18
### Added
- Code checks and format: gofmt, errcheck, gocyclo, gosimple

## [2.0.1] - 2018-07-17
### Added
- App description and version in CLI

## [2.0.0] - 2018-07-13
### Added
- Apache 2.0 LICENSE
- Dependencies management
- Better example files
- Real CLI (`loginapp serve [configfile]`)
- Support loglevel configuration
- Exponential retry backoff when at startup when setting up
  provider

- `/healthz` endpoint (for k8s) --> Check: provider is setup, provider availability

### Changed
- Code refactoring
- Cleaner config format
- More debug (ex: middleware logger for incoming requests)
- Move code to root directory

### Removed
- No more `alpine` Docker image, only scratch

## [1.1.1] - 2018-04-24
### Added
- quay.io repository: [fydrah/loginapp](quay.io/fydrah/loginapp)

### Removed
- DockerHub repository

### Fixed
- Typos

## [1.1.0] - 2018-04-06
### Added
- Docker images (scratch and alpine)

### Fixed
- Typos and add precise title for cross client field

### Removed
- Useless entrypoint

## [1.0.0] - 2017-12-28
### Added
- Init


[Unreleased]: https://github.com/fydrah/loginapp/compare/2.6.0...HEAD
[2.5.0]: https://github.com/fydrah/loginapp/compare/2.5.0...2.6.0
[2.4.1]: https://github.com/fydrah/loginapp/compare/2.4.1...2.5.0
[2.4.0]: https://github.com/fydrah/loginapp/compare/2.4.0...2.4.1
[2.3.0]: https://github.com/fydrah/loginapp/compare/2.3.0...2.4.0
[2.2.0]: https://github.com/fydrah/loginapp/compare/2.2.0...2.3.0
[2.1.0]: https://github.com/fydrah/loginapp/compare/2.1.0...2.2.0
[2.0.2]: https://github.com/fydrah/loginapp/compare/2.0.2...2.1.0
[2.0.1]: https://github.com/fydrah/loginapp/compare/2.0.1...2.0.2
[2.0.0]: https://github.com/fydrah/loginapp/compare/2.0.0...2.0.1
[1.1.1]: https://github.com/fydrah/loginapp/compare/1.1.1...2.0.0
[1.1.0]: https://github.com/fydrah/loginapp/compare/1.1.0...1.1.1
[1.0.0]: https://github.com/fydrah/loginapp/compare/1.0.0...1.1.0
