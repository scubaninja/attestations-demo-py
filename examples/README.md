# Examples
This directory contains a simple script to demonstrate how to use the `jq` tool to process a raw attesation file that comes from GitHub.

The `sample-artifact` directory holds:
* An artifact created by running the `build.yaml` workflow in this repo to produce an `artifact.zip`
* The unzipped Python package in `whl` format
* The downloaded attestations generated at build time (SBOM and provenance)

## Purpose
This is intended to memorialize the basics of processing a downloaded attestation file in a platform-agnostic way with a minimal toolkit.

## Usage
```sh
./print-attestation.sh < sample-artifact/attestations.jsonl
```
