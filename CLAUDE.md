# CLAUDE.md

Guidance for working in the `burmilla/os` repository (BurmillaOS, a maintained fork of RancherOS).

## What BurmillaOS is

A minimal Linux where almost everything runs in containers:

- **Kernel** comes from `burmilla/os-kernel` releases (`KERNEL_VERSION` in `Dockerfile.dapper`).
- **initrd and rootfs** are built from buildroot tarballs published by `burmilla/os-initrd-base` and `burmilla/os-base` (`OS_INITRD_BASE_URL_*`, `OS_BASE_URL_*`).
- **System Docker** (PID 1 side, stripped Docker 17.06 from `burmilla/os-system-docker`) runs system services: console, ntp, syslog, udev, network, user Docker, etc.
- **User Docker** runs inside the console container's namespaces; the engine version is selected with `rancher.docker.engine` / `ros engine switch` from the `burmilla/os-services` index.
- **`ros`** (this repo, one multicall Go binary) is init, the CLI and most helpers (`cloud-init-save`, `console-init`, `docker-init`, `respawn`, ...). `main.go` dispatches on `argv[0]`.

## Repository layout

- `main.go` – multicall entrypoint mapping binary names to `cmd/*` functions.
- `cmd/control/` – the `ros` CLI and in-container helpers (`console_init.go`, `docker_init.go`, `user_docker.go`, `engine.go`, `install.go`, `os.go` for upgrades, `service/` for `ros service`).
- `cmd/init`, `pkg/init/` – boot sequence (`pkg/init/hypervisor` auto-enables VM tools, `pkg/init/modules` loads modules, etc.).
- `pkg/compose/` – starts system services; `reload.go` resolves services and the user Docker engine from the os-services repository.
- `pkg/libcompose/` – thin compatibility layer over `compose-go` (v1 compose fields like `net:` are mapped in `config/compat.go`).
- `pkg/dfs/` – System Docker bootstrap; `pkg/netconf`, `pkg/hostname`, `pkg/sysinit`, `pkg/util`.
- `os-config.tpl.yml` – default system config (services, repositories, kernel cmdline defaults). Templated at build time.
- `images/` – system container images built during the build (`00-rootfs`, `01-base`, `02-console` = Debian 13 default console, `02-syslog`, ...).
- `scripts/` – dapper build steps (`build`, `package-*`, `release-*`), boot configs (`global.cfg`, `isolinux.cfg`, `grub.cfg`) and dev helpers (`run`, `run-common`).
- `config/` – Go config types and schema; `vendor/` – vendored Go modules.

## Build and test

Everything builds inside dapper (`Dockerfile.dapper`, needs Docker on the host):

- `make pr-validation` – what CI runs on pull requests (validation without a full kernel/image build).
- `make release` – full build; artifacts land in `dist/` (ISO, `vmlinuz`, `initrd`, `rootfs.tar.gz`).
- `make run` – build and boot in QEMU for manual testing. `make hyperv`, `make vmware`, `make azurebase` build platform ISOs with preloaded service images.

Outside dapper, Go code can be checked directly (Go modules, Go version in `go.mod`):

```sh
go build ./...
go vet ./cmd/... ./pkg/...
go test ./cmd/... ./pkg/...
```

Many code paths need root and a real System Docker, so unit tests cover only parts of the logic; boot-test changes to init, console or Docker startup.

## Release workflow

1. Bump pinned components in `Dockerfile.dapper` (kernel, os-base, os-initrd-base, system-docker) when their repos publish new releases.
2. Merge to `master`, tag `vX.Y.Z[-betaN|-rcN]`, then run the `release` workflow (`.github/workflows/create-release.yml`, manual dispatch from the tag). It runs `make release` and `dist/publish.sh` (Docker Hub images + GitHub release assets).
3. Add the version to `burmilla/releases` (branch `v3.0.x`, used via `OS_RELEASES_YML`) so `ros os upgrade` offers it.
4. Update docs in `burmilla/burmilla.github.io`.

Related repositories:

- `burmilla/os-services` – optional services and Docker engines. v3 uses branch `v3.0.x` (`OS_SERVICES_REPO` + `/v3.0.x` in `os-config.tpl.yml`). Service images are built by its `build-service` / `build-docker` workflows, which skip tags that already exist on Docker Hub, so image changes need a new tag.
- `burmilla/os-kernel`, `burmilla/os-base`, `burmilla/os-initrd-base`, `burmilla/os-system-docker`.

## v3.0 status (written during v3.0.0-rc1 preparation)

v3.0.0-beta1 shipped kernel 6.12.95, buildroot 2025.02.16, Debian 13 console, System Docker 17.06.113, `ros` moved from libcompose to compose-go with Go modules, cgroup v2 and UEFI installation.

Known beta1 issues (release discussion #226) and their RC1 fixes:

- **AppArmor not enabled** – the Debian 13 console contains systemd, so Debian's `/etc/init.d/apparmor` exits early inside a container. `console-init` now mounts securityfs and loads `/etc/apparmor.d` itself, before `/run/console-done` so user Docker sees AppArmor. Kernel cmdline already has `apparmor=1 security=apparmor`.
- **Hyper-V tools crashing** – os-services image built with bookworm glibc but run on bullseye; fixed in os-services (`os-hypervvmtools:6.12.114-burmilla-1`, built for kernel 6.12.114), preload tags updated in `Makefile`. The image tag follows the kernel version, so rebuild it when the kernel is bumped.
- **ZFS service enable fails** – OpenZFS 2.0.1 does not build for 6.12; os-services moved to OpenZFS 2.4.4 on a Debian 13 build image (`os-zfs:v2.4.4-1`).
- **Docker unavailable after v2 → v3 upgrade** – the selected engine (e.g. docker-29.6.2) was missing from os-services `v3.0.x`. Missing engines were added there, and `pkg/compose/reload.go` now falls back to the default/latest engine when the configured one is not in the index.
- **multipath-tools** added to the console on user request.

Open items / caveats:

- arm64 kernel config has AppArmor disabled.
- `ros engine switch latest` and the engine fallback sort engine names as strings (29.10 < 29.9).
- Hyper-V VSS and fcopy daemons are built but not started.
- Fixes above were verified by build/static analysis only; boot-test on real hardware/VMs before the RC.
