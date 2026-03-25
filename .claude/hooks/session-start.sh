#!/bin/bash
set -euo pipefail

# Запускати лише у віддаленому Claude Code середовищі
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

cd "$CLAUDE_PROJECT_DIR"

# Налаштування git credentials
if [ -n "${GITHUB_TOKEN:-}" ]; then
  git config --global credential.helper store
  echo "https://oauth2:${GITHUB_TOKEN}@github.com" > ~/.git-credentials
  git remote set-url origin "https://github.com/ZaevIhor/logbot.git"
fi

# Завантаження Go залежностей
go mod download

# Перевірка збірки
go build ./...

# Запуск тестів
go test ./...
