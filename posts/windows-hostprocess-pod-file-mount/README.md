# Windows HostProcess Pod File Mount

See the full article here: [Windows HostProcess Pod File Mount]()

## Build the go-lang program which reads and edits a host file
`docker build -t hostprocesspod:v1.0.0 -f Dockerfile .`

## Deploy and AKS-Engine Windows HostProcess capable cluster
`Fill in variables in make-aks-engine-cluster.ps1`
`.\make-aks-engine-cluster.ps1`

## Deploy a Windows HostProcess pod
`.\kubectl --kubeconfig .\_output\<cluster_name>\kubeconfig\kubeconfig.<region>.json apply -f hostprocesspod.yaml`

## Check on pod status
`.\kubectl --kubeconfig .\_output\<cluster_name>\kubeconfig\kubeconfig.<region>.json get pods`

## Read pod logs
`.\kubectl --kubeconfig .\_output\<cluster_name>\kubeconfig\kubeconfig.<region>.json logs hostprocesspod`