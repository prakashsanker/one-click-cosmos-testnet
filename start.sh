#!/bin/bash

cp /validator-config/node_key.json /.test-chain/config/node_key.json &&
cp /validator-config/priv_validator_key.json /.test-chain/config/priv_validator_key.json &&
cp /validator-config/config.toml /.test-chain/config/config.toml &&
/dist/test-chaind start --home /.test-chain