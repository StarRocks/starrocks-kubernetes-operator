#!/bin/bash

MY_SELF=
PROBE_TIMEOUT=60
PROBE_INTERVAL=2
FE_QUERY_PORT=${FE_QUERY_PORT:-9030}
HOST_TYPE=${HOST_TYPE:-"IP"}

log_stderr()
{
    echo "[`date`] $@" >&2
}


show_compute_nodes()
{
    local svc=$1
    timeout 15 mysql --connec-timeout 2 -h $svc -P $FE_QUERY_PORT -u root --skip-column-names --batch -e 'show compute nodes;'
}

drop_my_self()
{
    local svc=$1
    local start=`date +%s`
    local memlist=
    local register_self=
    if [[ "$HOST_TYPE" == "IP" ]] ; then
        MY_SELF=`hostname -i`
    else
        MY_SELF=`hostname -f`
    fi

    while true
    do
        memlist=`show_compute_nodes $svc`
        register_self=`echo "$memlist" | grep '\<$MY_SELF\>'` | awk '{print $2}'`
        if [[ "x$register_self" != "x" ]] ; then
            log_stderr "drop my self $register_self"
            timeout 15 mysql --connect-timeout 2 -h $svc -P $QUERY_PORT -u root --skip-column-names --batch -e << EOF
alter system drop compute node $register_self;
EOF
            return 0
        fi

        local now=`date +%s`
        let "expire=start+PROBE_TIMEOUT"
        if [[ $expire -le $now ]] ; then
            log_stderr "Timed out, abort!"
            exit 1
        fi
        sleep $PROBE_INTERVAL 
    done
}

svc_name=$1
if [[ "x$svc_name" == "x" ]] ; then
    echo "Need a required parameter!"
    echo "  Example: $0 <fe_service_name>"
    exit 1
fi


drop_my_self $svc_name
 $STARROCK_HOME/be/bin/stop_cn.sh
