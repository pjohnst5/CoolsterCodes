$SUBSCRIPTION_ID='65ce76d9-2f1b-4c0f-849c-163ddd03b7a9'
$CLUSTER_NAME='win-host-process'
$LOCATION='westus2'
$API_MODEL='host-process-cluster.json'

az group create --subscription $SUBSCRIPTION_ID --location $LOCATION --name $CLUSTER_NAME

if [[ -z "${CLIENT_ID}" ]]; then
    AZSP=$(az ad sp create-for-rbac --name paujohns-aks-engine-$(date +"%s") --role="Owner" --scopes="/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${DNS_PREFIX}" -o json)
    CLIENT_ID=$(echo $AZSP | jq -r '.appId')
    CLIENT_SECRET=$(echo $AZSP | jq -r '.password')
fi
echo Using:
echo Client ID $CLIENT_ID
echo Client Secret $CLIENT_SECRET
sleep 180

aks-engine deploy \
    --subscription-id $SUBSCRIPTION_ID  \
    --client-id $CLIENT_ID \
    --client-secret $CLIENT_SECRET \
    --resource-group $DNS_PREFIX \
    --dns-prefix $DNS_PREFIX \
    --location $LOCATION \
    --api-model $API_MODEL \
    --force-overwrite
