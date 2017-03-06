FROM 3blades/core-server

RUN apt-get update && \
    apt-get install -yq --no-install-recommends \
    python3-dev \
    python3-pip && \
    apt-get clean && \
    apt-get autoclean && \
    apt-get autoremove

RUN pip3 install -q --upgrade pip wheel setuptools
RUN pip3 install \
    jupyter_kernel_gateway \
    psycopg2 \
    pymongo \
    cassandra-driver \
    mysqlclient

WORKDIR /


COPY runner /
