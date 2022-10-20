help:

lint:
	golangci-lint run --max-issues-per-linter 0 --max-same-issues 0

legal: .legal
.legal:
	addlicense -f ./templates/legal/header.txt -skip yaml -skip yml .
