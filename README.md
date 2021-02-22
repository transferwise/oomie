# OOMIE

Maps `/var/log/kern.log` OOM messages to running Kubernetes pods. This is useful when the primary container forks multiple child processes, as well as mapping OOM's from side-cars.

## Installation

First compile and build the docker images

```
$ make container
```

Update manifests in `example.yaml` to use image and apply

```
$ kubectl apply -f example.yaml
```
