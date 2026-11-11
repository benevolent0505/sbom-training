# SBOM の運用

## 何を SBOM の対象にしているか

SBOM は **コンテナイメージ** から生成します。ソースコード(`go.mod` や `package-lock.json`)ではなく、実際にデプロイされるものを対象にするためです。

| | ソースから生成 | イメージから生成(このプロジェクト) |
| --- | --- | --- |
| Go | `go.mod` の全モジュール。コード生成ツール(gqlgen)の依存も含む | バイナリに実際にリンクされたモジュールと Go 標準ライブラリだけ |
| npm | devDependencies を含む全パッケージ | Next.js の standalone 出力に含まれる本番用パッケージだけ |
| OS | 含まれない | ベースイメージの OS パッケージを含む |

ソース側の依存(開発ツールを含む)は、GitHub が `go.mod` と `package-lock.json` を直接読んで Dependency Graph に載せるので、そちらで把握できます。

## SBOM はどこにあるか

| 場所 | 内容 | 用途 |
| --- | --- | --- |
| ワークフローの artifact(`sbom-api` / `sbom-web`) | CycloneDX、SPDX、スキャン結果、対象イメージの digest(`*.image.txt`) | 各ビルドの記録。再スキャンの入力。90 日で削除される |
| GitHub Release の添付ファイル | CycloneDX と SPDX | リリースごとの恒久的な記録、利用者への提供。artifact が消えた後の再スキャンの入力 |
| Dependency Graph | パッケージ一覧 | Dependabot alerts による通知、Web UI での検索 |
| Code scanning | main とタグのビルドのスキャン結果(SARIF) | Security タブでの脆弱性の一覧(public リポジトリのみ) |

SBOM とイメージの対応は、artifact 内の `*.image.txt`(`ghcr.io/…/api@sha256:…`)と、イメージのタグ `sha-<コミット SHA>` でたどれます。

main の最新ビルドの SBOM を手元に取得するには:

```bash
gh run download "$(gh run list --workflow ci.yml --branch main --event push --status completed --limit 1 --json databaseId --jq '.[0].databaseId')" --name sbom-api --dir sbom
```

### attestation を使っていない理由

SBOM をイメージの digest に署名付きで紐付ける方法(cosign や GitHub の artifact attestation)は、構成を小さく保つために採用していません。SBOM の真正性は「このリポジトリの CI が保存した artifact / Release である」ことに依拠しています。

イメージを外部に配布し、受け取る側が SBOM を検証できる必要がある場合は、attestation の追加を検討します。その際は次の点に注意します。

- 鍵を持たない keyless 署名では、リポジトリ名・ワークフローのパス・コミット SHA などが Sigstore の公開ログに恒久的に記録されます。private リポジトリでも同じなので、名前を外に出せないリポジトリでは鍵ペアによる署名を使います。
- GitHub の artifact attestation は、public リポジトリでは無料で使えますが、private リポジトリでは上位プランが必要です。

## 脆弱性の基準

基準は [.grype.yaml](../.grype.yaml) に書いてあり、ローカル・CI・再スキャンのすべてで同じものを使います。

- **High 以上で失敗**(`fail-on-severity: high`)
- **修正版があるものだけが対象**(`only-fixed: true`)

`only-fixed` を有効にしているのは、ベースイメージの OS パッケージに「修正予定なし」の指摘が常に残るためです。これらは自分たちでは直せず、ゲートに含めると常に失敗して形骸化します。対象外の指摘も `*.grype.json` には記録されます。

## 脆弱性が見つかったとき

検出される経路は 3 つあります。

1. **PR の CI が失敗する** — その PR で入れた変更(依存の追加やベースイメージの更新)が原因
2. **再スキャンが Issue を起票する** — ビルド後に公表された脆弱性。再スキャンは [sbom-rescan.yml](../.github/workflows/sbom-rescan.yml) を手動または定期で実行する
3. **Dependabot alerts が通知する** — 同上。修正版があればセキュリティ更新の PR も作られる

対応は次の順で検討します。

1. **更新する。** 修正版に上げる PR を作る(Dependabot のセキュリティ更新 PR があればそれを使う)。脆弱性修正は cooldown を待ちません([dependency-policy.md](dependency-policy.md))。main にマージされると新しいイメージと SBOM が作られ、次の再スキャンで Issue が自動的に閉じます。
2. **影響がないと判断できる場合は除外する。** 該当機能を使っていないなど、根拠を確認したうえで `.grype.yaml` の `ignore` に、理由・判断者・見直し時期をコメントで添えて追記します。除外は PR でレビューします。
3. **どちらもすぐにできない場合。** Issue に状況と期限を書き、緩和策(該当機能の無効化など)を検討します。基準そのものを緩めて通すことはしません。

## ツールのバージョン

| | バージョンの決まり方 |
| --- | --- |
| CI | [anchore/sbom-action](https://github.com/anchore/sbom-action) と [anchore/scan-action](https://github.com/anchore/scan-action) が同梱する既定の Syft / Grype。Action を Dependabot が更新すると追随する |
| ローカル | [aqua.yaml](../aqua.yaml) で固定。手動で更新する |

脆弱性データベースは実行のたびに最新が取得されるため、同じ SBOM でも日によって結果が変わります(これが再スキャンの意味です)。

## このサンプルで扱っていないこと

- **ライセンスの検査。** SBOM にはライセンス情報が含まれますが、ポリシー判定はしていません。
- **VEX。** 「影響なし」の判断は `.grype.yaml` の `ignore` で管理しています。判断を SBOM の利用者にも伝える必要がある場合は、OpenVEX 文書を作り `.grype.yaml` の `vex-documents` で読み込ませる方法があります。
- **到達可能性の解析。** Grype は「そのバージョンのパッケージが含まれるか」で判定し、脆弱な関数を実際に呼んでいるかは見ません。
