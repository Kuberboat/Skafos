# Skafos

Skafos is a service mesh built on top of [Kuberboat](https://gitee.com/xx01cyx/kuberboat), a simplified implementation of Kubernetes.

## How to build

First, you should have Golang 1.18 installed. On MacOS, just run

```bash
brew install go@1.18
```

and set your `PATH` as 

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Now you are ready for building Skafos. Simply run

```bash
make
``` 

and you will see the executable under `out/bin`.
