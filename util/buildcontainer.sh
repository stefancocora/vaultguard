#!/bin/bash

[[ ${DEBUG} == 'true' ]] && set -x


ACTION=${1}
VERSION=${2}
NOCACHE=${3}
IMAGENAME=${4}
DEBUG=${5}

VERSION_PREFIX=${VERSION:0:4}
PROJECT_NAME=`grep "ELF_NAME =" Makefile | awk '{print $3}'`
CWD="${GOPATH}"'/src/github.com/stefancocora/'"${PROJECT_NAME}"

function usage() {

  echo "usage:"
  echo ""
  echo "$0 <action> <version_number> <withcache_or_nochace> <imagename>"
  echo "example:"
  echo "$0 build/rkt v1.2.3 nocache stefancocora/${PROJECT_NAME}"
  echo ""
}

function setupVersion() {

local LOCAL_VERSION=$1
local IMAGENAME=$2
local GIT_COMMIT=""
local GIT_DIRTY=""

GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD)}
if [[ -n "`git status --porcelain`" ]]
then
  GIT_DIRTY="+UNCOMMITEDCHANGES"
else
  GIT_DIRTY=""
fi

if [[ -z ${GIT_DIRTY} ]];
then
  echo "${LOCAL_VERSION}" > VERSION
  echo "${LOCAL_VERSION}-${GIT_COMMIT}" > BUILD
else
  echo "${LOCAL_VERSION}" > VERSION
  echo "${LOCAL_VERSION}-${GIT_COMMIT}${GIT_DIRTY}" > BUILD
fi

}

function build() {
  local LOCAL_VERSION=$1
  local NOCACHE=$2
  local IMAGENAME=$3
  setupVersion $LOCAL_VERSION
  if [[ -z ${NOCACHE} ]] || [[ ${NOCACHE} == 'withcache' ]]; then
    printf "\nwill build ${IMAGENAME} version: ${LOCAL_VERSION} nocache: ${NOCACHE}\n\n"
    docker build -t ${IMAGENAME}:${LOCAL_VERSION} .
  elif [[ ${NOCACHE} == 'nocache' ]]; then
    printf "\nwill build ${IMAGENAME} version: ${LOCAL_VERSION} nocache: ${NOCACHE}\n\n"
    docker build --no-cache --force-rm -t ${IMAGENAME}:${LOCAL_VERSION} .
  fi
  if [[ ${DEBUG} = 'false' ]] || [[ $? -eq 0 ]]
  then
    cleanup
  fi
}

function cleanup() {
  echo "cleaning BUILD and VERSION files ..."
  rm -f BUILD VERSION
}

function exportToRkt() {
  local IMAGE_VERSION=${1}
  local CONTAINER_IMAGE=${2}

  command -v docker2aci >/dev/null 2>&1 || { echo >&2 "The docker2aci cli is required.  Aborting..."; exit 1; }
  command -v rkt >/dev/null 2>&1 || { echo >&2 "The docker2aci cli is required.  Aborting..."; exit 1; }
  command -v actool >/dev/null 2>&1 || { echo >&2 "The docker2aci cli is required.  Aborting..."; exit 1; }

  if [[ ! -d tmp/ ]]
  then
    mkdir tmp/
  fi
  # docker save stefancocora/archlinux-nginx-mainline:v1.11.4-1 -o /tmp/archlinux-nginx-mainline.v1.11.4-1.tar
  cmd_ds="docker save ${CONTAINER_IMAGE}:${IMAGE_VERSION} -o tmp/${CONTAINER_IMAGE:13:20}.${IMAGE_VERSION}.tar"
  echo "==== will run cmd: ${cmd_ds}"
  eval ${cmd_ds}
  # docker2aci /tmp/archlinux-nginx-mainline.v1.11.4-1.tar
  cd tmp/ || exit 1
  cmd_d2a="docker2aci ${CONTAINER_IMAGE:13:20}.${IMAGE_VERSION}.tar"
  echo "=== will run cmd: ${cmd_d2a}"
  eval ${cmd_d2a}
  cd .. || exit 1
  # actool --debug validate stefancocora-archlinux-nginx-mainline-v1.11.4-1.aci
  cmd_av="actool --debug validate tmp/*.aci"
  echo "=== will run cmd: ${cmd_av}"
  eval ${cmd_av}
  # rkt fetch --insecure-options=image /tmp/stefancocora-archlinux-nginx-mainline-v1.11.4-1.aci
  cmd_rf="rkt fetch --insecure-options=image tmp/*.aci"
  echo "=== will run cmd: ${cmd_rf}"
  eval ${cmd_rf}
  rm -rf tmp/



}

