install:
	go install -v .

dist:
	kbld --images-annotation=false -f ./config  > ./dist/release.yaml
.PHONY: dist
