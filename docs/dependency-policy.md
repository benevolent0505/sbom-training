# 依存バージョンの選定基準

## cooldown: 公開から 7 日

ライブラリやツールは、**公開から 7 日以上経過したバージョン** を採用します。

公開直後のバージョンには、乗っ取られたアカウントから配布された悪意あるリリースや、すぐに取り下げられる不具合が含まれることがあります。こうした問題の多くは数日以内に発見・削除されるため、一定期間待つことで巻き込まれる可能性を下げられます。

## 対象ごとの仕組み

| 対象 | 仕組み | 設定 |
| --- | --- | --- |
| Dependabot の更新 PR(全エコシステム) | `cooldown.default-days: 7` | [.github/dependabot.yml](../.github/dependabot.yml) |
| npm(手動の `npm install`) | `min-release-age=7`。npm が 7 日未満のバージョンを解決対象から外す | [web/.npmrc](../web/.npmrc) |
| Go モジュール(手動の `go get`) | Go に同等の機能がないため、CI でモジュールの公開日を検査する | [scripts/check-cooldown.sh](../scripts/check-cooldown.sh) |
| GitHub Actions、ベースイメージ、ローカル用 CLI | 手動で追加・更新するときに公開日を確認する。以降の更新は Dependabot に任せる | — |

`min-release-age` は npm 11.10 以降で有効です。それより古い npm では設定が無視される(警告が出る)ので、Node.js 24 の新しいパッチに同梱される npm を使うか、`npm install --before=<7 日前の日付>` を指定してください。

`npm ci` はロックファイルどおりにインストールするだけで、公開日は検査しません。ロックファイルに入る時点(`npm install` と Dependabot)で基準を満たすようにしています。

## 例外: 脆弱性の修正

**脆弱性の修正は cooldown を待ちません。** 既知の脆弱性を抱え続けるリスクの方が大きいためです。

- Dependabot の **セキュリティ更新** には cooldown が適用されません(cooldown は定期的なバージョン更新だけに効きます)。
- Go モジュールを手動で上げる場合は、[.cooldown-exempt](../.cooldown-exempt) に `モジュール@バージョン` を理由付きで追記すると CI の検査を通過できます。7 日を過ぎたら行を削除します。
- npm で手動で上げる場合は、`web/.npmrc` の `min-release-age-exclude` に対象パッケージを一時的に追加します。

## その他の方針

- **GitHub Actions はコミット SHA で固定する。** タグは後から書き換えられるためです。バージョンはコメントで併記します。
- **ベースイメージはタグと digest の両方で固定する。**
- **ロックファイル(`go.sum`、`package-lock.json`)を必ずコミットする。**
