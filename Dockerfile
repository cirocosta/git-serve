ARG BUILDER_IMAGE=golang:alpine
ARG RUNTIME_IMAGE=alpine


FROM $BUILDER_IMAGE as builder

        WORKDIR /workspace

        COPY go.mod         go.mod
        COPY go.sum         go.sum
        COPY cmd        cmd
        COPY pkg         pkg


FROM builder AS git-serve

        RUN set -x && \
                CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on      \
                go build -a -v                                            \
                        -trimpath                                         \
                        -tags osusergo,netgo,static_build                 \
                        -o git-serve                                      \
                        ./cmd/git-serve


FROM builder AS git-serve-controller

        RUN set -x && \
                CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on      \
                go build -a -v                                            \
                        -trimpath                                         \
                        -tags osusergo,netgo,static_build                 \
                        -o git-serve                                      \
                        ./cmd/git-serve-controller


FROM $RUNTIME_IMAGE

        RUN set -ex && \
                apk add --no-cache --update git                         && \
                addgroup -g 1000 -S nonroot                             && \
                adduser -u 1000 -S nonroot -G nonroot

        USER nonroot:nonroot

        COPY --from=git-serve --chown=1000:1000 \
                /workspace/git-serve /usr/local/bin/git-serve
        COPY --from=git-serve-controller --chown=1000:1000 \
                /workspace/git-serve /usr/local/bin/git-serve-controller
