#!/usr/bin/env bash
# Go モジュールの cooldown チェック。
# api/go.mod が解決するモジュールのうち、公開から COOLDOWN_DAYS 日(既定 7 日)
# 未満のバージョンがあれば失敗する。
#
# Go には npm の min-release-age に相当する仕組みがないため、手動の
# `go get` で新しすぎるバージョンが入るのをこのスクリプトで検出する。
#
# 脆弱性修正などで cooldown を待てない場合は、.cooldown-exempt に
# `モジュール@バージョン` を理由のコメント付きで追記する。
set -euo pipefail

days=${COOLDOWN_DAYS:-7}
root=$(cd "$(dirname "$0")/.." && pwd)
exempt_file="$root/.cooldown-exempt"
cutoff=$(date -u -d "$days days ago" +%Y-%m-%dT%H:%M:%SZ)

modules=$(cd "$root/api" && go list -m \
  -f '{{if and .Time (not .Main)}}{{.Time.UTC.Format "2006-01-02T15:04:05Z"}} {{.Path}}@{{.Version}}{{end}}' all)

violations=""
while read -r published module; do
  [[ -z "$module" ]] && continue
  [[ "$published" > "$cutoff" ]] || continue
  if [[ -f "$exempt_file" ]] && sed 's/[[:space:]]*#.*//' "$exempt_file" | grep -qxF "$module"; then
    echo "除外(.cooldown-exempt): $module (公開: $published)"
    continue
  fi
  violations+="  $module (公開: $published)"$'\n'
done <<<"$modules"

if [[ -n "$violations" ]]; then
  echo "公開から $days 日未満の Go モジュールがあります:" >&2
  printf '%s' "$violations" >&2
  echo "cooldown を待つか、理由を添えて .cooldown-exempt に追記してください。" >&2
  exit 1
fi
echo "OK: すべての Go モジュールが公開から $days 日以上経過しています(基準: $cutoff 以前)"
