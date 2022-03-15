#source "$(dirname "$0")/common.sh"

#live_data "/var/flow" "testnet.devnet32"

json=$(wget -O - https://raw.githubusercontent.com/onflow/flow/m4ksio/reorganize-sporks.json/sporks.json)
network=$(jq '.networks.testnet.devnet32' <<< "$json")

jq -r .stateArtefacts.gcp.rootCheckpointFile <<< "$network" # ./root.checkpoint
jq -r .stateArtefacts.gcp.rootProtocolStateSnapshot <<< "$network" # ./public-root-information/root-protocol-state-snapshot.json
jq -r .stateArtefacts.gcp.nodeInfo <<< "$network" # ./public-root-information/node-infos.pub.json


SEED_ADDRESS=$(jq -r .seedNodes[0].address <<< "$network")
SEED_KEY=$(jq -r .seedNodes[0].key <<< "$network")

GCP_BUCKET=$(jq -r .stateArtefacts.gcp.executionStateBucket <<< "$network")


echo $SEED_ADDRESS
echo $SEED_KEY
echo $GCP_BUCKET

# make sure you build a docker image first
# dir where data has been downloaded should be mounted
docker run \
  -v /var/flow:/data/ \
  -p 5005:5005 \
  flow-dps-live:v0.23 \
  -a 0.0.0.0:5005 \
  -i /data/devnet-32/index \
  -b /data/devnet-32/bootstrap \
  -c /data/devnet-32/root.checkpoint \
  -d /data/devnet-32/protocol \
  -l info \
  -u "$GCP_BUCKET" \
  --seed-address="$SEED_ADDRESS" \
  --seed-key="$SEED_KEY"


