FROM 3blades/core-server

RUN apt-get update && \
    apt-get install -yq --no-install-recommends \
    libcurl4-gnutls-dev \
    libssl-dev \
    libzmq3-dev \
    python3-pip \
    r-base \
    r-cran-jsonlite && \
    apt-get clean && \
    apt-get autoclean && \
    apt-get autoremove

RUN pip3 install -U pip setuptools wheel
RUN pip3 install jupyter_kernel_gateway

RUN Rscript --vanilla -e "install.packages(c('crayon', 'pbdZMQ', 'devtools'), repos='http://cran.r-project.org')"
RUN Rscript --vanilla -e "devtools::install_github(paste0('IRkernel/', c('repr', 'IRdisplay', 'IRkernel')))"
RUN Rscript --vanilla -e "IRkernel::installspec()"

RUN echo "en_US.UTF-8 UTF-8" > /etc/locale.gen && \
    locale-gen

WORKDIR /


COPY runner /
