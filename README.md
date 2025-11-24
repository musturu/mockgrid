# mockgrid

[![Security Scan](https://github.com/musturu/mockgrid/actions/workflows/scan.yml/badge.svg?branch=main)](https://github.com/musturu/mockgrid/actions/workflows/scan.yml)
[![License](https://img.shields.io/github/license/musturu/mockgrid)](LICENSE)

mockgrid is a small, mock email server that mimics parts of SendGrid's API so you can test sending emails and template rendering locally without talking to the real SendGrid service.

## Scope
- Provide an HTTP API compatible with common SendGrid flows used by applications (mail send, templates, tracking).
- Allow sending emails through a real SMTP server configured by the operator (so you can wire tests to a local or remote SMTP endpoint).
- Support template rendering from local templates or (optionally) a SendGrid-like template provider.
- Provide attachment handling for testing file uploads and delivery.

## Key functionalities
- HTTP endpoint that accepts SendGrid-like POST requests and forwards messages via SMTP.
- Template rendering engine with support for local template directories and remote templates.
- Attachment handling with secure, temporary storage.
- Tracking pixel support for open tracking (test-friendly).
- Configurable via environment variables, config file, or CLI flags.

## Development status
This project is actively under development. Features, config options, and APIs are subject to change. Use it for testing and experimentation, and expect behavior to evolve.

## Credits
Special thanks to `ykanezawa` and their `sendgrid-dev` (https://github.com/yKanazawa/sendgrid-dev)  repository — their work was a helpful starting point for this project.

## Contributing
- Bug reports and PRs welcome. Please open issues for design discussions before large changes.

# License
This project is published under the terms in the `LICENSE` file.
