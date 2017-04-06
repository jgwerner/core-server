FROM ubuntu:16.04

RUN apt-get update && \
    apt-get install -yq --no-install-recommends \
    software-properties-common

RUN add-apt-repository -y main
RUN add-apt-repository -y universe
RUN add-apt-repository -y restricted
RUN add-apt-repository -y multiverse

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -yq --no-install-recommends \
    wget \
    bzip2 \
    ca-certificates \
    build-essential \
    libpq-dev \
    libmysqlclient-dev \
    openjdk-8-jre-headless \
    python-dev \
    python-pip && \
    apt-get clean && \
    apt-get autoclean && \
    apt-get autoremove

RUN echo "en_US.UTF-8 UTF-8" > /etc/locale.gen && \
    locale-gen

RUN pip install -q --upgrade pip wheel setuptools
RUN pip install -q jupyter_kernel_gateway

WORKDIR /

ENV APACHE_SPARK_VERSION 2.0.2
ENV HADOOP_VERSION 2.7

RUN wget -qO- http://d3kbcqa49mib13.cloudfront.net/spark-${APACHE_SPARK_VERSION}-bin-hadoop${HADOOP_VERSION}.tgz | tar xz
RUN mv spark-${APACHE_SPARK_VERSION}-bin-hadoop${HADOOP_VERSION} spark

ENV SPARK_HOME=/spark
ENV PATH=$SPARK_HOME/bin:$PATH
ENV PYTHONPATH $SPARK_HOME/python:$SPARK_HOME/python/lib/py4j-0.10.3-src.zip
ENV KERNEL_NAME python

WORKDIR /

COPY runner /
