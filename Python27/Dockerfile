FROM 3blades/core-server

RUN apt-get update && \
    apt-get install -yq --no-install-recommends \
    python-dev \
    python-pip && \
    apt-get clean && \
    apt-get autoclean && \
    apt-get autoremove

RUN pip install -q --upgrade pip wheel
RUN pip install \
    jupyter_kernel_gateway \
    psycopg2 \
    pymongo \
    cassandra-driver \
    mysqlclient

WORKDIR /

COPY runner /
