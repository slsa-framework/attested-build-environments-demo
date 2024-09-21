# CLI for SLSA BuildEnv Track

This is a proof-of-concept for the OpenSSF SLSA draft
[Attested Build Environments (BuildEnv) track](https://github.com/slsa-framework/slsa/issues/975).

The CLI in this repo implements vTPM-based attestation and
integrity checking of a Linux VM image. This repo also provides
demo GHA workflows showcasing how to meet SLSA BuildEnv L1 and L2 (WIP).

## How To Use

### Generate the initramfs

From a fresh Ubuntu 20+ VM, install the initramfs scripts:
```
sudo initramfs/install.sh
```

Generate the initramfs:
```
sudo mkinitramfs -o image-attestation.img
```

### Update GRUB

TODO

### CLI

Requires Go 1.21+

## TODOs

* Implement DSSE signing for `ref-values` command
* Modify `verify` command to use reference value attestations, rather than raw inputs
* Document verifier VM attestation flow
* Document private key config and signing attestation
* Add binding attestation + signature for the job id
* Add build image components for container-based build
* Add verification of SLSA Provenance + VSA generation
* Add verification of "boot" in container-based build environment
* Add mock build platform
* Add mock L3 container-based build environment deployment with HW TPM

## Disclaimer

This project is not ready for production use.
