# Copyright 2025 Acuvity, Inc.
# SPDX-License-Identifier: Apache-2.0
SHELL := bash
.SHELLFLAGS := -e -c -o pipefail
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MKFILE_DIR := $(shell echo $(dir $(abspath $(lastword $(MAKEFILE_LIST)))) | sed 's:/$$::')

OTELCOL_BIN := $(MKFILE_DIR)/otelcol-acuvity/otelcol-acuvity

OTELCOL_DEPS := $(MKFILE_DIR)/go.mod
OTELCOL_DEPS += $(MKFILE_DIR)/go.sum
OTELCOL_DEPS += $(MKFILE_DIR)/builder-config.yaml
OTELCOL_DEPS += $(MKFILE_DIR)/metadata.yaml
OTELCOL_DEPS += $(shell find $(MKFILE_DIR) -type d -name otelcol-acuvity -prune -o -type f -name '*.go' -print)

BUILDER_CONFIG ?= $(MKFILE_DIR)/builder-config.yaml
OTELCOL_CONFIG ?= $(MKFILE_DIR)/otelcol-config.yaml

# adjust to whatever endpont you are using for development in your otelcol-config.yaml
GENERATE_TRACES_OTLP_ENDPOINT ?= 127.0.0.1:4317

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


build: $(OTELCOL_BIN) ## Uses ocb to generate and build the OpenTelemetry Collector

$(OTELCOL_BIN): $(OTELCOL_DEPS)
	ocb --config=$(BUILDER_CONFIG)

.PHONY: clean
clean: ## Cleans the build
	-rm -rvf $(MKFILE_DIR)/otelcol-acuvity/

.PHONY: run
run: ## Runs the OpenTelemetry Collector
	$(MKFILE_DIR)/otelcol-acuvity/otelcol-acuvity --config $(OTELCOL_CONFIG)

.PHONY: generate-traces
generate-traces: ## Generates traces for testing
	telemetrygen traces --otlp-insecure --otlp-endpoint $(GENERATE_TRACES_OTLP_ENDPOINT)
