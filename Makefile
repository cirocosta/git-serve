install:
	go install -v .

dist:
	ytt -f ./config | \
		kbld --images-annotation=false -f- > \
			./dist/release-no-auth.yaml
	ytt -f ./config --data-value enable_auth=true | \
		kbld --images-annotation=false -f- > \
			./dist/release-with-auth.yaml
.PHONY: dist
