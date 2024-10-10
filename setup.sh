#!/bin/sh

python3 -m venv venv && \
. venv/bin/activate && \
pip3 install -U pip && \
pip3 install wheel && \
pip3 install --upgrade  -r requirements.txt
