#!/bin/bash

# enable unofficial bash strict mode
set -o errexit
set -o nounset
set -o pipefail

action=${1:-"dump"}

if [ "$action" = "delete" ]; then
    echo "Delete backup..."
    /bin/bash /delete.sh
elif [ "$action" = "restore" ]; then
    echo "Restore backup..."
    /bin/bash /restore.sh
else
    echo "Dump backup to remote..."
    /bin/bash /dump.sh
fi