# Установка в пользовательский профиль: ~/.local/bin, ~/.config/systemd/user, ~/.config/archiveopds.
# Пути переопределяются переменными (см. make help); для упаковки — make install DESTDIR=/tmp/stage
PREFIX      ?= $(HOME)/.local
BINDIR      ?= $(PREFIX)/bin
CONFIGDIR   ?= $(HOME)/.config
DESTDIR     ?=

OUTPUT      ?= bin/archiveopds
LDFLAGS     ?= -s -w

SYSTEMD_USER_UNIT_DIR ?= $(CONFIGDIR)/systemd/user
ARCHIVEOPDS_CONFIG    ?= $(CONFIGDIR)/archiveopds

.PHONY: all build test test-race lint clean install uninstall enable disable help

all: build

help:
	@echo "Targets:"
	@echo "  make / make build  — собрать $(OUTPUT)"
	@echo "  make test          — go test ./..."
	@echo "  make test-race     — go test -race ./... (slower, для CI/локали)"
	@echo "  make lint          — golangci-lint (install: https://golangci-lint.run/usage/install/)"
	@echo "  make install       — пользовательский systemd unit, бинарник, пример env"
	@echo "  make enable        — systemctl --user enable --now archiveopds"
	@echo "  make disable       — остановить и отключить пользовательский unit"
	@echo "  make uninstall     — удалить установленные файлы"
	@echo "  make clean         — убрать $(OUTPUT)"
	@echo "Variables: PREFIX=$(PREFIX) CONFIGDIR=$(CONFIGDIR) DESTDIR=… OUTPUT=$(OUTPUT)"

build:
	@mkdir -p $(dir $(OUTPUT))
	go build -trimpath -ldflags="$(LDFLAGS)" -o $(OUTPUT) ./cmd/archiveopds

test:
	go test ./...

test-race:
	go test -race -count=1 ./...

lint:
	golangci-lint run

clean:
	rm -f $(OUTPUT)

install: build
	install -Dm755 $(OUTPUT) $(DESTDIR)$(BINDIR)/archiveopds
	install -Dm644 deploy/systemd/archiveopds.user.service $(DESTDIR)$(SYSTEMD_USER_UNIT_DIR)/archiveopds.service
	install -d -m755 $(DESTDIR)$(ARCHIVEOPDS_CONFIG)
	install -Dm644 deploy/systemd/archiveopds.env.example $(DESTDIR)$(ARCHIVEOPDS_CONFIG)/environment.example
	@if [ -z "$(DESTDIR)" ] && command -v systemctl >/dev/null 2>&1; then \
		systemctl --user daemon-reload 2>/dev/null \
			&& echo "systemd: пользовательский менеджер перезагружен" \
			|| echo "systemd: systemctl --user недоступен, daemon-reload пропущен"; \
	fi
	@echo ""
	@echo "Установлено:"
	@echo "  $(DESTDIR)$(BINDIR)/archiveopds"
	@echo "  $(DESTDIR)$(SYSTEMD_USER_UNIT_DIR)/archiveopds.service"
	@echo "  $(DESTDIR)$(ARCHIVEOPDS_CONFIG)/environment.example"
	@echo ""
	@echo "Дальше:"
	@echo "  cp $(DESTDIR)$(ARCHIVEOPDS_CONFIG)/environment.example $(DESTDIR)$(ARCHIVEOPDS_CONFIG)/environment"
	@echo "  \$$EDITOR $(DESTDIR)$(ARCHIVEOPDS_CONFIG)/environment   # ARCHIVEOPDS_ARCHIVE, ARCHIVEOPDS_BASE_URL, …"
	@echo "  make enable   # systemctl --user enable --now archiveopds"
	@echo "  journalctl --user -u archiveopds -f"

enable:
	systemctl --user daemon-reload
	systemctl --user enable --now archiveopds
	systemctl --user --no-pager --lines=0 status archiveopds

disable:
	-systemctl --user disable --now archiveopds
	systemctl --user daemon-reload

uninstall:
	@if [ -z "$(DESTDIR)" ] && systemctl --user is-enabled --quiet archiveopds 2>/dev/null; then \
		echo "Внимание: юнит archiveopds всё ещё включён; остановить и отключить: make disable"; \
	fi
	rm -f $(DESTDIR)$(BINDIR)/archiveopds
	rm -f $(DESTDIR)$(SYSTEMD_USER_UNIT_DIR)/archiveopds.service
	rm -f $(DESTDIR)$(ARCHIVEOPDS_CONFIG)/environment.example
	-rmdir $(DESTDIR)$(ARCHIVEOPDS_CONFIG) 2>/dev/null || true
	@if [ -z "$(DESTDIR)" ] && command -v systemctl >/dev/null 2>&1; then \
		systemctl --user daemon-reload; \
	fi
	@echo "Удалено (файл $(DESTDIR)$(ARCHIVEOPDS_CONFIG)/environment оставлен, если был создан)."
