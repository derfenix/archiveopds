# archiveopds

OPDS-сервер для локального каталога книг, разложенного по внешним ZIP-томам и индексу **INPX** (как в типичных «библиотечных» раздачах). Клиенты (например, читалки с поддержкой OPDS) получают навигацию, поиск и ссылки на выдачу файлов.

## Флибуста

**Программа изначально ориентирована на разметку и комплект файлов архива сайта [Флибуста](https://flibusta.is/)** (каталог `.inpx`, строки `.inp`, тома `.zip` с FB2 и др.). Другие коллекции с тем же форматом тоже могут подойти, но сценарий и проверки делались вокруг этого варианта. Это не официальный проект Флибусты и не аффилирован с сайтом.

## Сборка и запуск

Требуется Go из `go.mod` (см. версию в файле).

```bash
go build -o archiveopds ./cmd/archiveopds
./archiveopds serve --archive /path/to/library/root --base-url http://127.0.0.1:8080
```

Список флагов и переменных окружения:

```bash
./archiveopds config
```

Подробнее про память, поиск, лимиты выдачи книг и завершение процесса — в [docs/runtime.md](docs/runtime.md).

## CI и релизы

- **GitHub Actions:** [`.github/workflows/`](.github/workflows/) — на push/PR в `main` или `master` запускаются `go test` и `go build`; при push тега вида `v*` (например `v1.0.0`) — сборка артефактов через [GoReleaser](https://goreleaser.com/) и публикация GitHub Release (бинарники linux/windows/darwin, amd64/arm64, архивы + `checksums.txt`).
- **Forgejo / Gitea Actions:** те же сценарии в [`.gitea/workflows/`](.gitea/workflows/) (типичный путь для Codeberg и многих инстансов Forgejo) и дубликат в [`.forgejo/workflows/`](.forgejo/workflows/), если администратор включил только каталог `.forgejo/workflows`.

На Forgejo GoReleaser определяет хост по `origin` и создаёт релиз через API инстанса; в job передаётся `GITHUB_TOKEN` — так называется встроенный токен в совместимом слое Actions (достаточно прав **contents: write** для репозитория).

Локальная проверка упаковки без публикации:

```bash
go run github.com/goreleaser/goreleaser/v2@v2.4.8 release --snapshot --clean
```

## Лицензия

Проект распространяется на условиях **GNU General Public License v3.0**. Полный текст — в файле [LICENSE](LICENSE).
