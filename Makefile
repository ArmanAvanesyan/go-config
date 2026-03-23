.PHONY: tidy verify lint test test-race test-cover test-cover-pkg test-cover-integration test-integration bench bench-mem bench-local bench-local-smoke bench-compare-local bench-report-local bench-hyperfine-local bench-yaml-baseline-local bench-readme-refresh-local wasm-build wasm-build-docker wasm-verify wasm-verify-docker check-wasm report-local report-pr-local coverage-target-report fuzz vet clean build fmt format check help

help:
	@echo "Targets:"
	@echo "  tidy              - go mod tidy"
	@echo "  verify            - go mod verify"
	@echo "  fmt / format      - go fmt ./..."
	@echo "  vet               - go vet ./..."
	@echo "  lint              - golangci-lint run"
	@echo "  test              - go test ./..."
	@echo "  test-race         - go test ./... -race -cover (requires CGO + gcc)"
	@echo "  test-cover        - unit coverage profile (excludes examples/*, rustjson/rusttoml wrappers, testutil)"
	@echo "  test-cover-integration - integration coverage profile (excludes examples/*, rustjson/rusttoml wrappers, testutil)"
	@echo "  test-integration  - go test -tags=integration ./..."
	@echo "  bench             - benchmarks (time only)"
	@echo "  bench-mem         - benchmarks with allocation stats"
	@echo "  bench-local       - run tooling benchmark suite"
	@echo "  bench-local-smoke - run tooling benchmark suite (count=1)"
	@echo "  bench-compare-local - compare two tooling benchmark outputs"
	@echo "  bench-report-local  - extract tooling benchmark dashboard JSON"
	@echo "  bench-hyperfine-local - run tooling hyperfine orchestration"
	@echo "  bench-yaml-baseline-local - capture YAML cold/warm baseline snapshots"
	@echo "  bench-readme-refresh-local - refresh README benchmark table from summary.json"
	@echo "  wasm-build         - build Rust/WASM artifacts and copy into extensions/wasm (host rustc)"
	@echo "  wasm-build-docker  - same as wasm-build inside rust:1.94-bookworm (matches CI)"
	@echo "  wasm-verify        - wasm-build + diff only the four checked-in .wasm binaries vs HEAD"
	@echo "  wasm-verify-docker - wasm-build-docker then git diff vs HEAD (matches CI)"
	@echo "  check-wasm         - alias for wasm-verify"
	@echo "  report-local      - build unified tooling report (json + markdown)"
	@echo "  report-pr-local   - build unified tooling PR comment markdown"
	@echo "  coverage-target-report - retry package coverage and emit per-file strict report"
	@echo "  fuzz              - run all fuzz targets for 60s each"
	@echo "  build             - go build ./..."
	@echo "  clean             - remove coverage and build artifacts; clean test cache"
	@echo "  check             - fmt + vet + lint + test-race (full pre-commit pass)"

tidy:
	go mod tidy

verify:
	go mod verify

fmt format:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

# Race + coverage for the whole module (used in CI)
test-race:
ifeq ($(OS),Windows_NT)
	@where gcc >NUL 2>&1 || (echo "test-race requires gcc on PATH (install MinGW-w64 and retry)." && exit 1)
	@set CGO_ENABLED=1&& go test ./... -race -cover
else
	@CGO_ENABLED=1 go test ./... -race -cover
endif

# Coverage profile for the whole module
test-cover test-cover-pkg:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$pkgs = go list ./... | Where-Object { $$_ -notmatch '/examples/' -and $$_ -notmatch '/testutil$$' -and $$_ -notmatch '/extensions/wasm/parser/rustjson$$' -and $$_ -notmatch '/extensions/wasm/parser/rusttoml$$' }; go test $$pkgs -cover -coverprofile=coverage.out"
else
	@pkgs="$$(go list ./... | grep -Ev '/examples/|/testutil$$|/extensions/wasm/parser/rustjson$$|/extensions/wasm/parser/rusttoml$$')"; \
	go test $$pkgs -cover -coverprofile=coverage.out
endif

