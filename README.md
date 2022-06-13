# kubernetes-white-list-injector
在对权限要求相对较高的云上场景中，您需要将Pod的IP地址（或者pod所在的node节点地址）动态的加入或移出指定服务的白名单，如数据库、slb等，以实现对权限最细粒度的控制。您可以通过kubernetes-whitelist-injector组件为Pod添加Annotation，动态的将IP地址加入或移出指定的白名单

### 部署 && 配置，以本地k8s集群实现aliyun rds whiltelist为例
##### 1. 配置ak/sk 
编辑manager.yaml,   vi ./config/manager/manager.yaml，添加secret：
``` yaml
apiVersion: v1
kind: Secret
metadata:
  name: aliyun
  namespace: controller-manager
data:
  accessKeyId:               #your ak
  accessKeySecret:           #your sk
```

设置 secret volume ：
```yaml
volumeMounts:
        - name: aliyun
          mountPath: "/aliyunSecret"     #固定路径
          readOnly: true

volumes:
      - name: aliyun
        secret:
          secretName: aliyun
```

##### 2. 部署controller-manager
k8s需事先安装cert-manager： kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.3.1/cert-manager.yaml
```
make docker-builder 
make deploy   
```
##### 3. 创建whitelist实例
修改配置（./config/samples/apps_v1alpha1_whitelist.yaml）
```yaml
apiVersion: apps.whitelist.fly.io/v1alpha1
kind: Whitelist
metadata:
  name: whitelist-sample
spec:
  # TODO(user): Add fields here
  provider: aliyun
  service: rds
  ipLevel: Node
  serviceId: {your rdsId}
  annotations:               #可选
    DBInstanceIPArrayName: {your rds whiltelist group}    #可选
```
部署： kubectl apply -f ./config/samples/apps_v1alpha1_whitelist.yaml

##### 4. 部署应用，并能自动注册白名单
编辑应用deployment：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx-deployment
  template:
    metadata:
      labels:
        app: nginx-deployment
        whitelist.fly.io/whitelist-sample: ""                 #如果有多whitelist，可以添加多个labels，格式：  whitelist.fly.io/{whitelistName}:"" 
    spec:                                                     
      containers:
      - image: nginx:1.15
        imagePullPolicy: IfNotPresent
        name: nginx
        ports:
        - containerPort: 80
          protocol: TCP
        resources: {}
      dnsPolicy: ClusterFirst
      restartPolicy: Always
  ```
