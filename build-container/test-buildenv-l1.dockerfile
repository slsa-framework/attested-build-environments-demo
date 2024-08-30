FROM ubuntu:24.04

RUN apt-get update \
    && apt-get install -y -q \
        build-essential \
        curl \
        git \
        python3 \
        tar \
        wget \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /tmp
RUN wget https://github.com/chkimes/image-attestation/blob/buildenv-l1-container/build-container/hello.py
RUN python3 hello.py
