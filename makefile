git_push_with_tag:
	git add .
	git commit -m "$m"
	git push
	git tag "$m"
	git push --tags
