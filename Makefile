build:
	mkdir -p dist
	go build -o dist/git-serve -v ./cmd/git-serve
	go build -o dist/git-serve-controller -v ./cmd/git-serve-controller


run: build
	./dist/git-serve-controller


install-crds:
	kapp deploy -a git-serve-controller -f ./config/crd


release:
	mkdir -p dist
	kbld --images-annotation=false -f config > dist/release.yaml


generate:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen crd rbac:roleName=role \
		paths=./pkg/apis/v1alpha1
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object \
		paths=./pkg/apis/v1alpha1


deploy:
	kapp deploy -a git-serve -f dist/release.yaml
