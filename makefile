git_push_with_tag: lint
	git add .
	git commit -m "$m"
	git push
	git tag "$m"
	git push --tags

lint:
	./cmd/scripts/golangci.sh
