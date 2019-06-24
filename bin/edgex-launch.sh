#!/bin/bash
#
# Copyright (c) 2018
# Mainflux
#
# SPDX-License-Identifier: Apache-2.0
#

###
# Launches all EdgeX Go binaries (must be previously built).
#
# Expects that Consul and MongoDB are already installed and running.
#
###

DIR=$PWD
CMD=../cmd

# Kill all edgex-* stuff
function cleanup {
	pkill edgex
}

###
# Export Client
###
#cd $CMD/export-client
#exec -a edgex-export-client ./export-client &
#cd $DIR

###
# Export Distro
###
cd $CMD/export-distro
exec -a edgex-export-distro ./export-distro &
cd $DIR


trap cleanup EXIT

while : ; do sleep 1 ; done