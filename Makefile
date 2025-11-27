.PHONY: install format test clean

VENV_NAME := venv
PYTHON := $(VENV_NAME)/bin/python
PIP := $(VENV_NAME)/bin/pip
BLACK := $(VENV_NAME)/bin/black

install:
	python3 -m venv $(VENV_NAME)
	$(PIP) install -r requirements.txt

format:
	$(BLACK) .

format-check:
	$(BLACK) --check .

test:
	$(PYTHON) tests/_tools/runner.py

clean:
	rm -rf $(VENV_NAME)
	find . -type d -name "__pycache__" -exec rm -rf {} +
