#!/bin/bash

# Parameters: <address> <from-date> [<extrargs>, ...]
if [ "$#" -lt 2 ]; then
    echo "This script is used to show the connectivity status of a gateway from a specific date."
    echo "This perform calls to get bstates events pushed by the specified address." 
    echo "Usage: $0 <address> <from-date> [<extrargs>, ...]"
    echo "Example: $0 61abdc949d199f2f \"2024-05-30 10:07:00 +0200 CEST\" "
    exit 1
fi

# Optional parameters
extrargs=""
if [ "$#" -gt 2 ]; then
    extrargs="${@:3}"
fi

address=$1
date=$2

idefix event get bstates -a ${address} --since="${date}" --timeout 60s --meta-filter {\"source_type\":\"gateway\"} ${extrargs} \
    --field-alias BOOT_COUNTER=00_BC \
    --field-alias EVENT_COUNTER=01_EC \
    --field-alias CSQ=02_CSQ \
    --field-alias OPERATOR=03_ACT \
    --field-alias ACT=04_ACT \
    --field-alias CPSI_SYSTEMMODE=05_SystemMode \
    --field-alias OPERATOR=06_OPERATOR \
    --field-alias CREG=07_CREG \
    --field-alias CPSI_OPMODE=08_OperationMode \
    --field-alias CPSI_MCC=09_MCC \
    --field-alias CPSI_MNC=10_MNC \
    --field-alias CPSI_CELLID=11_CELLID \
    --field-alias CPSI_PCELLID=12_PCELLID \
    --field-alias CPSI_SCELLID=13_SCELLID \
    --field-alias CPSI_LAC_TAC=14_LAC_TAC \
    --field-alias ICC=15_ICC \
    --field-alias IMSI=16_IMSI \
    --field-alias CPSI=17_CPSI \
    --field-alias PPP=18_PPP \
    --field-alias CLOUD=19_CLOUD \
    --field-match '^(CPSI$|CPSI_(MNC|MCC|LAC_TAC|OPMODE|SYSTEMMODE|OPERATOR|CELLID|PCELLID|SCELLID)$|BOOT_COUNTER|EVENT_COUNTER|PRIORITY_EVENT|NOTHING_TO_SEND|CSQ|CREG$|ACT$|OPERATOR$|PPP|CLOUD|TUN|ICC$|IMSI$|IMEI$)' 

