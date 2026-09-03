#!/usr/bin/env bash
# bump.sh — чеклист публикации sprob (О5.4 TECH_DEBT.md).
#
# Правила:
#   - Версия — семантическая: v1.0.2xx — фиксы/мелочи; v1.1.x — новые модули/руты;
#   - Миграции модулей: ДО публикации проверить уникальность таймстампов
#     (bun-migrate пишет по числовому префиксу — дубликат «съедает» миграцию);
#   - НЕ публиковать поверх чужого незакоммиченного WIP (проверяет скрипт);
#   - После публикации — обновить ВСЕ проекты: rdkb ×5, portal, pros
#     (go get github.com/pro-assistance-dev/sprob@v<ver>) и таблицу версий в AGENT.md.
#
# Использование:
#   ./scripts/bump.sh v1.0.251            # тег + push + подсказки по синхронизации
#   SKIP_TAG=1 ./scripts/bump.sh v1.0.251 # только проверки (без тега/push)

set -euo pipefail
cd "$(dirname "$0")/.."

VER="${1:-}"
if [[ -z "$VER" ]]; then
  echo "Использование: $0 v<версия>   (например: $0 v1.0.251)" >&2
  exit 1
fi

echo "==> Публикация sprob $VER"

# 1. Рабочее дерево чистое (нет чужого WIP)?
if [[ -n "$(git status --porcelain)" ]]; then
  echo "!! В рабочем дереве есть незакоммиченные изменения — публикация поверх WIP опасна." >&2
  echo "   Закоммитьте/откатите их, затем повторите." >&2
  exit 1
fi

# 2. Сборка, vet, тесты
echo "==> go build ./..."
go build ./... || exit 1
echo "==> go vet ./..."
go vet ./... || exit 1
echo "==> go test ./... (без БД: helpers/http, helpers/project, config — с .env.test)"
go test ./helpers/... ./config/... 2>&1 | grep -vE '^ok|no test files' || true

# 3. Уникальность таймстампов миграций (все модули + корень)
#    Сравниваются ЧИСЛОВЫЕ версии (первые 14 цифр имени), а не полные имена:
#    у bun версия — числовой префикс, разные суффиксы имени не спасают от коллизии.
#    Допустимая пара для одной версии — только `*.up.sql` + `*.down.sql` одной миграции.
echo "==> Миграции: проверка уникальности версий"
DUPS=$(find . -path '*/migrations/*.sql' -type f -printf '%f\n' \
  | awk '{ f=$0; sub(/\.sql$/, "", f); v=substr(f,1,14); k=""; \
           if (f ~ /\.up$/) k="up"; else if (f ~ /\.down$/) k="down"; print v, k }' \
  | sort \
  | awk '{ if ($1==v) { n++; if ($2=="up") up=1; else if ($2=="down") down=1 } \
           else { check(); v=$1; n=1; up=0; down=0; \
                  if ($2=="up") up=1; else if ($2=="down") down=1 } } \
         END { check() } \
         function check() { if (n>1 && !(n==2 && up==1 && down==1)) \
           print v " x" n }')
if [[ -n "$DUPS" ]]; then
  echo "!! Коллизия таймстампов миграций (версия встречается >1 раза, пара up+down не легитимна):" >&2
  echo "$DUPS" | sed 's/^/    /' >&2
  exit 1
fi
echo "    OK"

# 4. Тег + push
if [ "${SKIP_TAG:-0}" = "1" ]; then
  echo "==> SKIP_TAG=1 — тег и push пропущены"
else
  # тег уже есть?
  if git rev-parse "$VER" >/dev/null 2>&1; then
    echo "!! Тег $VER уже существует: $(git log -1 --oneline "$VER")" >&2
    exit 1
  fi
  git tag "$VER"
  git push origin master
  git push origin "$VER"
  echo "==> Опубликован тег $VER"
fi

# 5. Подсказки по синхронизации проектов
cat <<EOF

==> Дальше (вручную, по правилу синхронизации AGENT.md):
    for p in ~/prog/work/{rdkb/hr,rdkb/map,rdkb/food,rdkb/leiter,rdkb/incident,portal/portal,pros}/server; do
      (cd \$p && go get github.com/pro-assistance-dev/sprob@$VER && go build ./...)
    done
    # + коммит go.mod/go.sum в каждом проекте и обновление таблицы версий в AGENT.md
EOF
