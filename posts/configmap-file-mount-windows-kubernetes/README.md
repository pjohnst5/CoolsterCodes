# ConfigMap File Mount on a Windows Kubernetes Pod

See the full article here: [ConfigMap File Mount on a Windows Kubernetes Pod](https://coolstercodes.com/configmap-file-mount-on-a-windows-kubernetes-pod/)

## Deploy an AKS-Engine Windows cluster
`Fill in subscription id in make-aks-engine-cluster.ps1`  
`.\make-aks-engine-cluster.ps1`

## Build the go-lang program which reads and edits a host file
`cd App`  
`docker build -t configmapfilemount:v1.0.0 -f Dockerfile .`

## Deploy a Windows pod with configMap mounted as a file
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json apply -f configmapfilemount.yaml`

## Read pod logs
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json logs configmapfilemount`
