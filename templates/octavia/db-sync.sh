#!/bin/bash
set -ex

octavia-db-manage upgrade head
