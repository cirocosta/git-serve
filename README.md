# git-serve

a toy git server with that lets you clone/push any repositories you want.


- [usage](#usage)
  - [cli](#cli)
    - [no auth](#no-auth)
    - [with auth](#with-auth)
  - [kubernetes](#kubernetes)
    - [no auth](#no-auth-1)
    - [with auth](#with-auth-1)
- [license](#license)


## usage

`git-serve` can be used either as a CLI (like, `git serve [args]`), or as an
extension to kubernetes that lets you provision git servers inside a cluster.


### cli

```console
$ git-serve --help

Usage of git-serve:
  -data-dir string
        directory where repositories will be stored (default "/tmp/git-serve")
  -git string
        absolute path to git executable (default "/usr/bin/git")
  -http-bind-addr string
        address to bind the http server to (default ":8080")
  -http-no-auth
        disable default use of basic auth for http
  -http-password string
        password (default "admin")
  -http-username string
        username (default "admin")
  -ssh-authorized-keys string
        path to public keys to authorized (ssh format)
  -ssh-bind-addr string
        address to bind the ssh server to (default ":2222")
  -ssh-host-key string
        path to private key to use for the ssh server
  -ssh-no-auth
        disable default use of public key auth for ssh
  -v    turn verbose logs on/off
```

ps.: any of the flags above can be set via environment variables prefixed with
`GIT_SERVE_`, for instance, to se `-http-password`, use
`GIT_SERVE_HTTP_PASSWORD`.


#### no auth

by default, `git-serve` expects to serve repositories over SSH and HTTP with
auth. to disable that, make use of the `*-no-auth` flags, for instance:

```bash
git-serve -http-no-auth -ssh-no-auth
```

which would let you clone/push/pull with no need for providing any credentials
at all:

```
git clone ssh://localhost:2222/foo.git .
git clone http://localhost:2222/foo.git .
```

note: by default (i.e., unless overwritten by `-ssh-host-key`), the SSH
server's public key that is used has the following fingerprint:

```
SHA256:PJo73EJabnFPeCNm1vssGMLsJSv7I9LztZrTwQOIdb8.
```


#### with auth

authn and authz is configured independently via transport-specific flags:

- for http: `-http-username` and `-http-password` configure, correspondingly,
  the username and password that must be provided via basic auth

- for ssh: `-ssh-authorized-keys` configured the set of client public keys that
  the server authorizes. note that you can also configure the server's keys via
  `-ssh-host-key`.


example:

1. generate a strong password for http's basic auth:

```console
$ gpg --gen-random --armor 0 24 | tee password.txt
lJnhFm7EKVYEOovPbq7+x2J5DKeQr6u7
```

2. generate both server and client SSH keys

```bash
for who in server client; do
  ssh-keygen -b 4096 -t rsa -f $who -q -N "" -C gitserve
done
```

to check out the fingerprint of the server pub key that has been generated
(will be added to your `~/.ssh/known_hosts` in the first connection attempt):

```console
$ ssh-keygen -lf ./server.pub
4096 SHA256:y8DKXGUYdySAFGnRzbPUmFaCLbDbWOa10ieOdkF4aZg gitserve (RSA)
```


3. start git-serve pointing at those


```console
git-serve \
  -http-username=admin \
  -http-password=$(cat password.txt) \
  -ssh-host-key ./server \
  -ssh-authorized-keys ./client.pub
```

ps.: by default, ports are: `http=8080,ssh=2222`.


### kubernetes

`git-serve` can also be used as an extension to kubernetes to provision servers
on-demand.

to install the custom resource definition:

```
kubectl apply -f https://github.com/cirocosta/git-serve/releases/latest/download/release.yaml
```

once installed, you should have a new kubernetes kind: GitServer.

```console
$ kubectl explain gitserver
KIND:     GitServer
VERSION:  ops.tips/v1alpha1
...
```


#### spec

```yaml
# a GIT server that makes use of every auth
# feature that there is: for `http` and `ssh`.
#
apiVersion: ops.tips/v1alpha1
kind: GitServer
metadata:
  name: git-server
spec:
  # image to base the pods of
  #
  image: cirocosta/git-serve

  http:
    auth:
      # grab username from a specific field in a
      # secret.
      #
      username:
        valueFrom:
          secretKeyRef:
            name: secret
            key: username

      # grab password from a specific field in a
      # secret.
      #
      password:
        valueFrom:
          secretKeyRef:
            name: secret
            key: password
  ssh:
    auth:
      # grab clients pub keys from a specific
      # field in a secret.
      #
      authorizedKeys:
        valueFrom:
          secretKeyRef:
            name: secret
            key: ssh-authorizedkeys

      # grab the server's private key from a
      # specific field in a secret.
      #
      hostKey:
        valueFrom:
          secretKeyRef:
            name: secret
            key: ssh-privatekey
status:
  observedGeneration: <int>
  conditions:
    - type: Ready
      status: True
```


## license

MIT
