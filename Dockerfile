FROM ubuntu:16.04

RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

# Copy compiled runner
COPY runner /
