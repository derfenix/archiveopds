# archiveopds

**English:** [README_EN.md](README_EN.md)

OPDS-сервер для локального каталога книг, разложенного по внешним ZIP-томам и индексу **INPX** (как в типичных «библиотечных» раздачах). Клиенты (например, читалки с поддержкой OPDS) получают навигацию, поиск и ссылки на выдачу файлов.

## Флибуста

**Программа изначально ориентирована на разметку и комплект файлов архива сайта [Флибуста](https://flibusta.is/)** (каталог `.inpx`, строки `.inp`, тома `.zip` с FB2 и др.). Другие коллекции с тем же форматом тоже могут подойти, но сценарий и проверки делались вокруг этого варианта. Это не официальный проект Флибусты и не аффилирован с сайтом.

## Сборка и запуск

Требуется Go из `go.mod` (см. версию в файле).

```bash
make              # бинарник в bin/archiveopds
./bin/archiveopds serve --archive /path/to/library/root --base-url http://127.0.0.1:8080
```

Установка в систему (нужны права root): `sudo make install`, затем `sudo systemd-sysusers`, скопируйте `environment.example` в `/etc/archiveopds/environment`, `systemctl enable --now archiveopds`. Переменные: `PREFIX`, `DESTDIR` (см. `make help`).

**Arch Linux:** [deploy/archlinux/PKGBUILD](deploy/archlinux/PKGBUILD) качает **архив исходников тега** с Forgejo (`$url/archive/v$pkgver.tar.gz`, как в Gitea). В `PKGBUILD` выставьте `pkgver` под существующий тег **`v$pkgver`**, затем `updpkgsums` и `makepkg -si` из `deploy/archlinux`. При необходимости замените `url` на свой инстанс.

Список флагов и переменных окружения:

```bash
./archiveopds config
```

Подробнее про память, поиск, лимиты выдачи книг и завершение процесса — в [docs/runtime.md](docs/runtime.md).

**systemd:** unit [deploy/systemd/archiveopds.service](deploy/systemd/archiveopds.service) читает переменные из **`/etc/archiveopds/environment`** (префикс `ARCHIVEOPDS_*`, как у `archiveopds config`); образец — [deploy/systemd/archiveopds.env.example](deploy/systemd/archiveopds.env.example). Пользователь `archiveopds` можно завести через [deploy/sysusers.d/archiveopds.conf](deploy/sysusers.d/archiveopds.conf) и **`systemd-sysusers`**. Бинарник в unit задан как `/usr/local/bin/archiveopds` — при необходимости поправьте путь или используйте `systemctl edit`.

## CI и релизы

- **GitHub Actions:** [`.github/workflows/`](.github/workflows/) — на push/PR в `main` или `master` запускаются `go test` и `go build`; при push тега вида `v*` (например `v1.0.0`) — сборка артефактов через [GoReleaser](https://goreleaser.com/) и публикация GitHub Release (бинарники linux/windows/darwin, amd64/arm64, архивы + `checksums.txt`).
- **Forgejo / Gitea Actions:** дубликаты workflow в [`.forgejo/workflows/`](.forgejo/workflows/).

На Forgejo GoReleaser определяет хост по `origin` и создаёт релиз через API инстанса; в job передаётся `GITHUB_TOKEN` — так называется встроенный токен в совместимом слое Actions (достаточно прав **contents: write** для репозитория).

Локальная проверка упаковки без публикации:

```bash
go run github.com/goreleaser/goreleaser/v2@v2.4.8 release --snapshot --clean
```

## Лицензия

Проект распространяется на условиях **GNU General Public License v3.0**. Полный текст — в файле [LICENSE](LICENSE).
