#!/bin/bash
# ============================================================
# onlyRedis 多平台打包脚本
# 制品统一命名，平台用目录区分:
#   build/bin/darwin-arm64/onlyRedis.app
#   build/bin/darwin-amd64/onlyRedis.app
#   build/bin/windows-amd64/onlyRedis.exe
# ============================================================

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="$PROJECT_DIR/build/bin"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${CYAN}============================================================${NC}"
echo -e "${CYAN}  onlyRedis 多平台打包${NC}"
echo -e "${CYAN}  macOS Apple Silicon  |  macOS Intel  |  Windows${NC}"
echo -e "${CYAN}============================================================${NC}"
echo ""

# 检查 wails 是否安装
if ! command -v wails &>/dev/null; then
  echo -e "${RED}[错误] 未找到 wails CLI，请先安装: go install github.com/wailsapp/wails/v2/cmd/wails@latest${NC}"
  exit 1
fi

echo -e "${GREEN}[✓] wails: $(wails version 2>&1 | head -1)${NC}"
echo ""

cd "$PROJECT_DIR"

# 清理
if [[ "$1" == "--clean" ]]; then
  echo -e "${YELLOW}[清理] 删除旧制品...${NC}"
  rm -rf "$BIN_DIR"
  echo ""
fi

mkdir -p "$BIN_DIR"

# ============================================================
# 构建函数
# ============================================================
build_target() {
  local platform="$1"   # darwin/arm64 | darwin/amd64 | windows/amd64
  local out_dir="$2"    # darwin-arm64 | darwin-amd64 | windows-amd64
  local skip_frontend="${3:-false}"

  echo -e "${YELLOW}[构建] $platform ...${NC}"

  local extra_args=""
  if [[ "$skip_frontend" == "true" ]]; then
    extra_args="-s"
  fi

  wails build -platform "$platform" $extra_args

  # 将制品移入平台目录
  mkdir -p "$BIN_DIR/$out_dir"

  if [[ "$platform" == windows/* ]]; then
    # Windows 交叉编译产物可能是 onlyRedis.exe 或 onlyRedis-amd64.exe
    local src
    if [[ -f "$BIN_DIR/onlyRedis.exe" ]]; then
      src="$BIN_DIR/onlyRedis.exe"
    elif [[ -f "$BIN_DIR/onlyRedis-amd64.exe" ]]; then
      src="$BIN_DIR/onlyRedis-amd64.exe"
    fi
    local dst="$BIN_DIR/$out_dir/onlyRedis.exe"
  else
    # macOS 单平台构建：build/bin/onlyRedis.app
    # macOS 多平台构建：build/bin/onlyRedis-arm64.app
    local src
    if [[ -d "$BIN_DIR/onlyRedis.app" ]]; then
      src="$BIN_DIR/onlyRedis.app"
    elif [[ -d "$BIN_DIR/onlyRedis-arm64.app" ]]; then
      src="$BIN_DIR/onlyRedis-arm64.app"
    elif [[ -d "$BIN_DIR/onlyRedis-amd64.app" ]]; then
      src="$BIN_DIR/onlyRedis-amd64.app"
    fi
    local dst="$BIN_DIR/$out_dir/onlyRedis.app"
  fi

  if [[ -e "$src" ]]; then
    rm -rf "$dst" 2>/dev/null || true
    mv "$src" "$dst"
    local size=$(du -sh "$dst" 2>/dev/null | cut -f1)
    echo -e "${GREEN}  ✓ $dst  (${size})${NC}"
  else
    echo -e "${RED}  ✗ 制品未生成: $src${NC}"
    return 1
  fi
  echo ""
}

# ============================================================
# 依次构建三个平台（首个不跳过前端，后续跳过）
# ============================================================
echo -e "${YELLOW}[打包] 开始多平台构建...${NC}"
echo ""

build_target "darwin/arm64"  "darwin-arm64"  false
build_target "darwin/amd64"  "darwin-amd64"  true
build_target "windows/amd64" "windows-amd64" true

# ============================================================
# 汇总
# ============================================================
echo -e "${GREEN}============================================================${NC}"
echo -e "${GREEN}  打包完成！制品列表:${NC}"
echo -e "${GREEN}============================================================${NC}"
echo ""

print_artifact() {
  local label="$1"
  local path="$2"
  if [[ -e "$path" ]]; then
    local size=$(du -sh "$path" 2>/dev/null | cut -f1)
    echo -e "  ${CYAN}$label${NC}"
    echo -e "    ${path}"
    echo -e "    大小: ${size}"
    echo ""
  else
    echo -e "  ${RED}$label — 未生成${NC}"
    echo ""
  fi
}

print_artifact "macOS Apple Silicon (M 芯片)"  "$BIN_DIR/darwin-arm64/onlyRedis.app"
print_artifact "macOS Intel (x86_64)"          "$BIN_DIR/darwin-amd64/onlyRedis.app"
print_artifact "Windows (x64)"                "$BIN_DIR/windows-amd64/onlyRedis.exe"

echo -e "${GREEN}============================================================${NC}"
echo ""
echo -e "用法:"
echo -e "  ./build.sh          # 增量构建"
echo -e "  ./build.sh --clean  # 清理旧制品后重新构建"
echo ""
