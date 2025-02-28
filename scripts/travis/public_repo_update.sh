#!/usr/bin/env bash

set -e

if [ -z "$1" ]; then
    echo "Error: script argument VERSION is not specified."
    exit 1
fi
VERSION=$1

if [ -z "${PUBLIC_GITHUB_TOKEN}" ]; then
    echo "Error: PUBLIC_GITHUB_TOKEN environment variable is not set"
    exit 1
fi
TC_PUBLIC_REPO=turbonomic-container-platform

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SRC_DIR=${SCRIPT_DIR}/../../deploy
OUTPUT_DIR=${SCRIPT_DIR}/../../_output
HELM=${SCRIPT_DIR}/../../bin/helm
TARGET=data-ingestion-framework

if ! command -v ${HELM} > /dev/null 2>&1; then
    HELM=helm
    if ! command -v helm > /dev/null 2>&1; then
        echo "Error: helm could not be found."
        exit 1
    fi
fi

if ! command -v git > /dev/null 2>&1; then
    echo "Error: git could not be found."
    exit 1
fi

echo "===> Cloning public repo..."; 
mkdir ${OUTPUT_DIR}
cd ${OUTPUT_DIR}
git clone https://${PUBLIC_GITHUB_TOKEN}@github.com/IBM/${TC_PUBLIC_REPO}.git
cd ${TC_PUBLIC_REPO}

echo "===> Create folders"
rm -rf ${TARGET}
mkdir -p ${TARGET}/examples
cd ${TARGET}

# copy examples files
echo "===> Copy examples files"
cp -r ${SRC_DIR}/../example/* examples/

# Insert current version
echo "===> Updating TurboDIF version in yaml files"
find ./ -type f -name '*.y*' -exec sed -i.bak "s|<TURBODIF_VERSION>|${VERSION}|g" {} +
find ./ -name '*.bak' -type f -delete

# Add README
echo "===> Adding README"
echo "See the [documentation](https://www.ibm.com/docs/en/tarm/latest?topic=configuration-data-ingestion-framework)" > README.md

# commit all modified source files to the public repo
echo "===> Commit modified files to public repo"
cd .. 
git add .
if ! git diff --quiet --cached; then
    git commit -m "${TARGET} ${VERSION}"
    git push
else
    echo "No changed files"
fi

# cleanup
rm -rf ${OUTPUT_DIR}

echo ""
echo "Update public repo complete."
