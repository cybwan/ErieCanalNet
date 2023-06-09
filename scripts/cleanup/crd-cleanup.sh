#!/bin/bash

# This script is only for cleaning up CRDs if performing an ECNET upgrade that includes
# CRD updates. This will delete existing CRDs and Custom Resources.

# shellcheck disable=SC1091

kubectl delete --ignore-not-found --recursive -f ./charts/ecnet/crds/
