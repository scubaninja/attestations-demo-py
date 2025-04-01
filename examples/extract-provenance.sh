#! /bin/bash
set -euo pipefail 

provenancePredicateType="https://slsa.dev/provenance/v1"
provenanceFile="provenance"

while read -r line
do
  base64Payload=$(echo $line | jq -r '.["dsseEnvelope"]["payload"]')
  decodedPayload=$(echo $base64Payload | base64 -d)
  predicateType=$(echo $decodedPayload | jq -r '.["predicateType"]')

  if [ "$predicateType" == "$provenancePredicateType" ]
  then
    subjectDigest=$(echo $decodedPayload | jq -r '.["subject"][0]["digest"]["sha256"]')
    filename=$(echo "$provenanceFile-$subjectDigest.json")

    echo
    echo $decodedPayload > $filename
    echo "Provenance predicate extracted to $filename"
    echo
    break 
  else
    continue
  fi
done