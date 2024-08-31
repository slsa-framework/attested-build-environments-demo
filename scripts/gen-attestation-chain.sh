#!/bin/bash

PROVENANCE_FILE=examples/chkimes-image-attestation-attestation-675331.sigstore.json
TPM_QUOTE_FILE=examples/attest.json
LAUNCH_STATEMENT_FILE=examples/launch.in-toto.json

echo "DOWNLOAD SLSA PROVENANCE ATTESTATION FOR VMLINUZ"

curl -o $PROVENANCE_FILE https://github.com/chkimes/image-attestation/attestations/675331/download

echo -e "\n\nGENERATE VM INSTANCE LAUNCH STATEMENT"

attestor -t $PROVENANCE_FILE -e $TPM_QUOTE_FILE -o $LAUNCH_STATEMENT_FILE --pretty-print examples/vm.id