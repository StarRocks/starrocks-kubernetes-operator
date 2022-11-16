背景
当前kubernetes已成为云上调度系统的事实标准，k8s社区也成为仅次于linux的全球第二大生态圈。k8s的发展推动了各种软件和系统云原生化，用户可以得到软件和系统提供商更加系统和省心的服务。starrocks作为新一代数据湖仓系统需要给用户提供基于各种云之上的服务体系，需要构建一套能够在各种云服务之上为用户提供数据湖仓服务的数据服务体系。
目前starrocks还没有基于kubernetes提供服务化的方案，使用原生kubernetes资源来部署starrocks集群在使用层面会有诸多不便也无法进行自动化运维，用户无法从统一或者终态的视角掌握自己的集群从而不用关心具体的部署和运维实现。
使用原生kubernetes资源进行starrocks服务化部署会存在一下不便：
1. 原生资源部署整个集群对每个组件都需要一个手动组织一个yaml，这些yaml都需要用户管理维护，部署成本和维护成本相比于手动虚机部署并没有降低；
2. 编写yaml需要对于kubernetes多种原生资源的能力具有很深的了解，编写yaml一个完整可用的集群yaml较为复杂 ，如果用户有多个集群需要用户自己管理不同资源之间的组合关系，维护成本高；
3. 对于一些特性化的需求例如：服务探活不仅仅依赖端口探测，部署各个组件需要特定规则的亲和性，以及特定规则的弹性调度例如外部服务主动发起的倍数弹性扩缩容等等这些无法使用原生资源完成。基于以上原因对于starrocks来说需要开发一套属于自己的opertor。
   设计目标
   对于构建starrocks operator是一个随着starrocks功能和性能不断提升不断迭代升级更新的过程，对于starrocks operator的最终目标是不断提高starrocks 数据湖仓服务的可调度性、可用性和系统容错性。目前starrocks数据湖仓还没有一套完整的可部署方案，基于目前用户需求和定位定义阶段性目标如下：
   一期
   一期目标可总结为完成整个starrocks服务的容器化以及单个CRD资源自动化部署整个集群两个目标。
   一期主要包含以下能力：
- 部署
- 通用运维工作
    - 服务自动重启
    - 上下线扩缩容
    - 升级
      二期
      未来二期目标,从当前需求来看主要有以下几部分：
1. 实现服务pod的自管能力为服务的自定义探活以及自定义调度和部署奠定基础。
2. 实现倍数扩缩容以及自定义弹性自动扩缩容能力。
3. fe和be失败重启后特性化重新加入集群的能力。(fe的重启后重新加入集群存在的问题下面有详解)
   当前一二期都有明确需要解决的问题和实现，但是对于构建一套完备湖仓服务体系并不足够，在三期或者说未来我们需要在以下几个层面演进和优化设计：1. 基于kubernetes服务化的监控服务体系。2. 更加精细化服务支持比如：自定义探活，数据的远程备份和还原，大数据处理服务化等。3. 基于kubernetes的serverless服务运维平台构建。
   用户使用资源
   对于用户部署使用整个starrocks服务只需要感知一种资源就是starrocksCluster资源，以下为部分资源CRD信息展示：
````yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.9.0
  creationTimestamp: null
  name: starrocksclusters.starrocks.com.starrocks.com
