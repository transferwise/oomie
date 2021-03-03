module github.com/kgtw/oomie

go 1.15

// Override can be removed once a release is made that includes https://github.com/google/cadvisor/pull/2817
replace github.com/google/cadvisor => github.com/kgtw/cadvisor v0.38.1-0.20210301182203-dcbeba1642af

require (
	github.com/google/cadvisor v0.38.8
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/magefile/mage v1.11.0 // indirect
	github.com/sirupsen/logrus v1.8.0
	golang.org/x/sys v0.0.0-20210220050731-9a76102bfb43 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	k8s.io/klog/v2 v2.4.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
)
