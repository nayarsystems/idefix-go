#!/bin/bash

# Parameters: <address> <from-date>
if [ "$#" -ne 2 ]; then
    echo "This script is used to show the connectivity status of a gateway from a specific date."
    echo "This perform calls to get bstates events pushed by the specified address." 
    echo "Usage: $0 <address> <from-date>"
    echo "Example: $0 61abdc949d199f2f \"2024-05-30 10:07:00 +0200 CEST\" "
    exit 1
fi

address=$1
date=$2

idefix event get bstates -a ${address} --since="${date}" --timeout 60s --meta-filter {\"source_type\":\"gateway\"} --field-align --continue \
    --field-alias BOOT_COUNTER=00_BC \
    --field-alias EVENT_COUNTER=01_EC \
    --field-alias CSQ=02_CSQ \
    --field-alias ACT=03_ACT \
    --field-alias CPSI_SYSTEMMODE=04_SystemMode \
    --field-alias OPERATOR=05_OPERATOR \
    --field-alias CREG=06_CREG \
    --field-alias CPSI_OPMODE=07_OperationMode \
    --field-alias CPSI_MCC=08_MMC \
    --field-alias CPSI_MNC=09_MMC \
    --field-alias CPSI_CELLID=10_CELLID \
    --field-alias CPSI_PCELLID=11_PCELLID \
    --field-alias CPSI_SCELLID=12_SCELLID \
    --field-alias CPSI_LAC_TAC=13_LAC_TAC \
    --field-alias ICC=14_ICC \
    --field-alias CPSI=15_CPSI \
    --field-alias PPP=16_PPP \
    --field-alias CLOUD=17_CLOUD \
    --field-match '^(CPSI$|CPSI_(MNC|MCC|LAC_TAC|OPMODE|SYSTEMMODE|OPERATOR|CELLID|PCELLID|SCELLID)$|BOOT_COUNTER|EVENT_COUNTER|PRIORITY_EVENT|NOTHING_TO_SEND|CSQ|CREG$|ACT$|PPP|CLOUD|TUN|ICC$|IMEI$)' 

