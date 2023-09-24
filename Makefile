define HELP_TEXT
Welcome to aetherfs!

Targets:
help		provides help text
legal		prepends legal header to source code

endef
export HELP_TEXT

help:
	@echo "$$HELP_TEXT"

legal: .legal
.legal:
	addlicense -f ./templates/legal/header.txt -skip yaml -skip yml .
