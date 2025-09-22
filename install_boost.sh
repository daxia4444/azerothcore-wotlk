#!/bin/bash

# Boost 1.66.0 安装脚本 (空间优化版)
# 专门解决磁盘空间不足问题

set -e

echo "=== Boost 1.66.0 安装脚本 (空间优化版) ==="

# 检查是否为root用户
if [[ $EUID -ne 0 ]]; then
   echo "此脚本需要root权限运行，请使用 sudo 执行"
   exit 1
fi

# 设置变量
BOOST_VERSION="1.66.0"
BOOST_VERSION_UNDERSCORE="1_66_0"
INSTALL_PREFIX="/usr/local/boost-1.66"
TEMP_DIR="/data/wow/azerothcore-wotlk/boost_temp"
PROJECT_DIR="/data/wow/azerothcore-wotlk"

echo "项目目录: $PROJECT_DIR"
echo "临时目录: $TEMP_DIR"
echo "安装目录: $INSTALL_PREFIX"

# 清理函数
cleanup() {
    echo "清理临时文件..."
    rm -rf "$TEMP_DIR"
    # 清理/tmp中的旧文件
    rm -rf /tmp/boost_build*
}

# 设置清理陷阱
trap cleanup EXIT

# 检查可用空间
echo "检查磁盘空间..."
AVAILABLE_SPACE=$(df "$PROJECT_DIR" | awk 'NR==2 {print $4}')
REQUIRED_SPACE=2097152  # 2GB in KB

if [ "$AVAILABLE_SPACE" -lt "$REQUIRED_SPACE" ]; then
    echo "警告: 可用空间不足"
    echo "可用空间: $(($AVAILABLE_SPACE/1024))MB"
    echo "建议空间: 2048MB"
    echo ""
    echo "正在清理临时文件..."
    
    # 清理/tmp中的旧文件
    rm -rf /tmp/boost_build* /tmp/vscode-* /tmp/pipe_* /tmp/tmp-* 2>/dev/null || true
    
    # 重新检查空间
    AVAILABLE_SPACE=$(df "$PROJECT_DIR" | awk 'NR==2 {print $4}')
    echo "清理后可用空间: $(($AVAILABLE_SPACE/1024))MB"
    
    if [ "$AVAILABLE_SPACE" -lt 1048576 ]; then  # 1GB
        echo "错误: 空间仍然不足，需要至少1GB空间"
        echo "请手动清理磁盘空间后重试"
        exit 1
    fi
fi

# 创建临时目录
echo "创建临时构建目录..."
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

# 下载Boost源码
echo "下载 Boost ${BOOST_VERSION} 源码..."

# 使用最可靠的下载源
BOOST_URL="https://sourceforge.net/projects/boost/files/boost/${BOOST_VERSION}/boost_${BOOST_VERSION_UNDERSCORE}.tar.gz/download"

echo "从 SourceForge 下载..."
if ! wget --timeout=60 --tries=2 -O boost_${BOOST_VERSION_UNDERSCORE}.tar.gz "$BOOST_URL"; then
    echo "SourceForge下载失败，尝试备用源..."
    BOOST_URL="https://archives.boost.io/release/${BOOST_VERSION}/source/boost_${BOOST_VERSION_UNDERSCORE}.tar.gz"
    if ! wget --timeout=60 --tries=2 -O boost_${BOOST_VERSION_UNDERSCORE}.tar.gz "$BOOST_URL"; then
        echo "所有下载源都失败了"
        echo "请手动下载文件到 $TEMP_DIR"
        echo "下载链接: https://www.boost.org/users/history/version_1_66_0.html"
        exit 1
    fi
fi

# 验证下载的文件
if ! file boost_${BOOST_VERSION_UNDERSCORE}.tar.gz | grep -q "gzip compressed"; then
    echo "下载的文件格式不正确"
    head -3 boost_${BOOST_VERSION_UNDERSCORE}.tar.gz
    exit 1
fi

FILE_SIZE=$(stat -c%s boost_${BOOST_VERSION_UNDERSCORE}.tar.gz)
if [ $FILE_SIZE -lt 50000000 ]; then
    echo "文件大小异常: $(($FILE_SIZE/1024/1024))MB"
    exit 1
