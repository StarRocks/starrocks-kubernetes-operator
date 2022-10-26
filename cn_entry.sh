#!/bin/bash

#1. find fe leader
#2. ad cn (myself) to fe

HOST_TYPE=${HOST_TYPE:-"IP"}
FE_QUERY_PORT=${FE_QUERY_PORT:-9030}
PROBE_LEADER_POD0_TIMEOUT=60
PROBE_INTERVAL=2
FE_LEADER=
HEARTBEAT_PORT=9050
MY_SELF=
MY_IP=`hostname -i`
MY_HOSTNAME=`hostname -f`
STARROCK_HOME=${STARROCK_HOME:-"/opt/starrocks"}

log_stderr()
{
    echo "[`date`] $@" >&2
}

show_frontends(){
        timeout 15 mysql --connec-timeout 2 -h $svc -P $FE_QUERY_PORT -u root --skip-column-names --batch -e 'show frontends;'
}

parse_confval_from_fe_conf()
{
    # a naive script to grep given confkey from fe conf file
    # assume conf format: ^\s*<key>\s*=\s*<value>\s*$
    local confkey=$1
    local confvalue=`grep "\<$confkey\>" $FE_CONFFILE | grep -v '^\s*#' | sed 's|^\s*'$confkey'\s*=\s*\(.*\)\s*$|\1|g'`
    echo "$confvalue"
}

collect_env_info()
{
    local fe_conffile=$STARROCK_HOME/fe/conf/fe.conf

    # heartbeat_port from conf file
    local heartbeat_port=`parse_confval_from_fe_conf "heartbeat_service_port"`
    if [[ "x$heartbeat_port" != "x" ]] ; then
        HEARTBEAT_PORT=$heartbeat_port
    fi

}

find_fe_leader() {
    local svc=$1
    local memlist=
    local leader=
    local start=`date +%s`
    while true
    do 
        memlist=`show_frontends $svc`
        leader=`echo "$memlist" | grep '\<LEADER\>' | awk '{print $2}'`
        if [[ "x$leader" != "x" ]]; then
            log_stderr "Find leader: $leader!"
            FE_LEADER=$leader
            return 0
        fi

        # no leader yet, check if needs timeout and quit
        log_stderr "No leader yet, has_member: $has_member ..."
        local now=`date +%s`
        let "expire=start+PROBE_LEADER_PODX_TIMEOUT"
        if [[ $expire -le $now ]] ; then 
            log_stderr "Timed out, abort!"
            exit 1
        fi

        sleep $PROBE_INTERVAL
    done
}
add_self_and_start()
{
    collect_env_info
    if [[ "x$HOST_TYPE" == "xIP"]] ; then
        MY_SELF=$MY_IP
    else
        MY_SELF=$MY_HOSTNAME
    fi


    timeout 15 mysql --connect-timeout 2 -h $FE_LEADER -P $QUERY_PORT -u root  -skip-column-names --batch << EOF
ALTER SYSTEM ADD COMPUTE NODE "$MY_SEFL:$HEARTBEAT_PORT"
EOF
    log_stderr "run start_cn.sh"
    $STARROCK_HOME/be/bin/start_cn.sh
}

svc_name=$1
if [[ "x$svc_name" == "x" ]] ; then
    echo "Need a required parameter!"
    echo "  Example: $0 <fe_service_name>"
    exit 1
fi

find_fe_leader $svc_name
add_sefl_and_start
