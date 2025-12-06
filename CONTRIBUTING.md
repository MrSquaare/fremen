# Contributing

## Table of Contents

- [Guidelines](#guidelines)
- [Disclaimer](#disclaimer)
- [Getting started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Building](#building)
- [Testing](#testing)

## Guidelines

See [GUIDELINES.md](GUIDELINES.md) for more information.

## Disclaimer

This guide is written for UNIX-like systems (Linux, macOS). Users on Windows or other platforms may need to adapt some commands.

## Getting started

### Prerequisites

- [Go 1.25.4+](https://go.dev/doc/install/)
- [Make](https://www.gnu.org/software/make/)

### Installation

1. Clone the repository:

```shell script
git clone https://github.com/MrSquaare/fremen.git
```

2. Setup the project:

```shell script
make install
```

## Building

Build the project:

```shell script
make build
```

## Testing

Lint the code:

```shell script
make lint
```

Lint the code with automatic fixes:

```shell script
make lint-fix
```

Run integration tests:

```shell script
make test
```

Run integration tests with coverage report:

```shell script
make test-coverage
```
