# CHANGELOG
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

[v1.6](https://github.com/StarRocks/starrocks-kubernetes-operatorr/compare/v1..6..v1.5)
## Added
This version add v1 version for starrockscluster CRD. now v1alpha1 all function added in v1. in future v1alpha1 will verify the new functions, If the function is ok and client like will added to v1.  
[Feature] support deploy one more starrocks clusters in one namespace.
[Feature] can export be and cn to public when client needs.
[Feature] convert the fe deploy mode to parallel mode.[#73](https://github.com/StarRocks/starrocks-kubernetes-operator/issues/73) [#69](https://github.com/StarRocks/starrocks-kubernetes-operator/issues/69)
[Feature] support ability to restart all pods about fe, cn,be. when the configmap update and verified, the ability can restart all pods for effect.[#68](https://github.com/StarRocks/starrocks-kubernetes-operator/issues/68)

[v1.5](https://github.com/StarRocks/starrocks-kubernetes-operatorr/compare/v1.5...v1.4)
## Added
[Feature] support annotations for all components. [#48](https://github.com/StarRocks/starrocks-kubernetes-operator/issues/48)

[v1.4](https://github.com/StarRocks/starrocks-kubernetes-operatorr/compare/v1.4...v1.3)
## Added
[Feature] use EmptyDir for be/log. [#57](https://github.com/StarRocks/starrocks-kubernetes-operator/issues/57)
[Feature] support imagePullSecrets. [#44](https://github.com/StarRocks/starrocks-kubernetes-operator/issues/44)
[Feature] support load data from out k8s.[#62](https://github.com/StarRocks/starrocks-kubernetes-operator/issues/62)


[v1.3](https://github.com/StarRocks/starrocks-kubernetes-operatorr/compare/v1.3...v1.2)
## Added
support scheduling,Preemption and Eviction
## Changed

## Fixed
fix the common user access persistentVolume.

## Deleted

# [v1.2](https://github.com/StarRocks/starrocks-kubernetes-operatorr/compare/v1.2...v1.2)

## Added
* use statefulset and v2 autoscaling.

## Changed

## Deleted
