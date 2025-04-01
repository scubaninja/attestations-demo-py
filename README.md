# attestations-demo-py

This repository demonstrates the use of the
[`actions/attest-build-provenance`][1] and [`actions/attest-sbom`][2] actions to
generate attestations for software artifacts built with GitHub Actions.

We use a trivial Python project to show what a typical workflow might look like
(see [.github/workflows/build.yml][2]).

## Artifact Attestation

First, we need to ensure that we have the proper permissions set for our
workflow:

```yaml
permissions:
  contents: read
  id-token: write
  attestations: write
```

Developers are probably already familiar with `contents: read` as this is
necessary for the workflow to clone the repository.

The `id-token: write` permision is necessary for the creation of the
attestation. The `attest-*` actions will request an OIDC token from the GitHub
Actions runtime -- the identity information present in the OIDC token is
embedded in the generated attestation and provides the verifiable papertrail
that allows us to assert that the built artifact originated from this workflow.

The `attestations: write` permission is necessary to write the attestation to
the GitHub API and associate it with the current repository.

With the proper permissions set, we can build our Python project:

```yaml
- name: Install dependencies
  run: python -m pip install -r requirements.txt
- name: Build package
  run: python -m build .
```

The result of the build step will be a Python wheel package in the `dist/`
directory.

With our artifact built, we use the `actions/attest-build-provenance` action to
generate a [build provenance attestation][4] which refers to the Python package.

```yaml
- name: Attest build provenance
  uses: actions/attest-build-provenance@v1
  with:
    subject-path: "dist/*.whl"
```

See an example of a build provenance attestation [here][5].

Developers may also wish to generate an Sofware Bill of Materials (SBOM) for
their project and then attest that SBOM in order to link it to this workflow
run. There are a number of tools which can be used to generate SBOMs for a
project -- in our example we use the `anchore/sbom-action`:

```yaml
- name: Generate SBOM
  uses: anchore/sbom-action@v0.15.11
  with:
    format: "spdx-json"
    output-file: "sbom.spdx.json"
```

This will scan the project directory and produce an inventory of all the
referenced dependencies. In this case, we're asking for an SBOM in the [SPDX][6]
format.

Once the SBOM has been generated, we can use the `attest-sbom` action to wrap
the SBOM in a verifiable attestation:

```yaml
- name: Attest SBOM
  uses: actions/attest-sbom@v1
  with:
    subject-path: "dist/*.whl"
    sbom-path: "sbom.spdx.json"
```

Note that The subject used for the SBOM attestation is the same that was used
for the build provenance attestation. In both cases, we're binding the
attestation to the build artifact generated in this workflow.

See an example of an SBOM attestation [here][7].

## Attestation Verification

The "verify" job in the workflow shows an example of using the `gh` CLI to
verify the attestations which were generated for the build artifact:

```yaml
- name: Verify build artifact
  env:
    GH_TOKEN: ${{ github.token }}
    run: ./bin/gh attestation verify *.whl --owner github
```

The output will look something like the following, showing both the build
provenance and SBOM attestations:

```
github/attest-demo	https://spdx.dev/Document/v2.3	.github/workflows/build.yml@refs/heads/main
github/attest-demo	https://slsa.dev/provenance/v1	.github/workflows/build.yml@refs/heads/main
```

[1]: https://github.com/actions/attest-build-provenance
[2]: https://github.com/actions/attest-sbom
[3]: .github/workflows/build.yml
[4]: https://slsa.dev/spec/v1.0/provenance
[5]: https://github.com/github/attest-demo/attestations/762556
[6]: https://spdx.dev/
[7]: https://github.com/github/attest-demo/attestations/762557
