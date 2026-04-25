# archiveopds

**Русский:** [README.md](README.md)

OPDS server for a local book library laid out as external **ZIP** volumes plus an **INPX** index (as in typical “library” distributions). Clients (e.g. OPDS-capable readers) get navigation, search, and acquisition links for files.

## Flibusta

**The program is primarily aimed at the layout and file set of the [Flibusta](https://flibusta.is/) library archive** (`.inpx` catalogue, `.inp` lines, `.zip` volumes with FB2, etc.). Other collections using the same format may work, but development and checks target this case. This is **not** an official Flibusta project and is **not** affiliated with the site.

## Build and run

Use the Go version from `go.mod`.

```bash
make              # binary at bin/archiveopds
./bin/archiveopds serve --archive /path/to/library/root --base-url http://127.0.0.1:8080
```

System install (root): `sudo make install`, then `sudo systemd-sysusers`, copy `environment.example` to `/etc/archiveopds/environment`, `systemctl enable --now archiveopds`. Variables: `PREFIX`, `DESTDIR` (see `make help`).

**Arch Linux:** [deploy/archlinux/PKGBUILD](deploy/archlinux/PKGBUILD) downloads the **tag source archive** from Forgejo (`$url/archive/v$pkgver.tar.gz`, Gitea-style). Set `pkgver` in `PKGBUILD` to match an existing `**v$pkgver`** tag, then `updpkgsums` and `makepkg -si` from `deploy/archlinux`. Change `url` if you use another instance.

Flags and environment variables:

```bash
./archiveopds config
```

Memory usage, search behaviour, book delivery limits, and process shutdown are described in [docs/runtime.md](docs/runtime.md) (Russian).

**systemd:** [deploy/systemd/archiveopds.service](deploy/systemd/archiveopds.service) reads `**/etc/archiveopds/environment`** (`ARCHIVEOPDS_*` prefix, same as `archiveopds config`); example — [deploy/systemd/archiveopds.env.example](deploy/systemd/archiveopds.env.example). Create the `archiveopds` user with [deploy/sysusers.d/archiveopds.conf](deploy/sysusers.d/archiveopds.conf) and `**systemd-sysusers`**. The unit expects the binary at `/usr/local/bin/archiveopds` — adjust the path or use `systemctl edit` if needed.

## CI and releases

- **GitHub Actions:** `[.github/workflows/](.github/workflows/)` — on push/PR to `main` or `master`, [golangci-lint](https://golangci-lint.run/) (see `[.golangci.yml](.golangci.yml)`), `go test`, and `go build`; on push of a `v`* tag (e.g. `v1.0.0`), [GoReleaser](https://goreleaser.com/) builds artifacts and publishes a GitHub Release (linux/windows/darwin, amd64/arm64, archives + `checksums.txt`).
- **Local lint:** `make lint` (requires `golangci-lint` on `PATH`).
- **Forgejo / Gitea Actions:** workflow copies under `[.forgejo/workflows/](.forgejo/workflows/)`.

On Forgejo, GoReleaser uses `origin` and the instance API to publish releases; the job uses `GITHUB_TOKEN` — the built-in Actions-compat token (needs **contents: write** on the repo).

Local packaging check without publishing:

```bash
go run github.com/goreleaser/goreleaser/v2@v2.4.8 release --snapshot --clean
```

## License

Licensed under the **GNU General Public License v3.0**. Full text: [LICENSE](LICENSE).