#!/bin/bash
set -e

HCPDEVSUBSCRIPTION="ARO HCP E2E"
az account set --name "${HCPDEVSUBSCRIPTION}"

SUBSCRIPTION=$(az account show -o json | jq -r '.id')
TENANT=$(az account show -o json | jq -r '.tenantId')

APP_ID=$(az ad app create --display-name "aro-github-actions-identity" | jq -r '.appId')
OBJ_ID=$(az ad sp create --id $APP_ID | jq -r '.id')

# Create role assignment and federate credentials
az role assignment create --assignee "${OBJ_ID}" --role Contributor --scope /subscriptions/${SUBSCRIPTION}
az role assignment create --assignee "${OBJ_ID}" --role "Role Based Access Control Administrator" --scope /subscriptions/${SUBSCRIPTION}

az ad app federated-credential create --id "${APP_ID}" --parameters \
'{
    "audiences": [
        "api://AzureADTokenExchange"
    ],
    "description": "https://github.com/Azure/ARO-HCP runner",
    "issuer": "https://token.actions.githubusercontent.com",
    "name": "aro-hcp-pr-runner",
    "subject": "repo:Azure/ARO-HCP:pull_request"
}'

echo "----------- Configure GitHub with the below secrets -----------"
echo "AZURE_CLIENT_ID: ${APP_ID}"
echo "AZURE_SUBSCRIPTION_ID: ${SUBSCRIPTION}"
echo "AZURE_TENANT_ID: ${TENANT}"
