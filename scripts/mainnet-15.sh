source common.sh

live_data "/var/flow" "mainnet-15"  "https://storage.googleapis.com/flow-genesis-bootstrap/mainnet-15-execution" "c8d4c90a40ec74aaf408bd7205d533e8b1901016f54695cbd71e0be4cae8725a"

# make sure you build a docker image first
# dir where data has been downloaded should be mounted
docker run \
  -v /var/flow:/data/ \
  -p 5005:5005 \
  flow-dps-live:v0.23 \
  -a 0.0.0.0:5005 \
  -i /data/mainnet-15/index \
  -b /data/mainnet-15/bootstrap \
  -c /data/mainnet-15/root.checkpoint \
  -d /data/mainnet-15/protocol \
  -l info \
  -u flow_public_mainnet15_execution_state \
  --seed-address="access-007.mainnet15.nodes.onflow.org:3569" \
  --seed-key="28a0d9edd0de3f15866dfe4aea1560c4504fe313fc6ca3f63a63e4f98d0e295144692a58ebe7f7894349198613f65b2d960abf99ec2625e247b1c78ba5bf2eae"

