#!/bin/bash

SELF_ROLE=
# Editlog port: default 9010
EDIT_LOG_PORT=9010
# Query port: default 9030
QUERY_PORT=9030
PROBE_INTERVAL=2
# now fe leader
FE_LEADER=
# host_type, default "IP"
HOST_TYPE=${HOST_TYPE:-"IP"}
#pod domain name for access
POD_HOSTNAME=$(hostname -f)
#pod ip
POD_IP=$(hostname -i)
#help host for transfer master
SELF_HOST=
#left_node_lists node lists of except self.
LEFT_NODE_LSITS=
PROBE_LEADER_PODX_TIMEOUT=120 # at most 60 attempts
STARROCK_HOME=${STARROCK_HOME:-"/opt/starrocks"}

show_frontends() {
    local svc=$1
    # ensure `mysql` command can be ended with 15 seconds
    # "show frontends" query will hang when there is no leader yet in the cluster
    timeout 15 mysql --connect-timeout 2 -h $svc -P $QUERY_PORT -u root --skip-column-names --batch -e 'show frontends;'
}

log_stderr() {
    echo "[$(date)] $@" >&2
}

#get mysql host
# get last hosts
# get self role
assignment_variable() {
    if [[ $HOST_TYPE == "IP" ]]; then
        SELF_HOST=$POD_IP
    else
        SELF_HOST=$POD_HOSTNAME
    fi

    local svc=$1
    local memlist=
    while true; do
        memlist=$(show_frontends $svc)
        local leader=$(echo "$memlist" | grep '\<LEADER\>' | awk '{print $2}')

        if [[ "x$leader" != "x" ]]; then
            FE_LEADER=$leader
            QUERY_PORT=$(echo "$memlist" | grep '\<LEADER\>' | awk '{print $2}')
            EDIT_LOG_PORT=$(echo "$memlist" | grep '\<LEADER\>' | awk '{print $5}')
            SELF_ROLE=$(echo "$memlist" | grep "\<$SELF_HOST\>" | awk '{print $6}')
            LEFT_NODE_LSITS=$(echo "$memlist" | grep -v "\<$SELF_HOST\>" | awk '{print $1}' | xargs echo | tr ' ' ',')
            return 0
        fi

        sleep $PROBE_INTERVAL

    done
}

transfer_master() {
    java -jar $STARROCK_HOME/fe/lib/je-7.3.7.jar DbGroupAdmin -helperHosts $SELF_HOST:$EDIT_LOG_PORT -groupName PALO_JOURNAL_GROUP -transferMaster $LEFT_NODE_LSITS 30
}

set_self_not_leader() {
    while true; do
        assignment_variable
        if [[ "x$SELF_ROLE" == "xLEADER" ]]; then
            transfer_master
            local now=$(date +%s)
            let "expire=start+PROBE_LEADER_PODX_TIMEOUT"
            if [[ $expire -le $now ]]; then
                log_stderr "Timed out, abort!"
                exit 1
            fi
            sleep $PROBE_INTERVAL
        else
            return 0
        fi
    done
}

stop_follower_observer() {
    timeout 30 mysql --connect-timeout 2 -h $FE_LEADER -P $QUERY_PORT -u root --skip-column-names --batch -e "ALTER SYSTEM DROP follower '$SELF_HOST:$EDIT_LOG_PORT';"
    $STARROCK_HOME/fe/bin/stop_fe.sh
}

svc_name=$1
if [[ "x$svc_name" == "x" ]]; then
    echo "Need a required parameter!"
    echo "Example: $0 <fe_service_name>"
    exit 1
fi

set_self_not_leader $svc_name
stop_follower_observer

