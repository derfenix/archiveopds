# Установка: PREFIX по умолчанию /usr/local; для пакетов — make install DESTDIR=/tmp/stage
PREFIX      ?= /usr/local
BINDIR      ?= $(PREFIX)/bin
DESTDIR     ?=

OUTPUT      ?= bin/archiveopds
LDFLAGS     ?= -s -w

SYSTEMD_UNIT_DIR ?= $(DESTDIR)/etc/systemd/system
SYSUSERS_DIR     ?= $(DESTDIR)/usr/lib/sysusers.d
ARCHIVEOPDS_ETC  ?= $(DESTDIR)/etc/archiveopds

.PHONY: all build test test-race lint clean install uninstall help

all: build

help:
	@echo "Targets:"
	@echo "  make / make build  — собрать $(OUTPUT)"
	@echo "  make test          — go test ./..."
	@echo "  make test-race     — go test -race ./... (slower, для CI/локали)"
	@echo "  make lint          — golangci-lint (install: https://golangci-lint.run/usage/install/)"
	@echo "  make install       — бинарник, systemd unit, sysusers, пример env"
	@echo "  make uninstall     — удалить установленные файлы"
	@echo "  make clean         — убрать $(OUTPUT)"
	@echo "Variables: PREFIX=$(PREFIX) DESTDIR=… OUTPUT=$(OUTPUT)"

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
	install -Dm644 deploy/systemd/archiveopds.service $(SYSTEMD_UNIT_DIR)/archiveopds.service
	install -Dm644 deploy/sysusers.d/archiveopds.conf $(SYSUSERS_DIR)/archiveopds.conf
	install -d -m755 $(ARCHIVEOPDS_ETC)
	install -Dm644 deploy/systemd/archiveopds.env.example $(ARCHIVEOPDS_ETC)/environment.example

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/archiveopds
	rm -f $(SYSTEMD_UNIT_DIR)/archiveopds.service
	rm -f $(SYSUSERS_DIR)/archiveopds.conf
	rm -f $(ARCHIVEOPDS_ETC)/environment.example
