FROM nvidia/cuda:8.0-cudnn6-devel-ubuntu16.04

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
	locales \
    ca-certificates \
    build-essential \
    libpq-dev \
    libmysqlclient-dev  && \
    apt-get clean && \
    apt-get autoclean && \
    apt-get autoremove

RUN echo "en_US.UTF-8 UTF-8" > /etc/locale.gen && \
    locale-gen

# Configure environment
ENV CONDA_DIR /opt/conda
ENV PATH $CONDA_DIR/bin:$PATH
ENV SHELL /bin/bash
ENV LC_ALL en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US.UTF-8

# Install conda
RUN cd /tmp && \
    mkdir -p $CONDA_DIR && \
    wget --quiet https://repo.continuum.io/miniconda/Miniconda3-4.2.12-Linux-x86_64.sh && \
    echo "c59b3dd3cad550ac7596e0d599b91e75d88826db132e4146030ef471bb434e9a *Miniconda3-4.2.12-Linux-x86_64.sh" | sha256sum -c - && \
    /bin/bash Miniconda3-4.2.12-Linux-x86_64.sh -f -b -p $CONDA_DIR && \
    rm Miniconda3-4.2.12-Linux-x86_64.sh && \
    $CONDA_DIR/bin/conda install --quiet --yes conda==4.2.12 && \
    $CONDA_DIR/bin/conda config --system --add channels conda-forge && \
    conda clean -tipsy

# Copy compiled runner
COPY runner /
