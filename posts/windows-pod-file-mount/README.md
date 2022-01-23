# Windows Pod Host File Mount on Kubernetes

See the full article here: [Windows Pod Host File Mount on Kubernetes](https://coolstercodes.com/windows-pod-host-file-mount-on-kubernetes/)

## Deploy an AKS-Engine Windows cluster
`Fill in subscription id in make-aks-engine-cluster.ps1`  
`.\make-aks-engine-cluster.ps1`

## Set proper perms to directory on host
`<RDP to windows agent node>`  
`icacls C:\users\azureuser /grant BUILTIN\Users:(OI)(CI)(F) /t`

## Build the go-lang program which reads and edits a host file
`cd App`  
`docker build -t windowspodfilemount:v1.0.0 -f Dockerfile .`

## Deploy a Windows pod with host file mount
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json apply -f windowspodfilemount.yaml`

## Read pod logs
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json logs winpodfilemount`

## Alternative, if you just need to write to host, and does not need to be existing dir/file:
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json apply -f windowspodfilemount_alternative.yaml`