spec:
  group: starrocks.com.starrocks.com
  names:
    kind: StarRocksCluster
    listKind: StarRocksClusterList
    plural: starrocksclusters
    singular: starrockscluster
  scope: Namespaced
  versions:
    - name: v1alpha1
      schema:
        openAPIV3Schema:
          description: StarRocksCluster is the Schema for the starrocksclusters API
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: StarRocksClusterSpec defines the desired state of StarRocksCluster
              properties:
                beSpec:
                  properties:
                    image:
                      type: string
                    limits:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Limits describes the maximum amount of compute resources
                      allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                    replicas:
                      format: int32
                      type: integer
                    requests:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Requests describes the minimum amount of compute
                      resources required. If Requests is omitted for a container,
                      it defaults to Limits if that is explicitly specified, otherwise
                      to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                    service:
                      properties:
                        annotations:
                          additionalProperties:
                            type: string
                          description: Additional annotations for the service
                          type: object
                        clusterIP:
                          description: ClusterIP is the clusterIP of service
                          type: string
                        labels:
                          additionalProperties:
                            type: string
                          description: Additional labels for the service
                          type: object
                        loadBalancerIP:
                          description: 'LoadBalancerIP is the loadBalancerIP of service
                          Optional: Defaults to omitted'
                          type: string
                        loadBalancerSourceRanges:
                          description: 'LoadBalancerSourceRanges is the loadBalancerSourceRanges
                          of service If specified and supported by the platform, this
                          will restrict traffic through the cloud-provider load-balancer
                          will be restricted to the specified client IPs. This field
                          will be ignored if the cloud-provider does not support the
                          feature." More info: https://kubernetes.io/docs/concepts/services-networking/service/#aws-nlb-support
                          Optional: Defaults to omitted'
                          items:
                            type: string
                          type: array
                        port:
                          description: "The port that will be exposed by this service.
                          \n NOTE: only used for TiDB"
                          format: int32
                          maximum: 65535
                          minimum: 1
                          type: integer
                        portName:
                          description: PortName is the name of service port
                          type: string
                        type:
                          description: Type of the real kubernetes service
                          type: string
                      type: object
                    storageVolumes:
                      items:
                        properties:
                          mountPath:
                            type: string
                          name:
                            type: string
                          storageClassName:
                            type: string
                          storageSize:
                            type: string
                        required:
                          - name
                          - storageSize
                        type: object
                      type: array
                  required:
                    - image
                    - replicas
                  type: object
                cnSpec:
                  properties:
                    image:
                      type: string
                    limits:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Limits describes the maximum amount of compute resources
                      allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                    replicas:
                      format: int32
                      type: integer
                    requests:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Requests describes the minimum amount of compute
                      resources required. If Requests is omitted for a container,
                      it defaults to Limits if that is explicitly specified, otherwise
                      to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                    service:
                      properties:
                        annotations:
                          additionalProperties:
                            type: string
                          description: Additional annotations for the service
                          type: object
                        clusterIP:
                          description: ClusterIP is the clusterIP of service
                          type: string
                        labels:
                          additionalProperties:
                            type: string
                          description: Additional labels for the service
                          type: object
                        loadBalancerIP:
                          description: 'LoadBalancerIP is the loadBalancerIP of service
                          Optional: Defaults to omitted'
                          type: string
                        loadBalancerSourceRanges:
                          description: 'LoadBalancerSourceRanges is the loadBalancerSourceRanges
                          of service If specified and supported by the platform, this
                          will restrict traffic through the cloud-provider load-balancer
                          will be restricted to the specified client IPs. This field
                          will be ignored if the cloud-provider does not support the
                          feature." More info: https://kubernetes.io/docs/concepts/services-networking/service/#aws-nlb-support
                          Optional: Defaults to omitted'
                          items:
                            type: string
                          type: array
                        port:
                          description: "The port that will be exposed by this service.
                          \n NOTE: only used for TiDB"
                          format: int32
                          maximum: 65535
                          minimum: 1
                          type: integer
                        portName:
                          description: PortName is the name of service port
                          type: string
                        type:
                          description: Type of the real kubernetes service
                          type: string
                      type: object
                  required:
                    - image
                    - replicas
                  type: object
                feSpec:
                  description: 'INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
                  Important: Run "make" to regenerate code after modifying this file'
                  properties:
                    image:
                      type: string
                    limits:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Limits describes the maximum amount of compute resources
                      allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                    replicas:
                      format: int32
                      type: integer
                    requests:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Requests describes the minimum amount of compute
                      resources required. If Requests is omitted for a container,
                      it defaults to Limits if that is explicitly specified, otherwise
                      to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                    service:
                      properties:
                        annotations:
                          additionalProperties:
                            type: string
                          description: Additional annotations for the service
                          type: object
                        clusterIP:
                          description: ClusterIP is the clusterIP of service
                          type: string
                        labels:
                          additionalProperties:
                            type: string
                          description: Additional labels for the service
                          type: object
                        loadBalancerIP:
                          description: 'LoadBalancerIP is the loadBalancerIP of service
                          Optional: Defaults to omitted'
                          type: string
                        loadBalancerSourceRanges:
                          description: 'LoadBalancerSourceRanges is the loadBalancerSourceRanges
                          of service If specified and supported by the platform, this
                          will restrict traffic through the cloud-provider load-balancer
                          will be restricted to the specified client IPs. This field
                          will be ignored if the cloud-provider does not support the
                          feature." More info: https://kubernetes.io/docs/concepts/services-networking/service/#aws-nlb-support
                          Optional: Defaults to omitted'
                          items:
                            type: string
                          type: array
                        port:
                          description: "The port that will be exposed by this service.
                          \n NOTE: only used for TiDB"
                          format: int32
                          maximum: 65535
                          minimum: 1
                          type: integer
                        portName:
                          description: PortName is the name of service port
                          type: string
                        type:
                          description: Type of the real kubernetes service
                          type: string
                      type: object
                    storageVolumes:
                      items:
                        properties:
                          mountPath:
                            type: string
                          name:
                            type: string
                          storageClassName:
                            type: string
                          storageSize:
                            type: string
                        required:
                          - name
                          - storageSize
                        type: object
                      type: array
                  required:
                    - image
                    - replicas
                  type: object
              required:
                - beSpec
                - cnSpec
                - feSpec
              type: object
            status:
              description: StarRocksClusterStatus defines the observed state of StarRocksCluster
              type: object
          required:
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
````
以上为设计实现的CRD部分内容展示，CRD的内容大多数部分更多是满足用户个性化需求设计，对于大多数用户来说需要更多的是基础能力，对于大多数用户来说能够满足基本能力只需要感知一些基本参数，其他大多数采用默认的方式部署就行。对于通用场景最简部署资源模型如下：
```yaml
apiVersion: starrocks.com/v1alpha1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
spec:
  serviceAccount: starrocksAccount
  starRocksFeSpec:
    replicas: 3
    image: starrocks.com/fe:2.40
    storageVolumes:
      - name: fe_storage
        storageClassName: fe_storage
        mountPath: /data/fe/meta
    requests:
      cpu: 4
      mem: 16Gi
  starRocksBeSpec:
    replicas: 6
    image: starrocks.com/be:2.40
    storageVolumes:
      - name: be_storage
        storageClassName: be_storage
        mountPath: /data/be
    requests:
      cpu: 4
      mem: 16Gi
  StarRocksCnSpec:
    replicas: 3
    image: starrocks.com/cn:2.40
    requests:
      cpu: 4
      mem: 16Gi
```

