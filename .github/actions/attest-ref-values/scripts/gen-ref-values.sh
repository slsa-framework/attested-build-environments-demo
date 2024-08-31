#!/bin/bash

# NOTE : Quote it else use array to avoid problems #
FILES="$@"

for f in $FILES
do
    echo "Generating in-toto SCAI attestation for $f ..."
  
    scai-gen assert --out-file /tmp/$f.ref.json REFERENCE_VALUE

    scai-gen rd file --name "$(basename $f)" --out-file /tmp/$f.rd.json $f

    scai-gen report --subject /tmp/$f.rd.json --out-file /tmp/$f.scai.json --pretty-print /tmp/$f.ref.json

    scai-gen sigstore --out-file /tmp/$f.scai.sigstore.json /tmp/$f.scai.json
done
