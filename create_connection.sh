#!/bin/bash
set -eu

cd `dirname $0`

PROFILE=""
PROFILE_OPTION=""
REGION="ap-northeast-1"
CONNECTION_NAME=""
PROVIDER_TYPE="GITHUB"

while getopts p:c: OPT; do
	case $OPT in
		p)
			PROFILE="$OPTARG"
			;;
		c)
			CONNECTION_NAME="$OPTARG"
			;;
	esac
done

if [ -z "${CONNECTION_NAME}" ]; then
	echo "CONNECTION_NAME (-c) is required"
	exit 1
fi

if [ -n "${PROFILE}" ]; then
	PROFILE_OPTION="--profile ${PROFILE}"
fi


aws apprunner create-connection --connection-name ${CONNECTION_NAME} --provider-type ${PROVIDER_TYPE} ${PROFILE_OPTION}