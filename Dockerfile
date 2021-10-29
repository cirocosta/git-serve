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

	RUN apk add --update git openssh-server bash
        COPY --from=builder /workspace/git-serve /usr/local/bin/git-serve

	RUN addgroup -g 1000 -S nonroot && \
	    adduser -u 1000 -S nonroot -G nonroot
	USER nonroot:nonroot

	# COPY --chown=1000:1000 ./ssh/sshd_config  /etc/ssh/sshd_config
	COPY --chown=1000:1000 ./ssh/ssh_host_key /etc/ssh/ssh_host_key
