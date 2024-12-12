#!/bin/bash

docker run --rm -it --shm-size=1024g --security-opt seccomp=unconfined -v /home/annalorimer/splitpt/tests/shadow_pt_tiny/:/mnt/ shadow_pt_tiny
