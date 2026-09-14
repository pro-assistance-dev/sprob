lint:
	./cmd/scripts/golangci.sh

update:
	./cmd/scripts/update_assister.sh

test:
	go test ./...

# Т8.1 (rdkb/TECH_DEBT.md): генератор TS-классов — тесты пакета codegen.
test_codegen:
	go test ./codegen/... -v
