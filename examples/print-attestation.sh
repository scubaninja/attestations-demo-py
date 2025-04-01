#! /bin/bash
set -euo pipefail 

while read -r line
do
  base64Payload=$(echo $line | jq -r '.["dsseEnvelope"]["payload"]')
  decodedPayload=$(echo $base64Payload | base64 -d)

  predicateType=$(echo $decodedPayload | jq -r '.["predicateType"]')

  if [ "$predicateType" == "https://spdx.dev/Document/v2.3" ]
  then
    echo
    echo "SBOM predicate found... skipping"
    continue
  fi

  echo
  echo "---- Provenance details ----"

  echo -n "Builder ID: "; echo $decodedPayload | jq -r '.["predicate"]["runDetails"]["builder"]["id"]'
  echo -n "Build Type: ";  echo $decodedPayload | jq -r '.["predicate"]["buildDefinition"]["buildType"]'
  echo

  echo -n "Triggering Workflow: ";  echo $decodedPayload | jq -r '.["predicate"]["buildDefinition"]["externalParameters"]["workflow"]["path"]'
  echo -n "Repo: ";  echo $decodedPayload | jq -r '.["predicate"]["buildDefinition"]["externalParameters"]["workflow"]["repository"]'
  echo -n "Ref: ";  echo $decodedPayload | jq -r '.["predicate"]["buildDefinition"]["externalParameters"]["workflow"]["ref"]'

  echo
done