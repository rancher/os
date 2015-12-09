#!/bin/bash
set -e

cd $(dirname $0)/..

package=$(go list)
prefix="${package}/vendor"
require="github.com/stretchr/testify/require" # the only test import

imports=(  )
importsLen=${#imports[@]}

collectImports() {
  imports=( $(GOOS=linux go list -f '{{join .Deps "\n"}}' | egrep "^${prefix}/" | sed s%"^${package}.*/vendor/"%./vendor/%) )
  imports=(
    "${imports[@]}" "./vendor/${require}"
    $(GOOS=linux go list -f '{{join .Deps "\n"}}' "${prefix}/${require}" | egrep "^${prefix}/" | sed s%"^${package}.*/vendor/"%./vendor/%)
  )
  echo importsLen: $importsLen
  echo collected imports: ${#imports[@]}
}

nonImports() {
  while read path; do
    skip=0
    for i in "${imports[@]}"; do
      [[ "${i}" == "${path}" || ${i} = ${path}/* ]] && skip=1 && break
    done
    [ "$skip" == "0" ] && echo ${path}
  done
}

collectImports

while [ ${#imports[@]} != ${importsLen} ]; do
  importsLen=${#imports[@]}
  echo '=====> Collected imports'
  for i in "${imports[@]}"; do
    echo ${i}
  done

  echo '=====> Removing unused packages'
  find  ./vendor -type d | nonImports | xargs -I{} rm -rf {}

  echo '=====> Removing empty dirs'
  emptyDirs=( $(find ./vendor -type d -empty) )
  while [ ${#emptyDirs[@]} -gt 0 ]; do
    rmdir ${emptyDirs[@]}
    emptyDirs=( $(find ./vendor -type d -empty) )
  done

  collectImports
done

echo '=====> Done!'
