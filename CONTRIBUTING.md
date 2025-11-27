# Contributing

## Table of Contents

- [Guidelines](#guidelines)
- [Disclaimer](#disclaimer)
- [Getting started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Testing](#testing)

## Guidelines

See [GUIDELINES.md](GUIDELINES.md) for more information.

## Disclaimer

This guide is written for UNIX-like systems (Linux, macOS). Users on Windows or other platforms may need to adapt some commands.

## Getting started

### Prerequisites

- [Python 3.7+](https://www.python.org/downloads/)
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

## Testing

Format the code:

```shell script
make format
```

Run functional tests:

```shell script
make test
```
