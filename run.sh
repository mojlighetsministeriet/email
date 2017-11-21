#!/bin/sh

ARCH=$(uname -m)

if [ $ARCH = "armv8" ]; then
  ./email-arm64
else
  ./email-amd64
fi
