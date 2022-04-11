# Windows HostProcess Pod Host File Access

See the full article here: [Windows HostProcess Pod Host File Access](https://coolstercodes.com/windows-hostprocess-pod-host-file-access/)

## Create a HostProcess capable Kubernetes clusters
`Fill in subscription id in make-aks-engine-cluster.ps1`  
`.\make-aks-engine-cluster.ps1`

## Make a HostProcess pod
`cd App`  
`docker build -t windowspodfilemount:v1.0.0 -f Dockerfile .`

## Deploy the HostProcess pod
`.\kubectl --kubeconfig .\_output\hostprocess-cluster\kubeconfig\kubeconfig.westus2.json apply -f hostprocesspod.yaml`

## Read pod logs
`.\kubectl --kubeconfig .\_output\hostprocess-cluster\kubeconfig\kubeconfig.westus2.json logs hostprocesspod`
