# Gexec

[![General Workflow](https://github.com/gexec/gexec/actions/workflows/general.yml/badge.svg)](https://github.com/gexec/gexec/actions/workflows/general.yml) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/10812ff088364821976ecaf4341a0225)](https://app.codacy.com/gh/gexec/gexec/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Discord](https://img.shields.io/discord/1335976189025849395)](https://discord.gg/Yda8rD4ZkJ) [![Go Reference](https://pkg.go.dev/badge/github.com/gexec/gexec.svg)](https://pkg.go.dev/github.com/gexec/gexec) [![GitHub Repo](https://img.shields.io/badge/github-repo-yellowgreen)](https://github.com/gexec/gexec) [![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com)

> [!CAUTION]
> This project is in active development and does not provide any stable release
> yet, you can expect breaking changes until our first real release!

With Gexec we are building a generic execution platform for Ansible, OpenTofu
and Terraform. Besides that it should be possible to execute any kind of script
which is supported by [Asdf][asdf] plugins. Some plugins are installed as part
of the containers, if you are installing this project differently it is up to
you to install and configure [Asdf][asdf].

## Install

You can download prebuilt binaries from the [GitHub releases][releases] or from
our [download site][downloads]. Besides that we also prepared repositories for
DEB and RPM packages which can be found at [Cloudsmith][pkgrepo]. If you prefer
to use containers you could use our images published on [GHCR][ghcr]. You are a
Mac user? Just take a look at our [homebrew formula][homebrew]. If you need
further guidance how to install this take a look at our [documentation][docs].

Package repository hosting is graciously provided by [Cloudsmith][cloudsmith].
Cloudsmith is the only fully hosted, cloud-native, universal package management
solution, that enables your organization to create, store and share packages in
any format, to any place, with total confidence.

## Prerequisites

We use [mise][mise] to manage all required tools and their versions. Install it
by following the [official installation instructions][mise-install], then run
the following commands inside the repository to activate mise and install all
tools defined in `mise.toml`:

```console
mise trust
mise install
```

## Build

Since all required commands ar part of our [go-task][gotask] taskfile the
commands you got to execute are quite simple:

```console
git clone https://github.com/gexec/gexec.git
cd gexec

task fe:install fe:build be:build
./bin/gexec-server -h
```

We are embedding all the static assets into the binary so there is no need for
any webserver or anything else beside launching this binary.

## Development

If you are using the provided [DevContainers][devcontainer] you could directly
start without installing [mise][mise] on your system since it will launch a
feature for it.

To start developing on this project you have to execute only a few commands in
multiple terminal tabs or windows:

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

## Security

If you find a security issue please contact
[gexec@webhippie.de](mailto:gexec@webhippie.de) first.

## Contributing

Generally we are following [conventional commits][commits] when we apply
changes. That way we are able to generate proper changelogs for every release.
Please use always pull requests to integrate new functionalities or to fix
issues.

For the release process we are following [semantic versioning][semver] which
clearly indicates if a new version just resolves bugs, includes new features or
even includes breaking changes.

After installing the tools via `mise install` as described above set up the
pre-commit hooks so they run automatically on every commit:

```console
pre-commit install --hook-type pre-commit --hook-type commit-msg
```

> `pre-commit` is managed by mise and will be available after `mise install`.

If you have changed something on the source you should simply commit following
the mentioned conventions:

```console
git checkout -b feat/new-feature
git add --all
git commit -m 'feat: added awesome new feature'
git push --set-upstream origin feat/new-feature
```

After pushing your changes into the Git repository you should create a pull
request on GitHub. If the pull request have been merged and everything built
fine it will also create automatically a new release at least once a week.

## Authors

*   [Thomas Boerger](https://github.com/tboerger)

## License

Apache-2.0

## Copyright

```console
Copyright (c) 2025 Thomas Boerger <thomas@webhippie.de>
```

[asdf]: https://asdf-vm.com/
[releases]: https://github.com/gexec/gexec/releases
[downloads]: http://dl.gexec.eu
[pkgrepo]: https://cloudsmith.io/~gexec/repos/general/groups/
[ghcr]: https://github.com/orgs/gexec/packages
[homebrew]: https://github.com/gexec/homebrew-gexec
[docs]: https://gexec.eu
[cloudsmith]: https://cloudsmith.com/
[gotask]: https://taskfile.dev/installation/
[devcontainer]: https://containers.dev/
[mise]: https://mise.jdx.dev/
[mise-install]: https://mise.jdx.dev/getting-started.html
[commits]: https://www.conventionalcommits.org/en/v1.0.0/
[semver]: https://semver.org/
