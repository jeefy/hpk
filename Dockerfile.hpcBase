# Slurm Base image
FROM ubuntu:xenial

ENV LD_LIBRARY_PATH=/usr/local/lib/

RUN apt-get update -y && \
    apt-get install -y curl wget bzip2 make gcc lua5.3 liblua5.3-dev supervisor \
                       unzip openmpi-bin openmpi-common python-pip g++ vim

CMD ["bash"]
