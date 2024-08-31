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
RUN wget https://raw.githubusercontent.com/chkimes/image-attestation/main/build-container/hello.py
RUN python3 hello.py
