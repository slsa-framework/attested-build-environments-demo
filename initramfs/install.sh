#!/bin/bash

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

cp -r initramfs-tools /usr/share/initramfs-tools