## 技术实现
利用容器化docker技术实现starrocks各个组件容器化。
基于starrocks支持fqdn模式通信能力和pod环境变量映射能力构造kubernetes环境下服务启动个样化脚本。pod环境变量映射能够将podyaml中信息映射为容器环境变量。
使用kubernetes的CRD扩展kubernetes api的能力，设计starrocksCluster资源实现对应的operator完成各个服务组件的自动化部署：使用statefulset部署fe，be，使用deployment部署cn，使用statefulset部署be。statefulset是kubernetes提供的有状态服务部署运维资源，deploment是kubernetes提供的无状态服务部署运维资源。
使用service实现各个服务之间使用fqdn相互访问和通信的能力。service是kubernetes提供的负载均衡资源，除提供负载均衡之外能够使用容器名称作为内部域名在kubernetes内部供其他服务访问。
使用kubernetes 的configMap的可映射成镜像中对应文件，实现配置文件与镜像分离提供用户在云原生场景下配置文件自定义定制能力。
一期的主要目标实现各个服务组件的容器化以及实现单个资源部署整个starrocks集群。需要涉及到以下工作:
1. starrocks 整体具有fqdn模式通信的模式，目前2.4版本中已经具备需要额外配置启动信息。
2. 各个组件容器化改造。
3. 容器化服务启动脚本的云原生化改造，使用kubernetes的投射卷机制以及环境变量映射机制，利用starrocks的fqdn相互访问模式进行fe，be，cn组件个性化启动设计。
4. 设计starrocksCluster资源CRD如上图用户使用资源，实现各个组件基于原生资源自动化部署和上下线运维。
5. 基于kubernetes的configMap投射卷能力实现服务配置与镜像分离。
   设计折衷
   一期使用starrocksCluster资源将用户和底层实现隔离使用户不用关心具体部署和运维逻辑只关注自己想要的终态形式。底层使用原生资源部署和管理starrocks各个组件实例pod，在使用上和灵活性上不足，不能进行特性化设计，如：除端口探测方式探活服务外无法使用其他的方式去做活性探测；对于fe集群长时间运行后，某个节点重启后因为选择加入节点的generation较低无法实现自动加入，需要人工获取最新master节点重新设置启动参数。在使用原生资源的情况下，无法解决上述问题。从以上两例可看出使用原生资源在灵活性和可定制化设计上会存在较大的阻碍。
   自管pod模式进行定制化实现部署复杂度高需要实现原生资源的诸多能力，使用场景上存在诸多不确定性对于原生资源的诸多能力是否需要自管pod中支持存在不确定性。目前最迫切的需求更多是从无到有将整个starrocks数据湖仓云原生部署并提供一定的扩缩容能力，因此一期中在低底层实现中使用原生资源进行各个组件的部署和运维。
   Roadmap
   一期
   在一期内目标可总结为完成整个starrocks服务的容器化以及单个CRD资源自动化部署整个集群能力，能够将starrocks整个数据湖仓系统通过终态资源starrocksCluster在kubernetes部署并能实现通用运维能力，服务自动重启，上下线扩缩容，升级等。
   [] 各个服务容器化部署
   [] 自定义 CRD，starrocksCluster
   [] 实现 FE、BE 自动化部署和建连
   [] 服务重启自动加入集群能力
   [] 可手动调整集群规模能力
   容器化实现
   适应于云原生化部署运维需要编写starrocks各个组件的容器化部署脚本和镜像化。基于k8s云原生化容器部署初始化脚本；容器化启动脚本；容器运行时检测脚本。
   Initial初始化脚本
   如果是fe
