# image-attestation

This is a PoC for the OpenSSF SLSA workstream
[Hardware Attested Platforms](https://github.com/slsa-framework/slsa/issues/975).

The CLI in this repo implements vTPM-based attestation and
integrity checking of a Linux VM image.

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

### Disclaimer

This project is not ready for production use.
