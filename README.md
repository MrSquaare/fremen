# Fremen

A fast, parallelized security scanner for detecting infected packages in lockfiles.

## Table of Contents

- [About](#about)
  - [Built with](#built-with)
  - [Acknowledgments](#acknowledgments)
- [Getting started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Using](#using)
- [Contributing](#contributing)
- [License](#license)

## About

Fremen is a security tool designed to scan your project's lockfiles for known malicious package versions. It helps protect your supply chain by identifying compromised dependencies before they are installed or deployed.

It currently supports:
- **npm** (`package-lock.json`)
- **Yarn** (`yarn.lock`)
- **pnpm** (`pnpm-lock.yaml`)

Fremen is designed for performance, utilizing parallel execution to scan large directories and monorepos efficiently.

### Built with

- [Golang](https://go.dev/)

### Acknowledgments

- This project is based on the work of [Cobenian/shai-hulud-detect](https://github.com/Cobenian/shai-hulud-detect).
- It uses the same database format for identifying vulnerable package versions.

## Disclaimer

This tool is developed and tested primarily on UNIX-like systems (Linux, macOS).
The code has been written with the help of AI tools.

## Getting started

### Prerequisites

- Go 1.25+ installed on your system. You can download it from [go.dev](https://go.dev/dl/).

### Installation

1. Download the binary from the [Releases](https://github.com/MrSquaare/fremen/releases) page

2. Download or create a `database.txt` file containing the list of infected packages. By default, Fremen looks for `database.txt` where the script is located.

## Using

Run the scanner against your project directories:

```shell script
./fremen [paths...]
```

### Common Options

- **Recursive Scan:** Scan the current directory and all subdirectories.
  ```shell script
  ./fremen -r
  ```

- **Include Ignored Directories:** By default, `.git` and `node_modules` are ignored. You can include them if needed:
  ```shell script
  ./fremen -r --include-git --include-node-modules
  ```

- **Specify Database:** Use a custom database file.
  ```shell script
  ./fremen -d /path/to/database.txt
  ```

- **Full Report:** Show all projects, including clean ones.
  ```shell script
  ./fremen --full-report
  ```

- **JSON Output:** Generate a machine-readable JSON report.
  ```shell script
  ./fremen --json
  ```

For a full list of options, run:
```shell script
./fremen --help
```

## Contributing

Bug reports, feature requests, other issues and pull requests are welcome.
See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

## License

Distributed under the [MIT](https://choosealicense.com/licenses/mit/) License.
See [LICENSE](LICENSE) for more information.
