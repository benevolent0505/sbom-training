# ローカルで CI と同じ流れ(テスト → イメージビルド → SBOM 生成 → 脆弱性スキャン)を実行する。
# syft / grype は `aqua install` で導入できる(aqua.yaml)。

VERSION    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo local)
COMPONENTS := api web
SBOM_DIR   := sbom

export VERSION

.PHONY: help
help: ## このヘルプを表示する
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-16s %s\n", $$1, $$2}'

.PHONY: test
test: test-api test-web ## すべてのテストを実行する

.PHONY: test-api
test-api: ## API のテスト(go vet / go test)
	cd api && go vet ./... && go test ./...

.PHONY: test-web
test-web: ## Web のテスト(型チェック / vitest)
	cd web && npm ci && npm run typecheck && npm test

.PHONY: generate
generate: ## GraphQL スキーマから Go のコードを再生成する
	cd api && go generate ./...

.PHONY: check-cooldown
check-cooldown: ## Go モジュールが cooldown(公開から 7 日)を満たすか確認する
	./scripts/check-cooldown.sh

.PHONY: build
build: ## Docker イメージをビルドする
	docker compose build

.PHONY: sbom
sbom: build ## イメージから SBOM(CycloneDX / SPDX)を生成する
	@for c in $(COMPONENTS); do \
		./scripts/sbom-generate.sh docker:sbom-training/$$c:local sbom-training-$$c $(VERSION) $(SBOM_DIR) || exit 1; \
	done

.PHONY: scan
scan: ## 生成済みの SBOM を脆弱性スキャンする(基準は .grype.yaml)
	@status=0; for c in $(COMPONENTS); do \
		echo "== sbom-training-$$c =="; \
		./scripts/sbom-scan.sh $(SBOM_DIR)/sbom-training-$$c.cdx.json || status=$$?; \
	done; exit $$status

.PHONY: up
up: ## アプリケーションをローカルで起動する(http://localhost:3000)
	docker compose up --build -d

.PHONY: down
down: ## ローカルのアプリケーションを停止してデータを削除する
	docker compose down -v

.PHONY: clean
clean: ## SBOM とスキャン結果を削除する
	rm -rf $(SBOM_DIR)
