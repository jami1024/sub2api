#!/usr/bin/env bash
# 构建本地 sub2api 镜像，并生成可直接启动的本地 compose 文件。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SOURCE_COMPOSE="${REPO_ROOT}/deploy/docker-compose.local.yml"

IMAGE_NAME="sub2api:local"
OUTPUT_PATH="${REPO_ROOT}/deploy/docker-compose.local-self.yml"
SKIP_BUILD=false
GOPROXY_VALUE="${GOPROXY:-https://proxy.golang.org,direct}"
GOSUMDB_VALUE="${GOSUMDB:-sum.golang.org}"
COMMIT_VALUE="${COMMIT:-}"
DATE_VALUE="${DATE:-}"
VALIDATE_COMPOSE="${SUB2API_LOCAL_IMAGE_VALIDATE_COMPOSE:-true}"
BUILD_MEMORY_VALUE="${BUILD_MEMORY:-}"
BUILD_CPUSET_CPUS_VALUE="${BUILD_CPUSET_CPUS:-}"
BUILD_CPUS_VALUE="${BUILD_CPUS:-}"

usage() {
  cat <<EOF_USAGE
Usage: $(basename "$0") [options]

Options:
  --image <name>    本地镜像名，默认: sub2api:local
  --output <path>   输出 compose 文件路径，默认: deploy/docker-compose.local-self.yml
  --skip-build      只生成 compose 文件，不执行 docker build
  -h, --help        显示帮助

Environment:
  GOPROXY                         Go module proxy，默认: https://proxy.golang.org,direct
  GOSUMDB                         Go checksum database，默认: sum.golang.org
  COMMIT                          镜像内版本提交号，默认自动读取 git short HEAD
  DATE                            镜像内构建时间，默认当前 UTC 时间
  BUILD_MEMORY                    限制 docker build 内存，例如: 4g
  BUILD_CPUSET_CPUS               限制 docker build 使用的 CPU 编号，例如: 0-3
  BUILD_CPUS                      限制 docker build CPU 配额，例如: 4；需当前 Docker 支持 --cpus
  SUB2API_LOCAL_IMAGE_VALIDATE_COMPOSE=false  跳过 docker compose config 校验
EOF_USAGE
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

if [[ -z "${COMMIT_VALUE}" ]]; then
  if command -v git >/dev/null 2>&1 && git -C "${REPO_ROOT}" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    COMMIT_VALUE="$(git -C "${REPO_ROOT}" rev-parse --short HEAD)"
  else
    COMMIT_VALUE="unknown"
  fi
fi

if [[ -z "${DATE_VALUE}" ]]; then
  DATE_VALUE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
fi

if [[ "${SKIP_BUILD}" != "true" ]]; then
  DOCKER_BUILD_CMD=(docker build)
  if [[ -n "${BUILD_MEMORY_VALUE}" ]]; then
    DOCKER_BUILD_CMD+=(--memory "${BUILD_MEMORY_VALUE}")
  fi
  if [[ -n "${BUILD_CPUSET_CPUS_VALUE}" ]]; then
    DOCKER_BUILD_CMD+=(--cpuset-cpus "${BUILD_CPUSET_CPUS_VALUE}")
  fi
  if [[ -n "${BUILD_CPUS_VALUE}" ]]; then
    if docker build --help 2>/dev/null | grep -q -- '--cpus'; then
      DOCKER_BUILD_CMD+=(--cpus "${BUILD_CPUS_VALUE}")
    else
      echo "当前 docker build 不支持 BUILD_CPUS/--cpus；请改用 BUILD_CPUSET_CPUS，例如 BUILD_CPUSET_CPUS=0-3" >&2
      exit 1
    fi
  fi

  "${DOCKER_BUILD_CMD[@]}" -t "${IMAGE_NAME}" \
    --build-arg "GOPROXY=${GOPROXY_VALUE}" \
    --build-arg "GOSUMDB=${GOSUMDB_VALUE}" \
    --build-arg "COMMIT=${COMMIT_VALUE}" \
    --build-arg "DATE=${DATE_VALUE}" \
    -f "${REPO_ROOT}/Dockerfile" \
    "${REPO_ROOT}"
fi

mkdir -p "$(dirname "${OUTPUT_PATH}")"
TMP_OUTPUT="$(mktemp "$(dirname "${OUTPUT_PATH}")/.local-self.XXXXXX.yml")"
cleanup_tmp() {
  rm -f "${TMP_OUTPUT}"
}
trap cleanup_tmp EXIT

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
' "${SOURCE_COMPOSE}" | awk '
  /^    container_name: sub2api$/ {
    print "    container_name: sub2api-personal"
    next
  }

  /^    container_name: sub2api-postgres$/ {
    print "    container_name: sub2api-postgres-personal"
    next
  }

  /^    container_name: sub2api-redis$/ {
    print "    container_name: sub2api-redis-personal"
    next
  }

  /^      - .*:8080"$/ && !port_replaced {
    print "      - \"${BIND_HOST:-0.0.0.0}:${SERVER_PORT:-18080}:8080\""
    port_replaced=1
    next
  }

  {
    print
  }

  END {
    if (!port_replaced) {
      print "Failed to replace sub2api port mapping in compose template" > "/dev/stderr"
      exit 2
    }
  }
' > "${TMP_OUTPUT}"

if [[ "${VALIDATE_COMPOSE}" != "false" ]]; then
  if ! command -v docker >/dev/null 2>&1; then
    echo "docker command not found; set SUB2API_LOCAL_IMAGE_VALIDATE_COMPOSE=false to skip compose validation" >&2
    exit 1
  fi
  POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-sub2api-local-compose-check}" docker compose -f "${TMP_OUTPUT}" config >/dev/null
fi

mv "${TMP_OUTPUT}" "${OUTPUT_PATH}"
trap - EXIT

echo "Generated ${OUTPUT_PATH}"
echo "Image: ${IMAGE_NAME}"
echo "GOPROXY: ${GOPROXY_VALUE}"
echo "GOSUMDB: ${GOSUMDB_VALUE}"
echo "Commit: ${COMMIT_VALUE}"
echo "Date: ${DATE_VALUE}"
if [[ -n "${BUILD_MEMORY_VALUE}${BUILD_CPUSET_CPUS_VALUE}${BUILD_CPUS_VALUE}" ]]; then
  echo "Build memory limit: ${BUILD_MEMORY_VALUE:-unlimited}"
  echo "Build CPU set: ${BUILD_CPUSET_CPUS_VALUE:-unlimited}"
  echo "Build CPU quota: ${BUILD_CPUS_VALUE:-unlimited}"
fi
