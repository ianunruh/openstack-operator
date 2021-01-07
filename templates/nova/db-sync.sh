#!/bin/bash
set -ex

nova-manage api_db sync

nova-manage cell_v2 map_cell0 --database_connection $CELL_DATABASE

nova-manage db sync
