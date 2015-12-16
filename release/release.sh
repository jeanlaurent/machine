#!/bin/bash

# USAGE :
# GITHUB_TOKEN=XXXXX release/release.sh 0.5.x

#Put your github username here, while testing performing new releases
GITHUB_USER=jeanlaurent
#GITHUB_USER=docker
GITHUB_REPO=machine


function display {
  echo "🐳  $1"
  echo ""
}

function checkError {
  if [[ "$?" -ne 0 ]]; then
    echo "😡   $1"
    exit 1
  fi
}

VERSION=$1
GITHUB_VERSION="v${VERSION}"
PROJECT_URL="git@github.com:${GITHUB_USER}/${GITHUB_REPO}"

RELEASE_DIR="$(git rev-parse --show-toplevel)/../release-${VERSION}"
DOCKER_MACHINE="${GOPATH}/src/github.com/docker/machine/bin/docker-machine"
GITHUB_RELEASE_FILE="github-release-${VERSION}.md"


if [[ -z "${VERSION}" ]]; then
  #TODO: Check version is well formed
  echo "Missing version argument"
  exit 1
fi

if [[ -z "${GITHUB_TOKEN}" ]]; then
  echo "GITHUB_TOKEN missing"
  exit 1
fi

if [[ ! -f ${DOCKER_MACHINE} ]]; then
  echo "You must have a fresh build of docker-machine here : ${DOCKER_MACHINE}"
  exit 1
fi

command -v git > /dev/null 2>&1
checkError "You obviously need git, please consider installing it..."

command -v github-release > /dev/null 2>&1
checkError "github-release is not installed, go get -u github.com/aktau/github-release or check https://github.com/aktau/github-release, aborting."

command -v openssl > /dev/null 2>&1
checkError "You need openssl to generate binaries signature, brew install it, aborting."

display "Starting release ${VERSION} on ${PROJECT_URL} with token ${GITHUB_TOKEN}"

if [[ -d "${RELEASE_DIR}" ]]; then
  display "Cleaning up ${RELEASE_DIR}"
  rm -rdf "${RELEASE_DIR}"
  checkError "Can't clean up."
fi

display "Cloning into ${RELEASE_DIR} from ${PROJECT_URL}"

mkdir -p "${RELEASE_DIR}"
checkError "Can't create ${RELEASE_DIR}, aborting."
git clone -q "${PROJECT_URL}" "${RELEASE_DIR}"
checkError "Can't clone into ${RELEASE_DIR}, aborting."

cd "${RELEASE_DIR}"

display "Bump version number to ${VERSION}"
sed -i.bak s/"${VERSION}-dev"/"${VERSION}"/g version/version.go
checkError "Sed borkage..., aborting."

git add version/version.go
git commit -q -m"Bump version to ${VERSION}" -s
checkError "Can't git commit the version upgrade, aborting."
rm version/version.go.bak

display "Checking machine 'release' exist"
${DOCKER_MACHINE} ip release > /dev/null 2>&1
if [[ "$?" -ne 0 ]]; then
  display "machine 'release' does not exist, creating it"
  ${DOCKER_MACHINE} rm -f release 2> /dev/null
  ${DOCKER_MACHINE} create -d virtualbox release
fi
eval $("${DOCKER_MACHINE}" env release)
checkError "Machine 'release' is in a weird state, aborting."

display "Building in-container style"
USE_CONTAINER=1 make build-x
checkError "Build error, aborting."

# this is temporary -> Remove me once merged
mkdir release
cp ../machine/release/github-release-template.md release/github-release-template.md

display "Generating github release"
cp -f release/github-release-template.md "${GITHUB_RELEASE_FILE}"
checkError "Can't find github release template"
CONTRIBUTORS=$(git log "${GITHUB_VERSION}".. --format="%aN" --reverse | sort | uniq | awk '{printf "- %s\n", $0 }')
CHANGELOG=$(git log "${GITHUB_VERSION}".. --oneline)

CHECKSUM=""
cd bin/
for file in $(ls docker-machine*); do
  SHA256=$(openssl dgst -sha256 < "${file}")
  MD5=$(openssl dgst -md5 < "${file}")
  LINE=$(printf "\n * **%s**\n  * sha256 \`%s\`\n  * md5 \`%s\`\n\n" "${file}" "${SHA256}" "${MD5}")
  CHECKSUM="${CHECKSUM}${LINE}"
done
cd ..

sed -i.bak s/{{VERSION}}/"${GITHUB_VERSION}"/g "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${GITHUB_VERSION} ]"
sed -i.bak s/{{CHANGELOG}}/"${CHANGELOG}"/g "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CHANGELOG} ]"
sed -i.bak s/{{CONTRIBUTORS}}/"${CONTRIBUTORS}"/g "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CONTRIBUTORS} ]"
TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{CHECKSUM\}\}/$CHECKSUM}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CHECKSUM} ]"
#rm "${GITHUB_RELEASE_FILE}".bak # needs to be removed

RELEASE_DOCUMENTATION="$(cat ${GITHUB_RELEASE_FILE})"

display "Tagging and pushing tags"
git remote | grep remote.prod.url
if [[ "$?" -ne 0 ]]; then
  git remote add remote.prod.url "${PROJECT_URL}"
fi

git ls-remote --tags 2> /dev/null | grep "${GITHUB_VERSION}"
if [[ "$?" -ne 0 ]]; then
  git tag -d "${GITHUB_VERSION}"
  git push origin :refs/tags/"${GITHUB_VERSION}"
fi

display "Tagging release on github"
git tag "${GITHUB_VERSION}"
git push remote.prod.url "${GITHUB_VERSION}"

display "Checking if release already exists"
github-release info \
    --security-token  "${GITHUB_TOKEN}" \
    --user "${GITHUB_USER}" \
    --repo "${GITHUB_REPO}" \
    --tag "${GITHUB_VERSION}"

if [[ "$?" -ne 1 ]]; then
  display "Release already exists, cleaning it up."
  github-release delete \
      --security-token  "${GITHUB_TOKEN}" \
      --user "${GITHUB_USER}" \
      --repo "${GITHUB_REPO}" \
      --tag "${GITHUB_VERSION}"
  checkError "Could not delete release, aborting."
fi

display "Creating release on github"
github-release release \
    --security-token  "${GITHUB_TOKEN}" \
    --user "${GITHUB_USER}" \
    --repo "${GITHUB_REPO}" \
    --tag "${GITHUB_VERSION}" \
    --name "${GITHUB_VERSION}" \
    --description "${RELEASE_DOCUMENTATION}" \
    --pre-release
checkError "Could not create release, aborting."


display "Uploading binaries"
cd bin/
for file in $(ls docker-machine*); do
  display "Uploading ${file}..."
  github-release upload \
      --security-token  "${GITHUB_TOKEN}" \
      --user "${GITHUB_USER}" \
      --repo "${GITHUB_REPO}" \
      --tag "${GITHUB_VERSION}" \
      --name "${file}" \
      --file "${file}"
  if [[ "$?" -ne 0 ]]; then
    display "Could not upload ${file}, continuing with others."
  fi
done
cd ..

git remote rm remote.prod.url

rm ${GITHUB_RELEASE_FILE}
