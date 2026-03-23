SHELL := /usr/bin/env bash

BENCH_PATTERN ?= Benchmark
BENCH_COUNT ?= 10
RESULTS_DIR := results
RAW_DIR := $(RESULTS_DIR)/raw
COMPARE_DIR := $(RESULTS_DIR)/compare
DASH_DIR := $(RESULTS_DIR)/dashboards

LATEST_RAW := $(shell ls -1t $(RAW_DIR)/bench-*.txt 2>/dev/null | head -n 1)

.PHONY: dirs bench-go bench-go-smoke bench-compare bench-report bench-hyperfine bench-yaml-baseline clean

dirs:
	mkdir -p $(RAW_DIR) $(COMPARE_DIR) $(DASH_DIR)

bench-go: dirs
	BENCH_PATTERN=$(BENCH_PATTERN) BENCH_COUNT=$(BENCH_COUNT) ./run.sh

bench-go-smoke: dirs
	BENCH_PATTERN=$(BENCH_PATTERN) BENCH_COUNT=1 ./run.sh

bench-compare: dirs
	@if [ -z "$(BASE_FILE)" ] || [ -z "$(NEW_FILE)" ]; then \
	  echo "set BASE_FILE and NEW_FILE"; \
	  exit 1; \
	fi
	BASE_FILE=$(BASE_FILE) NEW_FILE=$(NEW_FILE) ./scripts/compare_benchmarks.sh

bench-report: dirs
	@if [ -z "$(INPUT)" ]; then \
	  echo "set INPUT=<benchmark output txt>"; \
	  exit 1; \
	fi
	python3 ./scripts/extract_go_bench.py "$(INPUT)" "$(DASH_DIR)/summary.json"

bench-hyperfine: dirs
	./scripts/run_hyperfine.sh

bench-yaml-baseline: dirs
	./scripts/capture_yaml_baseline.sh

clean:
	rm -rf $(RESULTS_DIR)
