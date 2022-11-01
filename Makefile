.PHONY: help

# Show this help prompt
help:
	@echo '  Usage:'
	@echo ''
	@echo '    make <target>'
	@echo ''
	@echo '  Targets:'
	@echo ''
	@awk '/^#/{ comment = substr($$0,3) } comment && /^[a-zA-Z][a-zA-Z0-9_-]+ ?:/{ print "   ", $$1, comment }' $(MAKEFILE_LIST) | column -t -s ':' | grep -v 'IGNORE' | sort | uniq

# Scan for vulnerabilities
nancy: deps
	@go list -mod=readonly -json -m all | nancy sleuth

# Dependencies
deps:
	@go mod tidy -compat=1.17
	@go mod download
	@go mod vendor