1. 获取servcie后端正在运行的fe实例，如果有且pod的序号为0设置为主节点启动方式，role设置为master，masterIp为空写入到/etc下一个配置文件。
2. 如果没有获取到service后端真正的fe实例，且序号不为0，sleep5秒重启。
3. 不是master，解析文件获取master的地址使用slave的方式启动。
   启动的时候通过环境变量设置fe的启动方式fqdn，脚本中设置helper的默认值为环境变量传递的默认值，helper的默认值通过初始化脚本设定。
   如果是be或者cn
1. 获取fe 的master的地址信息。
2. 通过mysql将be或者cn挂载到fe中，fe的地址获取通过初始化脚本获得存入文件和环境变量中。
   PreStop退出脚本
   容器化退出脚本在prestop中使用，在退出之前执行。
1. 通过dig命令获取fe service后端的服务列表，如果只有本机直接退出，不是的话先从fe集群中下掉自己。
2.
3. 如果是be将be置为不可用状态。
4. 如果是cn，先从fe中将cn删除再缩容掉cn。
   二期
   构建一套跟starrocks数仓平台服务化的operator是一个伴随starrocks发展长期迭代的过程，在一期中暂时使用原生资源的能力将整个starrocks部署和部分运维自动化掉，二期及以后得迭代中需要根据starrocks的特性以及用户需要设计具体服务能力的实现，如：对于be倍数扩缩容的能力，解决对于重新启动需要重新制定最新master加入的问题等等。从二期开始抛弃原生资源管控pod的能力，实现starrocks operator自管pod的能力，以自管pod为基础实现用户定制化需求。
   未来
   对于starrocks来说用户未来的需求将更加多样化，用户对于不同层面也将会不同能力的调度需求。在调度多样性化的同时服务体系的完备性，调度运维的高效性都应该是一个大数据操作中心平台应该具备的能力。未来对于starrocks operator的演进将从3个方面进行：调度服务多样化；服务体系完备性包括监控，日志收集，大数据接入等；调度运维高效性。