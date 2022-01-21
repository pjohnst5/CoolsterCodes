# Windows Pod File Mount

See the full article here: [Windows Pod File Mount]()

## Build the go-lang program which reads and edits a host file
`docker build -t windowspodfilemount:v1.0.0 -f Dockerfile .`

## Deploy and AKS-Engine Windows HostProcess capable cluster
`Fill in variables in make-aks-engine-cluster.ps1`
`.\make-aks-engine-cluster.ps1`

## Deploy a Windows pod with host file mount
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json apply -f windowspodfilemount.yaml`

## Check on pod status
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json get pods`

## Read pod logs
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json logs winpodfilemount`