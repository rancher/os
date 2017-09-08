#!/bin/bash
usage()
{
   echo "Usage: $0 <url_to_schema> <folder_to_save_generated_files>"
   echo "Usage: Do not use folder name as 'client' or 'v2'"
   exit 1
}

if [ "$#" -ne 2 ]
then
  usage
fi

if [ "$2" == "client" ] || [ "$2" == "v2" ]
then
  usage
fi

set -e -x

cd $(dirname $0)/../generator

source $(dirname "$0")/../scripts/common_functions

gen $1 $2 rename

echo Success