function runRktInteractive(){
  local IMAGE_VERSION=${1}
  local CONTAINER_IMAGE=${2}
  if [[ $3 == 'true' ]]
  then
    local DEBUG=true
  else
    local DEBUG=false
  fi


  # test the image
# sudo rkt --stage1-path=/usr/lib/rkt/stage1-images/stage1-fly.aci --debug=true run --interactive stefancocora/vaultguard:v0.0.1 --net=host --dns 8.8.8.8 --exec /bin/bash
echo "=== will run rkt interactively"
rkt_cmd="sudo rkt --insecure-options=image run \
     --interactive \
     --debug=${DEBUG} \
     --volume src,kind=host,source=${CWD},readOnly=true \
     --volume elf,kind=host,source=${GOPATH}/bin/${PROJECT_NAME},readOnly=true \
     --volume awscreds,kind=host,source=${HOME}/.aws/credentials,readOnly=true \
     --net=host \
     --dns=10.110.110.1 \
     ${CONTAINER_IMAGE}:${IMAGE_VERSION} \
     --mount volume=awscreds,target=/home/${PROJECT_NAME}/.aws/config \
     --mount volume=src,target=/src \
     --mount volume=elf,target=/${PROJECT_NAME} \
     --exec /bin/bash"
echo "will run ${rkt_cmd}"

# --net=none \
sudo rkt --insecure-options=image run \
--interactive \
--debug=${DEBUG} \
--volume src,kind=host,source=${CWD},readOnly=true \
--volume elf,kind=host,source=${GOPATH}/bin/${PROJECT_NAME},readOnly=true \
--volume awscreds,kind=host,source=${HOME}/.aws/credentials,readOnly=true \
--net=host \
--dns=10.110.110.1 \
${CONTAINER_IMAGE}:${IMAGE_VERSION} \
--mount volume=awscreds,target=/home/${PROJECT_NAME}/.aws/config \
--mount volume=src,target=/src \
--mount volume=elf,target=/${PROJECT_NAME} \
--exec /bin/bash

}

case ${ACTION} in
  build)
  case ${VERSION} in
      ${VERSION_PREFIX}.*)

      if [[ ${DEBUG} = 'true' ]]
      then
        echo "debug mode: "
        echo "version: ${VERSION}"
        echo "nocache: ${NOCACHE}"
        echo "imagename: ${IMAGENAME}"
        echo "debug: ${DEBUG}"
        echo "VERSION_PREFIX: ${VERSION_PREFIX}"
      fi

      build ${VERSION} ${NOCACHE} ${IMAGENAME}
      shift
      ;;
  esac
  shift
  ;;
  rkt)
  if [[ ${DEBUG} = 'true' ]]
  then
    echo "debug mode: "
    echo "action: ${ACTION}"
    echo "version: ${VERSION}"
    echo "imagename: ${IMAGENAME}"
    echo "debug: ${DEBUG}"
    echo "VERSION_PREFIX: ${VERSION_PREFIX}"
  fi

  exportToRkt ${VERSION} ${IMAGENAME}
  shift
  ;;
  rktinteractive)
  if [[ ${DEBUG} = 'true' ]]
  then
    echo "debug mode: "
    echo "action: ${ACTION}"
    echo "version: ${VERSION}"
    echo "imagename: ${IMAGENAME}"
    echo "debug: ${DEBUG}"
    echo "VERSION_PREFIX: ${VERSION_PREFIX}"
  fi

  runRktInteractive ${VERSION} ${IMAGENAME} ${DEBUG}
  shift
  ;;
  -h)
  usage
  exit 1
  shift
  ;;
  --help)
  usage
  exit 1
  shift
  ;;
  *)

  if [[ ${DEBUG} = 'true' ]]
  then
    echo "debug mode: "
    echo "action: ${ACTION}"
    echo "version: ${VERSION}"
    echo "nocache: ${NOCACHE}"
    echo "imagename: ${IMAGENAME}"
    echo "debug: ${DEBUG}"
    echo "VERSION_PREFIX: ${VERSION_PREFIX}"
  fi

  usage
  exit 1
  shift
  ;;
esac
