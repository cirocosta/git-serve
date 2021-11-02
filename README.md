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

### cli

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

under [./dist](./dist) you'll find two reference Kubernetes manifests that can
be used for deploying `git-serve` to a Kubernetes cluster.


#### no auth

make use of the `release-no-auth.yaml` manifest.


```bash
kubectl create namespace git-serve

kapp deploy \
  -a git-serve \
  --into-ns git-serve \
  -f https://github.com/cirocosta/git-serve/releases/latest/download/release-no-auth.yaml
```
```console
Target cluster 'https://127.0.0.1:45085' (nodes: kind-control-plane)

Changes

Namespace   Name                  Kind
git-serve   git-serve             Deployment
^           git-serve             Service
^           git-serve             ServiceAccount
^           registry-credentials  Secret

Op:      4 create, 0 delete, 0 update, 0 noop
Wait to: 4 reconcile, 0 delete, 0 noop

Continue? [yN]: y

11:58:43AM: ---- applying 2 changes [0/4 done] ----
11:58:44AM: create secret/registry-credentials (v1) namespace: git-serve
11:58:44AM: create serviceaccount/git-serve (v1) namespace: git-serve
11:58:44AM: ---- waiting on 2 changes [0/4 done] ----
11:58:44AM: ok: reconcile serviceaccount/git-serve (v1) namespace: git-serve
11:58:44AM: ok: reconcile secret/registry-credentials (v1) namespace: git-serve
..
11:58:46AM: ---- waiting complete [4/4 done] ----

Succeeded
```


#### with auth

first, make sure that you have
[secretgen-controller](https://github.com/vmware-tanzu/carvel-secretgen-controller)
- it provides to use the ability of declaratively expressing our intention of
having a secret filled with a strong password and another secret with SSH keys,
and then once reconciled, it makes those available for us.

the `release-with-auth.yaml` manifest makes use of those kubernetes resources
provided by secretgen, so you must make sure you have it installed first:

```bash
SECRETGEN_CONTROLLER_VERSION=0.6.0

kapp deploy -a secretgen-controller \
  -f https://github.com/vmware-tanzu/carvel-secretgen-controller/releases/download/v$SECRETGEN_CONTROLLER_VERSION/release.yml
```

then install ours:

```bash
kubectl create namespace git-serve

kapp deploy \
  -a git-serve \
  --into-ns git-serve \
  -f https://github.com/cirocosta/git-serve/releases/latest/download/release-with-auth.yaml
```
```console
Namespace  Name                       Kind            Conds.  Age  Op      Op st.  Wait to    Rs  Ri
default    git-serve                  Deployment      -       -    create  -       reconcile  -   -
^          git-serve                  Service         -       -    create  -       reconcile  -   -
^          git-serve                  ServiceAccount  -       -    create  -       reconcile  -   -
^          git-serve-http-creds       Password        -       -    create  -       reconcile  -   -
^          git-serve-ssh-client-keys  SSHKey          -       -    create  -       reconcile  -   -
^          git-serve-ssh-server-keys  SSHKey          -       -    create  -       reconcile  -   -
^          registry-credentials       Secret          -       -    create  -       reconcile  -   -
```

once deployed, we can grab the credentials from the secrets instantiated:

```
kubectl get secret git-serve-ssh-client-keys \
  -o jsonpath={.data.ssh-privatekey} | \
  base64 --decode
```


## kubernetes custom resource

complete spec:

```yaml
#       a GIT server that makes use of every auth
#       feature that there is: for `http` and `ssh`.
#
apiVersion: utxo.com.br/v1alpha1
kind: GitServer
metadata:
  name: git-server
spec:
  http:
    auth:
      # completely disabling auth would permit
      # anyone to `git clone `<server>/<repo>.git`
      #
      disabled: false

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
      # disabling ssh auth means that _anyone_ can
      # pull/push via the SSH transport without
      # presenting either basic auth credentials
      # or a private key that has been previously
      # authorized.
      #
      # note.: known_hosts verification will still
      # be performed at the client side unless disabled
      # (e.g., via StrictHostKeyChecking=no option.)
      #
      disabled: false

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
  sshServerKnownHosts: |-
    git-serve ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGm...43JXiUFFAaQ==
    git-serve.namespace.svc.cluster.local AAAAB3NzaC1yc2EA...43JXiUFFAaQ==
```


## license

MIT
