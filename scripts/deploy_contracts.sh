#!/bin/bash
# Exit script if any command fails (-e)
# Exit script if any variable is undefined (-u) 
# Exit script if any command in a pipeline fails (-o pipefail)
set -euo pipefail

# Helper to extract contract address from tx output
extract_address() {
  local txhash="$1"
  local address
  
  # Debug: Print the full transaction output
  echo "Debug: Full transaction output:" >&2
  sdknewd query tx "$txhash" --output json | jq '.' >&2
  
  # Debug: Print all event types
  echo "Debug: Event types:" >&2
  sdknewd query tx "$txhash" --output json | jq -r '.events[].type' >&2
  
  # Updated jq query to match the actual structure
  address=$(sdknewd query tx "$txhash" --output json | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
  
  if [[ -z "$address" || "$address" == "null" ]]; then
    echo "Failed to extract contract address from tx: $txhash" >&2
    exit 1
  fi
  echo "$address"
}

# Helper to extract code ID from store tx output
extract_code_id() {
  local txhash="$1"
  local code_id
  
  # Debug: Print the full transaction output
  echo "Debug: Full transaction output:" >&2
  sdknewd query tx "$txhash" --output json | jq '.' >&2
  
  # Debug: Print all event types
  echo "Debug: Event types:" >&2
  sdknewd query tx "$txhash" --output json | jq -r '.events[].type' >&2
  
  # Updated jq query to match the actual structure
  code_id=$(sdknewd query tx "$txhash" --output json | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
  
  if [[ -z "$code_id" || "$code_id" == "null" ]]; then
    echo "Failed to extract code ID from tx: $txhash" >&2
    exit 1
  fi
  echo "$code_id"
}

# Set these to your actual values
ADMIN="shareledger1j4ndn4qed0ukulc0a6a5fxe6yxmnm7wh59pwpn"
ISSUER="shareledger1j4ndn4qed0ukulc0a6a5fxe6yxmnm7wh59pwpn"
APPROVER="approver" # key name in keyring

# Set keyring backend
export KEYRING_BACKEND="test"

# DID Registry Contract
echo "=== Deploying DID Registry Contract ==="
echo "Uploading DID Registry wasm..."
DID_REGISTRY_STORE_TX=$(sdknewd tx wasm store ./wasmfiles/did_registry.wasm --from "$APPROVER" --gas auto --gas-adjustment 1.3 --broadcast-mode sync -y --output json | jq -r '.txhash')
echo "DID Registry store tx: $DID_REGISTRY_STORE_TX"
sleep 3
DID_REGISTRY_CODE_ID=$(extract_code_id "$DID_REGISTRY_STORE_TX")
echo "DID Registry code ID: $DID_REGISTRY_CODE_ID"

echo "Instantiating DID Registry..."
DID_REGISTRY_INIT='{}'
DID_REGISTRY_LABEL="DID Registry"
DID_REGISTRY_TX=$(sdknewd tx wasm instantiate "$DID_REGISTRY_CODE_ID" "$DID_REGISTRY_INIT" --label "$DID_REGISTRY_LABEL" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --admin "$ADMIN" --keyring-backend test --broadcast-mode sync --output json | jq -r '.txhash')
echo "DID Registry tx: $DID_REGISTRY_TX"
echo "Waiting for transaction to be processed..."
sleep 3
DID_REGISTRY_ADDRESS=$(extract_address "$DID_REGISTRY_TX")
echo "DID Registry address: $DID_REGISTRY_ADDRESS"

# HD Wallet Index Contract
echo "=== Deploying HD Wallet Index Contract ==="
echo "Uploading HD Wallet Index wasm..."
HD_WALLET_INDEX_STORE_TX=$(sdknewd tx wasm store ./wasmfiles/hd_wallet_index.wasm --from "$APPROVER" --gas auto --gas-adjustment 1.3 --broadcast-mode sync -y --output json | jq -r '.txhash')
echo "HD Wallet Index store tx: $HD_WALLET_INDEX_STORE_TX"
sleep 3
HD_WALLET_INDEX_CODE_ID=$(extract_code_id "$HD_WALLET_INDEX_STORE_TX")
echo "HD Wallet Index code ID: $HD_WALLET_INDEX_CODE_ID"

echo "Instantiating HD Wallet Index..."
HD_WALLET_INDEX_INIT="{\"issuer\":\"$ISSUER\",\"contract_owner\":\"$ADMIN\",\"allowed_callers\":null}"
HD_WALLET_INDEX_LABEL="HD Wallet Index"
HD_WALLET_INDEX_TX=$(sdknewd tx wasm instantiate "$HD_WALLET_INDEX_CODE_ID" "$HD_WALLET_INDEX_INIT" --label "$HD_WALLET_INDEX_LABEL" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --admin "$ADMIN" --keyring-backend test --broadcast-mode sync --output json | jq -r '.txhash')
echo "HD Wallet Index tx: $HD_WALLET_INDEX_TX"
echo "Waiting for transaction to be processed..."
sleep 3
HD_WALLET_INDEX_ADDRESS=$(extract_address "$HD_WALLET_INDEX_TX")
echo "HD Wallet Index address: $HD_WALLET_INDEX_ADDRESS"

# VCT Contract
echo "=== Deploying VCT Contract ==="
echo "Uploading VCT wasm..."
VCT_STORE_TX=$(sdknewd tx wasm store ./wasmfiles/vct.wasm --from "$APPROVER" --gas auto --gas-adjustment 1.3 --broadcast-mode sync -y --output json | jq -r '.txhash')
echo "VCT store tx: $VCT_STORE_TX"
sleep 3
VCT_CODE_ID=$(extract_code_id "$VCT_STORE_TX")
echo "VCT code ID: $VCT_CODE_ID"

echo "Instantiating VCT..."
VCT_INIT="{\"name\":\"Verifiable Credentials Token\",\"symbol\":\"VCT\",\"issuer\":\"$ISSUER\",\"contract_owner\":\"$ADMIN\",\"allowed_callers\":null,\"hd_wallet_index_address\":\"$HD_WALLET_INDEX_ADDRESS\",\"did_registry_address\":\"$DID_REGISTRY_ADDRESS\"}"
VCT_LABEL="Verifiable Credentials Token"
VCT_TX=$(sdknewd tx wasm instantiate "$VCT_CODE_ID" "$VCT_INIT" --label "$VCT_LABEL" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --admin "$ADMIN" --keyring-backend test --broadcast-mode sync --output json | jq -r '.txhash')
echo "VCT tx: $VCT_TX"
echo "Waiting for transaction to be processed..."
sleep 3
VCT_ADDRESS=$(extract_address "$VCT_TX")
echo "VCT address: $VCT_ADDRESS"

# dVCT Contract
echo "=== Deploying dVCT Contract ==="
echo "Uploading dVCT wasm..."
DVCT_STORE_TX=$(sdknewd tx wasm store ./wasmfiles/dvct.wasm --from "$APPROVER" --gas auto --gas-adjustment 1.3 --broadcast-mode sync -y --output json | jq -r '.txhash')
echo "dVCT store tx: $DVCT_STORE_TX"
sleep 3
DVCT_CODE_ID=$(extract_code_id "$DVCT_STORE_TX")
echo "dVCT code ID: $DVCT_CODE_ID"

echo "Instantiating dVCT..."
DVCT_INIT="{\"name\":\"Document Verifiable Credentials Token\",\"symbol\":\"dVCT\",\"issuer\":\"$ISSUER\",\"contract_owner\":\"$ADMIN\",\"allowed_callers\":null,\"vct_address\":\"$VCT_ADDRESS\",\"hd_wallet_index_address\":\"$HD_WALLET_INDEX_ADDRESS\",\"did_registry_address\":\"$DID_REGISTRY_ADDRESS\"}"
DVCT_LABEL="Document Verifiable Credentials Token"
DVCT_TX=$(sdknewd tx wasm instantiate "$DVCT_CODE_ID" "$DVCT_INIT" --label "$DVCT_LABEL" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --admin "$ADMIN" --keyring-backend test --broadcast-mode sync --output json | jq -r '.txhash')
echo "dVCT tx: $DVCT_TX"
echo "Waiting for transaction to be processed..."
sleep 3
DVCT_ADDRESS=$(extract_address "$DVCT_TX")
echo "dVCT address: $DVCT_ADDRESS"

# Fee Contract
echo "=== Deploying Fee Contract ==="
echo "Uploading Fee wasm..."
FEE_STORE_TX=$(sdknewd tx wasm store ./wasmfiles/fee.wasm --from "$APPROVER" --gas auto --gas-adjustment 1.3 --broadcast-mode sync -y --output json | jq -r '.txhash')
echo "Fee store tx: $FEE_STORE_TX"
sleep 3
FEE_CODE_ID=$(extract_code_id "$FEE_STORE_TX")
echo "Fee code ID: $FEE_CODE_ID"

echo "Instantiating Fee..."
FEE_INIT="{\"contract_owner\":\"$ADMIN\",\"allowed_caller\":\"$ADMIN\",\"fee_collector\":\"$ADMIN\"}"
FEE_LABEL="Fee Contract"
FEE_TX=$(sdknewd tx wasm instantiate "$FEE_CODE_ID" "$FEE_INIT" --label "$FEE_LABEL" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --admin "$ADMIN" --keyring-backend test --broadcast-mode sync --output json | jq -r '.txhash')
echo "Fee tx: $FEE_TX"
echo "Waiting for transaction to be processed..."
sleep 3
FEE_ADDRESS=$(extract_address "$FEE_TX")
echo "Fee address: $FEE_ADDRESS"

# Set Fee Amounts
echo "=== Setting Fee Amounts ==="
echo "Setting fee amount for VCT..."
VCT_FEE_TX=$(sdknewd tx wasm execute "$FEE_ADDRESS" "{\"set_fee_amount\": {\"fee_id\": \"vct_mint\", \"fee_amount\": {\"amount\": \"1\", \"denom\": \"nshr\"}}}" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --broadcast-mode sync --output json | jq -r '.txhash')
echo "VCT fee set tx: $VCT_FEE_TX"
sleep 3

echo "Setting fee amount for dVCT..."
DVCT_FEE_TX=$(sdknewd tx wasm execute "$FEE_ADDRESS" "{\"set_fee_amount\": {\"fee_id\": \"dvct_mint\", \"fee_amount\": {\"amount\": \"1\", \"denom\": \"nshr\"}}}" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --broadcast-mode sync --output json | jq -r '.txhash')
echo "dVCT fee set tx: $DVCT_FEE_TX"
sleep 3

# Set Fee Addresses
echo "=== Setting Fee Addresses ==="
echo "Setting fee address for VCT..."
VCT_SET_FEE_TX=$(sdknewd tx wasm execute "$VCT_ADDRESS" "{\"set_fee_address\": {\"fee_address\": \"$FEE_ADDRESS\"}}" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --broadcast-mode sync --output json | jq -r '.txhash')
echo "VCT set fee tx: $VCT_SET_FEE_TX"
sleep 3

echo "Setting fee address for dVCT..."
DVCT_SET_FEE_TX=$(sdknewd tx wasm execute "$DVCT_ADDRESS" "{\"set_fee_address\": {\"fee_address\": \"$FEE_ADDRESS\"}}" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --broadcast-mode sync --output json | jq -r '.txhash')
echo "dVCT set fee tx: $DVCT_SET_FEE_TX"
sleep 3

# Set Allowed Callers for HD Wallet Index
echo "=== Setting Allowed Callers for HD Wallet Index ==="
HD_WALLET_INDEX_SET_CALLERS_TX=$(sdknewd tx wasm execute "$HD_WALLET_INDEX_ADDRESS" "{\"set_allowed_callers\": {\"allowed_callers\": [\"$VCT_ADDRESS\", \"$DVCT_ADDRESS\"]}}" --from "$APPROVER" --gas auto --gas-adjustment 1.3 -y --broadcast-mode sync --output json | jq -r '.txhash')
echo "HD Wallet Index set allowed callers tx: $HD_WALLET_INDEX_SET_CALLERS_TX"
sleep 3

echo "=== Deployment Summary ==="
echo "DID Registry: $DID_REGISTRY_ADDRESS"
echo "HD Wallet Index: $HD_WALLET_INDEX_ADDRESS"
echo "VCT: $VCT_ADDRESS"
echo "dVCT: $DVCT_ADDRESS"
echo "Fee Contract: $FEE_ADDRESS"