fi

echo "✓ 下载成功 (大小: $(($FILE_SIZE/1024/1024))MB)"

# 解压源码
echo "解压源码..."
tar -xzf boost_${BOOST_VERSION_UNDERSCORE}.tar.gz
cd boost_${BOOST_VERSION_UNDERSCORE}

# 删除压缩包以节省空间
rm -f ../boost_${BOOST_VERSION_UNDERSCORE}.tar.gz

# 激活devtoolset-8环境
echo "激活 devtoolset-8 编译环境..."
source /opt/rh/devtoolset-8/enable
export CC=/opt/rh/devtoolset-8/root/usr/bin/gcc
export CXX=/opt/rh/devtoolset-8/root/usr/bin/g++

echo "使用的编译器版本:"
$CC --version | head -1

# 配置Boost构建系统
echo "配置 Boost 构建系统..."
./bootstrap.sh --prefix=$INSTALL_PREFIX --with-toolset=gcc

# 只编译必需的库以节省空间和时间
echo "编译 Boost (仅编译必需库)..."
NPROC=$(nproc)
echo "使用 $NPROC 个CPU核心进行并行编译..."

# 只编译项目需要的库
./b2 --toolset=gcc variant=release threading=multi link=shared runtime-link=shared \
     --with-system --with-filesystem --with-thread --with-date_time --with-regex \
     --with-serialization --with-program_options --with-iostreams \
     -j$NPROC

# 安装
echo "安装 Boost 到 $INSTALL_PREFIX..."
./b2 install --toolset=gcc variant=release threading=multi link=shared runtime-link=shared \
     --with-system --with-filesystem --with-thread --with-date_time --with-regex \
     --with-serialization --with-program_options --with-iostreams

# 创建符号链接
echo "创建符号链接..."
if [ ! -L /usr/local/boost ]; then
    ln -sf $INSTALL_PREFIX /usr/local/boost
fi

# 更新动态链接库配置
echo "更新动态链接库配置..."
echo "$INSTALL_PREFIX/lib" > /etc/ld.so.conf.d/boost-1.66.conf
ldconfig

# 设置环境变量
echo "设置环境变量..."
cat > /etc/profile.d/boost-1.66.sh << EOF
export BOOST_ROOT=$INSTALL_PREFIX
export Boost_ROOT=$INSTALL_PREFIX
export LD_LIBRARY_PATH=$INSTALL_PREFIX/lib:\$LD_LIBRARY_PATH
export PKG_CONFIG_PATH=$INSTALL_PREFIX/lib/pkgconfig:\$PKG_CONFIG_PATH
EOF

# 验证安装
echo "验证安装..."
if [ -f "$INSTALL_PREFIX/include/boost/version.hpp" ]; then
    echo "✓ Boost 头文件安装成功"
    INSTALLED_VERSION=$(grep "BOOST_VERSION " $INSTALL_PREFIX/include/boost/version.hpp | cut -d' ' -f3)
    echo "  安装的版本号: $INSTALLED_VERSION"
else
    echo "✗ Boost 头文件未找到"
    exit 1
fi

if [ -f "$INSTALL_PREFIX/lib/libboost_system.so" ]; then
    echo "✓ Boost 库文件安装成功"
    echo "  已安装的库文件:"
    ls -la $INSTALL_PREFIX/lib/libboost_*.so | head -5
else
    echo "✗ Boost 库文件未找到"
    exit 1
fi

echo ""
echo "=== 安装完成 ==="
echo "Boost 1.66.0 已成功安装到: $INSTALL_PREFIX"
echo ""
echo "请执行以下命令使环境变量生效:"
echo "source /etc/profile.d/boost-1.66.sh"
echo ""
echo "接下来可以重新编译您的项目:"
echo "cd $PROJECT_DIR"
echo "rm -rf build"
echo "mkdir build && cd build"
echo "source /etc/profile.d/boost-1.66.sh"
echo "cmake .."
echo "make -j\$(nproc)"