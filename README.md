# OOMIE

Maps `/var/log/kern.log` OOM messages to running Kubernetes pods. This is useful when the primary container forks multiple child processes, as well as mapping OOM's from init or sidecar containers.

## Installation

First compile and build the docker images

```
$ make container
```

Update manifests in `example.yaml` to use image and apply

```
$ kubectl apply -f example.yaml
```

```
$ kubectl get events -n demo-app
LAST SEEN   TYPE      REASON      OBJECT                         MESSAGE
30s         Warning   OOM         pod/demo-app-d944568f6-vnhk5   System OOM encountered, victim process: nginx, pid: 1270360, uid: 65534
```
