ARG BUILDER_IMAGE=golang:alpine
ARG RUNTIME_IMAGE=alpine


FROM $BUILDER_IMAGE as builder

        WORKDIR /workspace

        COPY . .

        RUN set -x && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
                go build -a -v \
			-trimpath \
			-tags osusergo,netgo,static_build \
			-o git-serve \
			.


FROM $RUNTIME_IMAGE

	RUN apk add --update git
        COPY --from=builder /workspace/git-serve /usr/local/bin/git-serve

	RUN addgroup -g 1000 -S nonroot && \
	    adduser -u 1000 -S nonroot -G nonroot
	USER nonroot:nonroot
