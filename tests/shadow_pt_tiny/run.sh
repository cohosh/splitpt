#!/bin/bash

# this should run inside the container

#set -euo pipefail

simdir=tornet__net-0.01__load-3.2__trial-1__pt-obfs4

tar xaf tornet__net-0.01__load-3.2__trial-1__pt-obfs4.tar.xz

SEED=`python3 -c 'import random; print(random.randint(0, 999999999))'`

tornettools simulate \
    --shadow /opt/bin/shadow \
    --args "--parallelism=36 --seed=${SEED} --template-directory=shadow.data.template" \
    --filename shadow.config.yaml \
    ${simdir}

tornettools parse ${simdir}
tornettools plot ${simdir} --prefix ${simdir}/pdfs
tornettools archive ${simdir}
