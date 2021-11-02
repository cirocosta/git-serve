build:
	mkdir -p dist
	go build -o dist/git-serve -v ./cmd/git-serve
	go build -o dist/git-serve-controller -v ./cmd/git-serve-controller


run-controller: build
	./dist/git-serve-controller


install-crds:
	kapp deploy -a git-serve-controller -f ./config/crd


generate:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen crd rbac:roleName=role \
		paths=./pkg/apis/v1alpha1
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object \
		paths=./pkg/apis/v1alpha1
