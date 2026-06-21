#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
VERSION="$(cat "$ROOT/version.txt")"

BOLD='\033[1m'
DIM='\033[2m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info()  { echo -e "${BOLD}${GREEN}[Build]${NC} $*"; }
warn()  { echo -e "${BOLD}${YELLOW}[Warn]${NC} $*"; }
error() { echo -e "${BOLD}${RED}[ERROR]${NC} $*"; }
step()  { echo; echo -e "${BOLD}━━ $* ━━${NC}"; }
ok()    { echo -e "${GREEN}✓${NC} $*"; }

usage() {
    cat <<EOF
Usage: $(basename "$0") [target]

Targets:
  build       (default)  Compile and package fat JAR
  shadowJar             Build only the shadow JAR (faster if compiled)
  clean                 Clean build artifacts
  compile               Compile only (no JAR)
  jar                   Build the standard JAR
  help                  Show this help

Examples:
  ./build.sh                    # Full build
  ./build.sh compile            # Just compile
  ./build.sh clean build        # Clean then build
EOF
    exit 0
}

clean_build() {
    step "Cleaning build artifacts"
    rm -rf "$ROOT/build"
    ok "Cleaned"
}

compile_only() {
    step "Compiling Java sources"
    "$ROOT/gradlew" compileJava --no-daemon --quiet
    ok "Compilation finished"
}

build_jar() {
    step "Building standard JAR"
    "$ROOT/gradlew" jar --no-daemon --quiet
    ok "Standard JAR built"
}

build_shadow() {
    step "Building shadow (fat) JAR"
    "$ROOT/gradlew" shadowJar --no-daemon --quiet
    ok "Shadow JAR built"
}

full_build() {
    step "Full build — novacore-engine v${VERSION}"
    "$ROOT/gradlew" build --no-daemon --quiet
    ok "Build finished"
}

# ── Main ──────────────────────────────────────────────────────────────────────

if [[ ! -x "$ROOT/gradlew" ]]; then
    error "gradlew not found or not executable at $ROOT/gradlew"
    echo "  Run 'chmod +x gradlew' and try again."
    exit 1
fi

if ! command -v java &>/dev/null; then
    error "Java not found. Install JDK 25+ and add it to PATH."
    exit 1
fi

JAVA_VER=$(java -version 2>&1 | awk -F '"' '/version/ {print $2}')
info "NovaCore-Engine v${VERSION} — NovaStepStudios"
info "Java: ${JAVA_VER}"

TARGET="${1:-build}"

case "$TARGET" in
    build)     full_build ;;
    shadowJar) build_shadow ;;
    clean)     clean_build ;;
    compile)   compile_only ;;
    jar)       build_jar ;;
    help|--help|-h) usage ;;
    *)
        error "Unknown target: $TARGET"
        echo
        usage
        ;;
esac

echo
echo -e "${BOLD}${GREEN}════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${GREEN}  Build: ${TARGET}${NC}"
echo -e "${BOLD}${GREEN}  JAR:   build/libs/novacore-engine.jar${NC}"
echo -e "${BOLD}${GREEN}════════════════════════════════════════════════${NC}"
echo
echo "Run: java -jar build/libs/novacore-engine.jar"
echo "     java -jar build/libs/novacore-engine.jar --port 7878 --ws-port 7879 --threads 32"
