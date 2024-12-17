include ./commons-test.mk

# ALL_MODULES includes ./* dirs (excludes . dir)
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort | grep -E '^./' )

# Append root module to all modules
GOMODULES = $(ALL_MODULES) $(PWD)

# Define a delegation target for each module
.PHONY: $(GOMODULES)
$(GOMODULES):
	@echo "Running target '$(TARGET)' in module '$@'"
	$(MAKE) -C $@ $(TARGET)

# Triggers each module's delegation target
.PHONY: for-all-target
for-all-target: $(GOMODULES)


.PHONY: lint-all
lint-all:
	@$(MAKE) for-all-target TARGET="lint"

.PHONY: test-all
test-all: 
	@$(MAKE) for-all-target TARGET="test-unit"

.PHONY: test-examples
test-examples:
	@echo "Running example tests..."
	$(MAKE) -C examples test

.PHONY: tidy-all
tidy-all:
	@$(MAKE) for-all-target TARGET="tidy"

## --------------------------------------

TCENV=tcvenv
PYTHONBIN=./$(TCENV)/bin

tcvenv: tcvenv/touchfile

tcvenv/touchfile:
	@echo "Creating docs $(TCENV)..."
	test -d $(TCENV) || python3 -m venv $(TCENV)
	@echo "Installing requirements..."
	. $(PYTHONBIN)/activate; pip install -Ur requirements.txt
	touch $(TCENV)/touchfile

clean-docs:
	@echo "Destroying docs $(TCENV)..."
	rm -rf $(TCENV)

.PHONY: serve-docs
serve-docs: tcvenv
	. $(PYTHONBIN)/activate; $(PYTHONBIN)/mkdocs serve
