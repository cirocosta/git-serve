# git-serve

an http git server with no auth that lets you clone & push to repositories
on-demand.

(pretty much a thin wrapper around github.com/nulab/go-git-http-xfer)

## usage

```console
$ git-serve --help

USAGE
  git-serve [<arg> ...]

FLAGS
  -addr :8080          address to bind the server to
  -directory /tmp/git  where git repositories should be stored
  -git /usr/bin/git    absolute path to git executable
  -verbose=false       turn verbose logs on/off
```

### locally

first, install it:

```bash
```

then

```bash
# start the server storing repositories at `/tmp/git-serve`.
#
git serve -directory=/tmp/git-serve


# clone an empty repository
#
git clone http://localhost:8080/foo


# get inside the repository and write something to a file
#
cd foo
echo "foo" > ./foo


# commit and push
#
git add --all . && git commit -m "foo" && git push origin HEAD
```


### kubernetes

1. run `git-serve` in the cluster as a deployment (fronted by a service)


```bash
kubectl create namespace git-serve

kapp deploy \
  -a git-serve \
  --into-ns git-serve \
  -f https://github.com/cirocosta/git-serve/releases/latest/download/release.yaml
```
```console
Target cluster 'https://127.0.0.1:45085' (nodes: kind-control-plane)

Changes

Namespace  Name                  Kind          
git-serve  git-serve            Deployment     
^          git-serve            Service        
^          git-serve            ServiceAccount 
^          registry-credentials  Secret        

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


2. make use of it

for instance, running a job `commit-and-push` from the default namespace, we
can target the server running in the `git-serve` namespace from within a pod
using plain `git clone` as

```
git clone 
```

```bash
kubectl apply -f <(echo '---
apiVersion: batch/v1
kind: Job
metadata:
  name: commit-and-push
spec:
  backoffLimit: 1
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: run
        image: golang
        command:
          - /bin/bash
          - -cex
          - |
            cd `mktemp -d`

            git clone http://git-serve.git-serve/foo.git .

            git config user.email "hello@example.com"
            git config user.name "hello"

            echo "foo" > ./foo
            git add --all .&& git commit -m "foo" && git push origin HEAD
')
```

## LICENSE

MIT
