# Windows HostProcess Pod: "failed to get application name from commandine: failed to find executable"

See the full article here: [Windows HostProcess Pod: "failed to get application name from commandine: failed to find executable"](https://coolstercodes.com/windows-hostprocess-pod-failed-to-get-application-name-from-commandine-failed-to-find-executable/)

## Deploy an Windows HostProcess cabale Kubernetes cluster using AKS-Engine
`Fill in subscription id in make-aks-engine-cluster.ps1`  
`.\make-aks-engine-cluster.ps1`

## Containerize a simple go-lang application
`docker build -t winhostprocessfailedtofindexecutable:v1.0.0 -f Dockerfile .`

## Launch the HostProcess Pod: "`failed to get application name from command line`" fixed
`.\kubectl --kubeconfig .\_output\win-cluster\kubeconfig\kubeconfig.westus2.json apply -f hostprocesspod.yaml`

