#!/usr/bin/env bash
# 构建本地 sub2api 镜像，并生成可直接启动的本地 compose 文件。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SOURCE_COMPOSE="${REPO_ROOT}/deploy/docker-compose.local.yml"

IMAGE_NAME="sub2api:local"
OUTPUT_PATH="${REPO_ROOT}/deploy/docker-compose.local-self.yml"
SKIP_BUILD=false

usage() {
  cat <<EOF
Usage: $(basename "$0") [options]

Options:
  --image <name>    本地镜像名，默认: sub2api:local
  --output <path>   输出 compose 文件路径，默认: deploy/docker-compose.local-self.yml
  --skip-build      只生成 compose 文件，不执行 docker build
  -h, --help        显示帮助
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --image)
      [[ $# -ge 2 ]] || { echo "Missing value for --image" >&2; exit 1; }
      IMAGE_NAME="$2"
      shift 2
      ;;
    --output)
      [[ $# -ge 2 ]] || { echo "Missing value for --output" >&2; exit 1; }
      OUTPUT_PATH="$2"
      shift 2
      ;;
    --skip-build)
      SKIP_BUILD=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ ! -f "${SOURCE_COMPOSE}" ]]; then
  echo "Source compose file not found: ${SOURCE_COMPOSE}" >&2
  exit 1
fi

if [[ "${SKIP_BUILD}" != "true" ]]; then
  docker build -t "${IMAGE_NAME}" \
    --build-arg GOPROXY=https://goproxy.cn,direct \
    --build-arg GOSUMDB=sum.golang.google.cn \
    -f "${REPO_ROOT}/Dockerfile" \
    "${REPO_ROOT}"
fi

mkdir -p "$(dirname "${OUTPUT_PATH}")"

awk -v image="${IMAGE_NAME}" '
  /^  sub2api:$/ {
    in_sub2api=1
    print
    next
  }

  in_sub2api && /^  [^ ]/ {
    if (!replaced) {
      print "Failed to find image line for sub2api service" > "/dev/stderr"
      exit 2
    }
    in_sub2api=0
  }

  in_sub2api && /^    image:/ {
    print "    image: " image
    print "    pull_policy: never"
    replaced=1
    next
  }

  in_sub2api && /^    pull_policy:/ {
    next
  }

  {
    print
  }

  END {
    if (!replaced) {
      print "Failed to replace sub2api image in compose template" > "/dev/stderr"
      exit 2
    }
  }
' "${SOURCE_COMPOSE}" > "${OUTPUT_PATH}"

echo "Generated ${OUTPUT_PATH}"
echo "Image: ${IMAGE_NAME}"