# Coverage profile for integration-tagged tests/packages.
test-cover-integration:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$pkgs = go list ./... | Where-Object { $$_ -notmatch '/examples/' -and $$_ -notmatch '/testutil$$' -and $$_ -notmatch '/extensions/wasm/parser/rustjson$$' -and $$_ -notmatch '/extensions/wasm/parser/rusttoml$$' }; go test -tags=integration $$pkgs -cover -coverprofile=coverage.integration.out"
else
	@pkgs="$$(go list ./... | grep -Ev '/examples/|/testutil$$|/extensions/wasm/parser/rustjson$$|/extensions/wasm/parser/rusttoml$$')"; \
	go test -tags=integration $$pkgs -cover -coverprofile=coverage.integration.out
endif

# Integration tests (requires //go:build integration in test files)
test-integration:
	go test -tags=integration ./...

# Benchmarks — time only
bench:
	go test ./... -run=^$$ -bench=.

# Benchmarks — time + allocation stats
bench-mem:
	go test ./... -run=^$$ -bench=. -benchmem

# Tooling benchmarks (comparative suite in nested module)
bench-local:
	$(MAKE) -C tooling/benchmarks bench-go

bench-local-smoke:
	$(MAKE) -C tooling/benchmarks bench-go-smoke

bench-compare-local:
	$(MAKE) -C tooling/benchmarks bench-compare

bench-report-local:
	$(MAKE) -C tooling/benchmarks bench-report

bench-hyperfine-local:
	$(MAKE) -C tooling/benchmarks bench-hyperfine

bench-yaml-baseline-local:
	$(MAKE) -C tooling/benchmarks bench-yaml-baseline

bench-readme-refresh-local:
	python3 tooling/benchmarks/scripts/update_readme_benchmark_table.py

# Rust `make -C rust all` copies only these; verify drift on binaries only (not .go sources under the same trees).
WASM_VERIFY_PATHS := \
	extensions/wasm/parser/rusttoml/toml_parser.wasm \
	extensions/wasm/parser/rustyaml/yaml_parser.wasm \
	extensions/wasm/parser/rustjson/json_parser.wasm \
	extensions/wasm/validator/rustpolicy/policy.wasm

wasm-build:
	$(MAKE) -C rust all

# Rebuild with the same toolchain image as CI (host rustc often emits different .wasm bytes).
wasm-build-docker:
	docker run --rm \
		-v "$(CURDIR):/ws" \
		-w /ws \
		rust:1.94-bookworm \
		bash -lc 'export PATH="/usr/local/cargo/bin:$$PATH" && apt-get update -qq && apt-get install -y --no-install-recommends make ca-certificates >/dev/null && rustup target add wasm32-wasip1 && make wasm-build'

wasm-verify: wasm-build
	git diff --exit-code HEAD -- $(WASM_VERIFY_PATHS)

# Build in Docker, then git diff on the host — git inside rust:1.94-bookworm often does not see /ws as a repo.
wasm-verify-docker: wasm-build-docker
	git diff --exit-code HEAD -- $(WASM_VERIFY_PATHS)

check-wasm: wasm-verify

report-local:
	python3 tooling/reports/scripts/build_report.py
	python3 tooling/reports/scripts/render_markdown.py

report-pr-local:
	python3 tooling/reports/scripts/build_report.py
	python3 tooling/reports/scripts/render_pr_comment.py

coverage-target-report:
	python3 tooling/scripts/coverage_retry_report.py

# Fuzz all targets matching Fuzz* for 60s each
fuzz:
	@for target in $$(go test -list 'Fuzz.*' ./... 2>/dev/null | grep -E '^Fuzz'); do \
		echo "fuzzing $$target ..."; \
		go test -run=^$$ -fuzz=$$target -fuzztime=60s ./...; \
	done

build:
	go build ./...

# Artifacts to remove on clean
CLEAN_ARTIFACTS := coverage.out cov cov_http cov_http.out cov_httpclient cov.out cov2.out coverage coverage_errors coverage_providers e.out full.out full2.out httpclient_cov.out

clean:
ifeq ($(OS),Windows_NT)
	-del /q /f $(CLEAN_ARTIFACTS) 2>/dev/null
else
	-rm -f $(CLEAN_ARTIFACTS)
endif
	go clean -testcache

check: fmt vet lint test-race
