#!/bin/bash
set -e

nova-manage cell_v2 list_cells | grep cell1 || nova-manage cell_v2 create_cell \
    --name=cell1 \
    --transport-url=$OS_DEFAULT__TRANSPORT_URL \
    --database_connection=$OS_DATABASE__CONNECTION

nova-manage db sync --local_cell
