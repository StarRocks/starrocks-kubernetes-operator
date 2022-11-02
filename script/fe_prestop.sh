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
STARROCKS_ROOT=${STARROCKS_HOME:-"/opt/starrocks"}
STARROCKS_HOME=$STARROCKS_ROOT/fe

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
        local leader_line=$(echo "$memlist" | grep '\<LEADER\>')

        if [[ "x$leader_line" != "x" ]]; then
            # | Name | IP | EditLogPort | HttpPort | QueryPort | RpcPort | Role |
            FE_LEADER=$(echo "$leader_line" | awk '{print $2}')
            EDIT_LOG_PORT=$(echo "$memlist" | grep '\<LEADER\>' | awk '{print $3}')
            QUERY_PORT=$(echo "$memlist" | grep '\<LEADER\>' | awk '{print $5}')
            SELF_ROLE=$(echo "$memlist" | grep "\<$SELF_HOST\>" | awk '{print $7}')
            LEFT_NODE_LSITS=$(echo "$memlist" | grep -v "\<$SELF_HOST\>" | awk '{print $1}' | xargs echo | tr ' ' ',')
            return 0
        fi

        sleep $PROBE_INTERVAL
    done
}

transfer_master() {
    java -jar $STARROCKS_HOME/lib/je-7.3.7.jar DbGroupAdmin -helperHosts $SELF_HOST:$EDIT_LOG_PORT -groupName PALO_JOURNAL_GROUP -transferMaster $LEFT_NODE_LSITS 30
}

set_self_not_leader() {
    local svc=$1
    local start=$(date +%s)
    while true; do
        assignment_variable $svc
        if [[ "x$SELF_ROLE" == "xLEADER" ]]; then
            if [[ "x$LEFT_NODE_LSITS" == "x" ]] ; then
                log_stderr "No other nodes left. Can't do master switch ..."
                exit 1
            fi
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

drop_follower_observer() {
    while true
    do
        timeout 30 mysql --connect-timeout 2 -h $FE_LEADER -P $QUERY_PORT -u root --skip-column-names --batch -e "ALTER SYSTEM DROP follower '$SELF_HOST:$EDIT_LOG_PORT';"
        local memlist=`show_frontends $FE_LEADER`
        if [[ -n "$memlist" ]] ; then
            if ! echo "$memlist" | grep -q -w "$SELF_HOST" &>/dev/null ; then
                # can't find myself from `show_frontends` any more
                break;
            else
                log_stderr "Can still find myself from 'show_frontends' output ..."
                sleep $PROBE_INTERVAL
            fi
        fi
    done
}

svc_env_var_name=${1:-"FE_SERVICE_NAME"}
svc_name=${!svc_env_var_name}
if [[ "x$svc_name" == "x" ]]; then
    log_stderr "Need a required parameter!"
    log_stderr "Example: $0 <fe_service_env_var_name>"
    exit 1
fi

set_self_not_leader $svc_name
drop_follower_observer
$STARROCKS_HOME/bin/stop_fe.sh
