#!/usr/bin/env bash
# Grype で SBOM を脆弱性スキャンする。判定基準は .grype.yaml に従う。
#
# 使い方: sbom-scan.sh <SBOM ファイル>
#
# 結果は標準出力に表で出し、SBOM と同じ場所に .grype.json / .grype.sarif としても保存する。
# GitHub Actions 上ではジョブサマリーにも書き出す。
# 終了コード: 0 = 基準内 / 2 = 基準を超える脆弱性あり / その他 = スキャン自体の失敗
set -uo pipefail

if [[ $# -lt 1 ]]; then
  sed -n '2,8p' "$0" >&2
  exit 64
fi

sbom=$1
base=${sbom%.cdx.json}
base=${base%.spdx.json}
json="$base.grype.json"
table="$base.grype.txt"
sarif="$base.grype.sarif"

"${GRYPE:-grype}" "sbom:$sbom" -o "json=$json" -o "table=$table" -o "sarif=$sarif"
status=$?

if [[ -s "$table" ]]; then
  cat "$table"
else
  echo "脆弱性は検出されませんでした"
fi

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "### 脆弱性スキャン: $(basename "$sbom")"
    echo
    case $status in
      0) echo "✅ ポリシー(.grype.yaml)の基準内です。" ;;
      2) echo "❌ ポリシー(.grype.yaml)の基準を超える脆弱性があります。" ;;
      *) echo "⚠️ スキャンに失敗しました(終了コード $status)。" ;;
    esac
    echo
    if [[ -s "$table" ]]; then
      echo '```'
      cat "$table"
      echo '```'
    else
      echo "検出なし"
    fi
  } >>"$GITHUB_STEP_SUMMARY"
fi

exit $status
