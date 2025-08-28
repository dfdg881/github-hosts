@echo off
setlocal enabledelayedexpansion

:: 切换到脚本所在目录
cd /d "%~dp0"

:: 版本号
set VERSION=1.0.0

:: 构建目录
set BUILD_DIR=..\build
set BINARY_NAME=github-hosts

:: 清理构建目录
echo 清理构建目录...
if exist %BUILD_DIR% rd /s /q %BUILD_DIR%
mkdir %BUILD_DIR%

:: 定义支持的平台
set PLATFORMS=^
    darwin/amd64^
    darwin/arm64^
    linux/amd64^
    linux/arm64^
    windows/amd64

:: 遍历平台进行构建
for %%p in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
        set GOOS=%%a
        set GOARCH=%%b

        :: 构建输出文件名
        set OUTPUT=%BUILD_DIR%\%BINARY_NAME%_!GOOS!_!GOARCH!
        if "!GOOS!"=="windows" set OUTPUT=!OUTPUT!.exe

        echo 正在构建 !GOOS!/!GOARCH!...

        :: 执行构建
        set GOOS=!GOOS!
        set GOARCH=!GOARCH!
        go build -o "!OUTPUT!" -ldflags="-s -w -X main.Version=%VERSION%" .

        if !errorlevel! equ 0 (
            echo √ 构建成功: !OUTPUT!
        ) else (
            echo × 构建失败: !GOOS!/!GOARCH!
        )
    )
)

:: 创建压缩包
echo 创建压缩包...
cd %BUILD_DIR%
for %%f in (*) do (
    if exist "%%f" (
        tar -czf "%%f.tar.gz" "%%f"
        del "%%f"
        echo √ 已创建: %%f.tar.gz
    )
)
cd ..

echo 构建完成！
dir /b %BUILD_DIR%

endlocal