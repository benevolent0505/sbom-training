#!/usr/bin/env bash
# Syft で SBOM を生成する。1 回のスキャンで CycloneDX と SPDX の両方を出力する。
#
# 使い方: sbom-generate.sh <source> <name> <version> [出力先ディレクトリ]
#   source  Syft のスキャン対象。例:
#             docker:sbom-training/api:local           ローカルの Docker イメージ
#             registry:ghcr.io/owner/repo/api@sha256:… レジストリ上のイメージ
#             dir:./api                                ソースディレクトリ
#   name    SBOM に記録するコンポーネント名。出力ファイル名にも使う
#   version SBOM に記録するバージョン(Git のコミット SHA やタグ)
set -euo pipefail

if [[ $# -lt 3 ]]; then
  sed -n '2,11p' "$0" >&2
  exit 64
fi

source=$1
name=$2
version=$3
outdir=${4:-sbom}

mkdir -p "$outdir"
"${SYFT:-syft}" scan "$source" \
  --source-name "$name" \
  --source-version "$version" \
  -o "cyclonedx-json=$outdir/$name.cdx.json" \
  -o "spdx-json=$outdir/$name.spdx.json"

echo "SBOM を出力しました: $outdir/$name.cdx.json, $outdir/$name.spdx.json" >&2
