build: build-git-serve build-git-serve-controller

build-%:
	mkdir -p dist
	CGO_ENABLED=0 go build \
		-trimpath -tags=osusergo,netgo,static_build \
		-o dist/$* ./cmd/$*


run: build-git-serve-controller
	./dist/git-serve-controller


install:
	go install -v ./cmd/git-serve


install-crds:
	kapp deploy -a git-serve-controller -f ./config/crd


k8s-release:
	mkdir -p dist
	kbld --images-annotation=false -f config > dist/git-serve.yaml


generate:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen crd rbac:roleName=role \
		paths=./pkg/apis/v1alpha1
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object \
		paths=./pkg/apis/v1alpha1


deploy: k8s-release
	kapp deploy -a git-serve -f dist


publish: build k8s-release
	./hack/publish.sh
