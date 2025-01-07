lint:
	./cmd/scripts/golangci.sh

update:
	./cmd/scripts/update_assister.sh

test:
	go test ./...
