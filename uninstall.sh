#!/bin/bash
set -e

function cleanup() {
  set +e
}

function set_variable_from_command() {
  if [[ ($# -ne 3) ]];
  then
    echo "Internal Error - Wrong number of parameters"
    exit 1
  fi
  VAR_NAME=$1
  DEFAULT_COMMAND=$2
  if [ -z "${!VAR_NAME}" ]
  then
      result=$(bash -c "${!DEFAULT_COMMAND}") || {
        ret=$?; echo "$3 - Code: $ret"; return $ret;
      }
      declare -gx "$VAR_NAME"=$result
  fi
  echo "[PRE-UNINSTALL]: $VAR_NAME is set to: ${!VAR_NAME}"
}

function print_help() {
   echo "Arguments:"
   echo "   --help: print this help"
   echo "   --test: source file only for testing"
}

if [[ ($# -eq 1 && $1 == '--help') ]];
then
  print_help
  exit 0
# The next line is required to easily unit-test the functions previously declared
elif [[ $# -eq 1 && $0 == '/opt/bats/libexec/bats-core/bats-exec-test' ]]
then
  echo "Testing..."
  return 0
elif [[ $# -ge 1 ]]
then
  echo "ERROR: Illegal parameters"
  print_help
  exit 1
fi


trap cleanup EXIT
URL=https://github.com/LiqoTech/liqo.git
HELM_VERSION=v3.2.3
HELM_ARCHIVE=helm-${HELM_VERSION}-linux-amd64.tar.gz
HELM_URL=https://get.helm.sh/$HELM_ARCHIVE
NAMESPACE_DEFAULT="liqo"

# Necessary Commands
commands="curl kubectl"

echo "[PRE-UNINSTALL]: Checking all pre-requisites are met"
for val in $commands; do
  if command -v $val > /dev/null; then
    echo "[PRE-UNINSTALL]: $val correctly found"
  else
    echo "[PRE-UNINSTALL] [FATAL] : $val not found. Exiting"
    exit 1
  fi
done

TMPDIR=$(mktemp -d)
mkdir -p $TMPDIR/bin/
echo "[PRE-UNINSTALL] [HELM] Checking HELM installation..."
echo "[PRE-UNINSTALL] [HELM]: Downloading Helm $HELM_VERSION"
curl --fail -L ${HELM_URL} | tar zxf - --directory="$TMPDIR/bin/" --wildcards '*/helm' --strip 1
if [ "$LIQO_SUFFIX" == "-ci" ] && [ ! -z "${LIQO_VERSION}" ]  ; then
  git clone "$URL" "$TMPDIR"/liqo
  cd "$TMPDIR"/liqo
  git checkout "$LIQO_VERSION" > /dev/null 2> /dev/null
  cd -
else
  git clone "$URL" "$TMPDIR"/liqo --depth 1
fi

NAMESPACE_COMMAND="echo $NAMESPACE_DEFAULT"
set_variable_from_command NAMESPACE NAMESPACE_COMMAND "[ERROR]: Error while creating the namespace... "

$TMPDIR/bin/helm uninstall liqo -n $NAMESPACE
echo "[UNINSTALL]: Uninstalling LIQO on your cluster..."
sleep 30
kubectl delete ns $NAMESPACE

kubectl delete MutatingWebhookConfiguration mutatepodtoleration
kubectl delete ValidatingWebhookConfiguration peering-request-operator

kubectl delete csr "peering-request-operator.$NAMESPACE" > /dev/null 2> /dev/null
kubectl delete csr "mutatepodtoleration.$NAMESPACE" > /dev/null 2> /dev/null
