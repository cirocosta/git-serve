#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly ROOT=$(cd "$(dirname $0)/.." && pwd)
readonly GIT_SERVE_SSH_PORT=${GIT_SERVE_SSH_PORT:-$($ROOT/tests/available-port.py)}
readonly GIT_SERVE_HTTP_PORT=${GIT_SERVE_HTTP_PORT:-$($ROOT/tests/available-port.py)}
readonly GIT_SERVE_DATA_DIR=${GIT_SERVE_DATA_DIR:-$(mktemp -d)}

main() {
        show_vars

        case $1 in
        no-auth) test_no_auth ;;

        auth) test_with_auth ;;

        *)
                echo "usage: $0 (auth|no-auth)"
                exit 1
                ;;

        esac
}

show_vars() {
        echo "vars:

	ROOT			$ROOT
	GIT_SERVE_SSH_PORT	$GIT_SERVE_SSH_PORT
	GIT_SERVE_HTTP_PORT	$GIT_SERVE_HTTP_PORT
	GIT_SERVE_DATA_DIR 	$GIT_SERVE_DATA_DIR
	"
}

test_with_auth() {
        local expected_revision
        local ssh_config_file
        local netrc_dir

        _log "test with auth"

        _start_server \
                -ssh-host-key=$ROOT/tests/testdata/server \
                -ssh-authorized-keys=$ROOT/tests/testdata/client.pub \
                -http-username=admin \
                -http-password=admin

        ssh_config_file=$(_prepare_ssh_config_file $GIT_SERVE_SSH_PORT)
        netrc_dir=$(_prepare_netrc_dir)

        export GIT_SSH_COMMAND="ssh -F $ssh_config_file"
        export HOME=$netrc_dir
        perform_basic_test

        _log "	>> succeeded!"
}

test_no_auth() {
        _log "test no auth"

        _start_server -ssh-no-auth

        export GIT_SSH_COMMAND="ssh -o StrictHostKeyChecking=no -p $GIT_SERVE_SSH_PORT"
        perform_basic_test

        _log "	>> succeeded!"
}

perform_basic_test() {
        local expected_revision

        {
                pushd $(mktemp -d)
                git clone ssh://localhost/foo.git .
                expected_revision=$(_make_deterministic_commit)
                git push origin HEAD
                popd
        }

        {
                pushd $(mktemp -d)
                git clone http://localhost:$GIT_SERVE_HTTP_PORT/foo.git .
                test $(git rev-parse HEAD) == $expected_revision && {
                        popd
                        return
                }

                echo "failed."
                exit 1
        }
}

_prepare_ssh_config_file() {
        local port=$1
        local fpath=$(mktemp)

        echo "Host localhost
	Port $port
	HostKeyAlias "[localhost]:2222"
	UserKnownHostsFile $ROOT/tests/testdata/known_hosts
	IdentityFile $ROOT/tests/testdata/client
	" >$fpath

        echo $fpath

}

_prepare_netrc_dir() {
        local dir=$(mktemp -d)

        printf "machine localhost\nlogin admin\npassword admin" >$dir/.netrc
        echo $dir
}

_make_deterministic_commit() {
        local revision=9031bbabfdfdbfb73d0d3bbbd8d2a894b0b5755d

        echo "foo" >README.md
        git add README.md

        env \
                GIT_AUTHOR_NAME=name \
                GIT_AUTHOR_EMAIL=email \
                GIT_AUTHOR_DATE='Fri Oct 31 00:00 2008' \
                GIT_COMMITTER_NAME=name \
                GIT_COMMITTER_EMAIL=email \
                GIT_COMMITTER_DATE='Fri Oct 31 00:00 2008' \
                git commit -q -m "first commit"

        printf "$revision"
}

_start_server() {
        git-serve \
                -http-bind-addr=:$GIT_SERVE_HTTP_PORT \
                -ssh-bind-addr=:$GIT_SERVE_SSH_PORT \
                -data-dir=$GIT_SERVE_DATA_DIR \
                $@ &>$GIT_SERVE_DATA_DIR/log.txt &

        trap "kill $!" EXIT

        sleep 1
}

_log() {
        printf '\n\t\033[1m%s\033[0m\n\n' "$1" 1>&2
}

main "$@"
