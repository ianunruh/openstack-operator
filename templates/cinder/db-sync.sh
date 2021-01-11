#!/bin/bash
set -ex

exec cinder-manage db sync
