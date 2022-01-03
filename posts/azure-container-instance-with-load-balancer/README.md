# Azure Container Instance with Load Balancer

Article: [Azure Container Instance with Load Balancer]()

## Instructions

### ARM template
```
az group create -n <rg-name> -l <location>
az deployment group create -f ./azure-container-instance-with-load-balancer.json -g <rg-name>
```

### az cli