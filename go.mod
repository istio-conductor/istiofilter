module github.com/istio-conductor/istiofilter

go 1.15

replace github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

// Old version had no license
replace github.com/chzyer/logex => github.com/chzyer/logex v1.1.11-0.20170329064859-445be9e134b2

// Avoid pulling in incompatible libraries
replace github.com/docker/distribution => github.com/docker/distribution v2.7.1+incompatible

// Avoid pulling in kubernetes/kubernetes
replace github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.8.8-0.20200421182805-c3e488f0d815

// Client-go does not handle different versions of mergo due to some breaking changes - use the matching version
replace github.com/imdario/mergo => github.com/imdario/mergo v0.3.5

// See https://github.com/kubernetes/kubernetes/issues/92867, there is a bug in the library
replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190815234213-e83c0a1c26c8

require (
	github.com/alessio/shellescape v1.4.1
	github.com/gogo/protobuf v1.3.1
	github.com/pascaldekloe/name v1.0.1
	github.com/sirupsen/logrus v1.7.0
	github.com/slok/kubewebhook/v2 v2.0.0-beta.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	istio.io/api v0.0.0-20210106050357-a795af9d993e
	istio.io/istio v0.0.0-20210107030922-8e6386c46ffb
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	sigs.k8s.io/controller-runtime v0.7.0
)
