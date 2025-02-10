# Gexec: Server

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/10812ff088364821976ecaf4341a0225)](https://app.codacy.com/gh/gexec/gexec/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Discord](https://img.shields.io/discord/1335976189025849395)](https://discord.gg/Yda8rD4ZkJ)

> [!CAUTION]
> This project is in active development and does not provide any stable release
> yet, you can expect breaking changes until our first real release!

TBD

## Install

You can download prebuilt binaries from the [GitHub releases][releases] or from
our [download site][downloads]. If you prefer to use containers you could use
our images published on [GHCR][ghcr]. You are a Mac user? Just take a look at
our [homebrew formula][homebrew]. If you need further guidance how to install
this take a look at our [documentation][docs].

## Build

If you are not familiar with [Nix][nix] it is up to you to have a working
environment for Go (>= 1.24.0) and Nodejs (20.x) as the setup won't we covered
within this guide. Please follow the official install instructions for
[Go][golang] and [Nodejs][nodejs]. Beside that we are using [go-task][gotask] to
define all commands to build this project.

```console
git clone https://github.com/gexec/gexec.git
cd gexec

task be:build fe:install fe:build
./bin/gexec -h
```

If you got [Nix][nix] and [Direnv][direnv] configured you can simply execute
the following commands to get al dependencies including [go-task][gotask] and
the required runtimes installed. You are also able to directly use the process
manager of [devenv][devenv]:

```console
cat << EOF > .envrc
use flake . --impure --extra-experimental-features nix-command
EOF

direnv allow
```

We are embedding all the static assets into the binary so there is no need for
any webserver or anything else beside launching this binary.

## Development

To start developing on this project you have to execute only a few commands. To
start development just execute those commands in different terminals:

```console
task watch:server
task watch:runner
task watch:frontend
```

The development server of the backend should be running on
[http://localhost:8080](http://localhost:8080) while the frontend should be
running on [http://localhost:5173](http://localhost:5173). Generally it supports
hot reloading which means the services are automatically restarted/reloaded on
code changes.

If you got [Nix][nix] configured you can simply execute the [devenv][devenv]
command to start the frontend, backend, MariaDB, PostgreSQL and Minio:

```console
devenv up
```

## Security

If you find a security issue please contact
[gexec@webhippie.de](mailto:gexec@webhippie.de) first.

## Contributing

Fork -> Patch -> Push -> Pull Request

## Authors

*   [Thomas Boerger](https://github.com/tboerger)

## License

Apache-2.0

## Copyright

```console
Copyright (c) 2025 Thomas Boerger <thomas@webhippie.de>
```

[releases]: https://github.com/gexec/gexec/releases
[downloads]: http://dl.gexec.eu
[ghcr]: https://ghcr.io/gexec/gexec
[homebrew]: https://github.com/gexec/homebrew-gexec
[docs]: https://gexec.eu
[nix]: https://nixos.org/
[golang]: http://golang.org/doc/install.html
[nodejs]: https://nodejs.org/en/download/package-manager/
[gotask]: https://taskfile.dev/installation/
[direnv]: https://direnv.net/
