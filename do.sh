#!/usr/bin/env bash

if [[ -z ${MYSQL_USER} ]]; then
    source ./docker/dev/docker.env
fi

# https://stackoverflow.com/a/21189044 with some updates to handle db_names
function parse_yaml {
   local prefix=$2
   local s='[[:space:]]*' w='[a-zA-Z0-9_]*' fs=$(echo @|tr @ '\034')
   sed -ne "s|^\($s\):|\1|" \
        -e "s|^\($s\)\($w\)$s:$s[\"']\(.*\)[\"']$s\$|\1$fs\2$fs\3|p" \
        -e "s|^\($s\)\($w\)$s:$s\(.*\)$s\$|\1$fs\2$fs\3|p"  $1 |
   awk -F$fs '{
      indent = length($1)/2;
      vname[indent] = $2;
      for (i in vname) {if (i > indent) {delete vname[i]}}
      if (length($3) > 0) {
         vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
         printf("%s%s%s=\"%s\"\n", "'$prefix'",vn, $2, $3);
      }
      if (indent == 0) {
        if(length(db_names) == 0) {
           db_names=$2 
        } else {
           db_names=db_names"|"$2
        }
      }
   } END {printf("db_names=\"%s\"\n",db_names)}'
}

eval $(parse_yaml config.yml)

validate() {
    IFS="|"
    for db in $db_names; do
        if [ "$db" == "$1" ]; then
            return 0 
        fi
    done
    return 1
}


# setting up ssh preferences. (optional) 
stablessh() {
    sudo /sbin/sysctl -w net.ipv4.tcp_keepalive_time=60 net.ipv4.tcp_keepalive_intvl=60 net.ipv4.tcp_keepalive_probes=5
    sudo sed -i '1{p;s/.*/TCPKeepAlive yes/;h;d;};/^TCPKeepAlive/{g;p;s/.*//;h;d;};$G' /etc/ssh/sshd_config
    sudo sed -i '1{p;s/.*/ClientAliveInterval 60/;h;d;};/^ClientAliveInterval/{g;p;s/.*//;h;d;};$G' /etc/ssh/sshd_config
    sudo sed -i '1{p;s/.*/ClientAliveCountMax 120/;h;d;};/^ClientAliveCountMax/{g;p;s/.*//;h;d;};$G' /etc/ssh/sshd_config
}

dbconn() {
    if [[ $1 ]]; then 
        if ! validate $1 ; then 
            echo "Invalid db name '$1' please choose out of : [${db_names}]" 1>&2
            exit 1
        fi
        db_names=$1
    else
        disconnect
    fi
    IFS="|"
    for db in $db_names; do
        project=${db}_project
        port=${db}_port
        ip=${db}_ip
        tmux new -d "gcloud compute --project ${!project} ssh bastion-01 \
            --ssh-flag='-o TCPKeepAlive=yes' --ssh-flag='-o ServerAliveInterval=60' --ssh-flag='-o ServerAliveCountMax=120' --ssh-flag='-n' --ssh-flag='-N' --ssh-flag='-o ExitOnForwardFailure=yes' \
            --zone 'asia-southeast1-a' -- -p 22 -L ${!port}:${!ip} > /dev/null 2>&1"
    done

    # netcat to check connection, print in purple.
    # Added Seperate loop to allow above loop spawn background tmux instances. 
    for db in $db_names; do
        port=${db}_port
        ip=${db}_ip
        until nc -z -v -w20 127.0.0.1 ${!port} 2> >(while read line; do echo -e "\e[01;35m$line\e[0m" >&2; done)
            do
                printf "\033[0;33mWaiting mysql connection for ${!port}:${!ip} to be ready...\033[0m\n"
                sleep 2
            done
    done
}

disconnect() {
    pkill -f tmux && sleep 0.1
}

runOnServer() {
    if ! validate $1 ; then 
        echo "Invalid env name '$1' please choose out of : [${db_names}]" 1>&2
        exit 1
    fi
    pass=$1_p
    user=$1_u
    port=$1_port
    cat ${@:3} | mysql -u ${!user} -p${!pass} -h 127.0.0.1 -P ${!port} -D $2 -A -v -v --comments --default-character-set=utf8
    if [[ $? -ne 0 ]]; then
        echo "MYSQL query failed" 1>&2
        exit 1
    fi
}

runscript() {
    read -e -p "Enter filename: " f 
    filepath=$(echo $f | xargs)
    runOnServer $1 $2 $filepath
}

run() {
    if ! validate $1 ; then 
        echo "Invalid env name '$1' please choose out of : [${db_names}]" 1>&2
        exit 1
    fi
    pass=$1_p
    user=$1_u
    port=$1_port
    connection=$(echo "${!user}:${!pass}@tcp(127.0.0.1:${!port})/$2?charset=utf8mb4&parseTime=true" | tr -d '\r')
    go run main.go --connection="$connection" ${@:3}
}


runlocal() {
    eval $(cat .env)
    connection=$(echo "${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(127.0.0.1:${MYSQL_PORT})/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=true" | tr -d '\r')
    go run main.go --connection="$connection" ${@:1}
}

xo() {
    # https://stackoverflow.com/questions/58403134/go-permission-denied-when-trying-to-create-a-file-in-a-newly-created-directory
    rm -rf xo_gen
    mkdir xo_gen xo_gen/enum xo_gen/table xo_gen/repo xo_gen/xo_wire
    chmod 0777 -R xo_gen xo_gen/enum xo_gen/table xo_gen/repo xo_gen/xo_wire

    eval $(cat .env)
    connection=$(echo "mysql://${MYSQL_USER}:${MYSQL_PASSWORD}@127.0.0.1:${MYSQL_PORT}/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=true" | tr -d '\r')
    go run ./tools/xo/main.go --connection="$connection"
    # go run github.com/tinhtran24/xo-patcher/xo --connection="$connection"
}

if [[ $1 = 'run' ]]; then
    run $2 $3 ${@:4}
elif [[ $1 = 'runLocal' ]]; then
    runlocal ${@:2}
elif [[ $1 = 'wire' ]]; then
    go get -d github.com/google/wire/cmd/wire@65ae46b7eaa1
    go run github.com/google/wire/cmd/wire gen xo-patcher/wire_app
elif [[ $1 = 'goimports' ]]; then
      go get golang.org/x/tools/cmd/goimports
    ~/go/bin/go-fanout --command="goimports -w" --chunk=5 -- xo_gen/enum/*
    ~/go/bin/go-fanout --command="goimports -w" --chunk=5 -- xo_gen/table/*
    ~/go/bin/go-fanout --command="goimports -w" --chunk=5 -- xo_gen/repo/*
    ~/go/bin/go-fanout --command="goimports -w" --chunk=5 -- xo_gen/xo_wire/*
elif [[ $1 = 'xo' ]]; then
    xo 
elif [[ $1 = 'stablessh' ]]; then
    stablessh
elif [[ $1 = 'dbconn' ]]; then
    dbconn $2
elif [[ $1 = 'disconnect' ]]; then
    disconnect
elif [[ $1 = 'runscript' ]]; then
    runscript $2 $3 
else
    echo "No usage found"
fi


# go install github.com/google/wire/cmd/wire@65ae46b7eaa146e99673e290251ea26f28139362
# if you install Above version of wire, it's old version which does not need to check pointer.
# /home/ketan/go/bin/wire ./wire_app
