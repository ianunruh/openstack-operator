#!/bin/bash
set -ex

revision=$(rally db revision)
if [ $revision = "None" ]; then
    rally db create
else
    rally db upgrade
fi
