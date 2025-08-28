#!/bin/bash

# 定义颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 切换到 scripts 目录
cd "$(dirname "$0")"

# 检查 go.mod 文件
if [ ! -f "go.mod" ]; then
    echo -e "${YELLOW}初始化 Go 模块...${NC}"
    go mod init github.com/TinsFox/github-hosts/scripts
    go mod tidy
fi

# 版本号
VERSION="1.0.0"

# 构建目录
BUILD_DIR="../build"
BINARY_NAME="github-hosts"

# 支持的平台
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

# 清理构建目录
echo -e "${YELLOW}清理构建目录...${NC}"
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# 遍历平台进行构建
for PLATFORM in "${PLATFORMS[@]}"; do
    # 分割平台信息
    IFS='/' read -r -a array <<< "$PLATFORM"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"

    # 构建输出文件名
    OUTPUT="$BUILD_DIR/${BINARY_NAME}_${GOOS}_${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi

    echo -e "${YELLOW}正在构建 $GOOS/$GOARCH...${NC}"

    # 执行构建
    GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT" -ldflags="-s -w -X main.Version=$VERSION" .

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 构建成功: $OUTPUT${NC}"
    else
        echo -e "${RED}✗ 构建失败: $GOOS/$GOARCH${NC}"
    fi
done

# 创建压缩包
echo -e "${YELLOW}创建压缩包...${NC}"
cd $BUILD_DIR
for FILE in *; do
    if [ -f "$FILE" ]; then
        tar -czf "${FILE}.tar.gz" "$FILE"
        echo -e "${GREEN}✓ 已创建: ${FILE}.tar.gz${NC}"
    fi
done
cd ..

echo -e "${GREEN}构建完成！${NC}"
