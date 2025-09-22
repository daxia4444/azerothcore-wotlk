#!/bin/bash

echo "=== Boost Configuration Test ==="
echo

# 检查Boost 1.66安装
echo "1. Checking Boost 1.66 installation:"
if [ -d "/usr/local/boost-1.66" ]; then
    echo "✅ Boost 1.66 directory exists: /usr/local/boost-1.66"
    echo "   Include dir: $(ls -d /usr/local/boost-1.66/include 2>/dev/null || echo 'NOT FOUND')"
    echo "   Library dir: $(ls -d /usr/local/boost-1.66/lib 2>/dev/null || echo 'NOT FOUND')"
    echo "   Library count: $(ls /usr/local/boost-1.66/lib/libboost_*.so 2>/dev/null | wc -l) shared libraries"
else
    echo "❌ Boost 1.66 directory not found"
fi

echo

# 检查系统默认Boost
echo "2. System default Boost version:"
if [ -f "/usr/include/boost/version.hpp" ]; then
    SYSTEM_VERSION=$(grep "BOOST_VERSION " /usr/include/boost/version.hpp | awk '{print $3}')
    MAJOR=$((SYSTEM_VERSION / 100000))
    MINOR=$(((SYSTEM_VERSION / 100) % 1000))
    PATCH=$((SYSTEM_VERSION % 100))
    echo "   System Boost version: $MAJOR.$MINOR.$PATCH (version code: $SYSTEM_VERSION)"
else
    echo "   System Boost not found"
fi

echo

# 检查新安装的Boost版本
echo "3. Installed Boost 1.66 version:"
if [ -f "/usr/local/boost-1.66/include/boost/version.hpp" ]; then
    NEW_VERSION=$(grep "BOOST_VERSION " /usr/local/boost-1.66/include/boost/version.hpp | awk '{print $3}')
    MAJOR=$((NEW_VERSION / 100000))
    MINOR=$(((NEW_VERSION / 100) % 1000))
    PATCH=$((NEW_VERSION % 100))
    echo "✅ New Boost version: $MAJOR.$MINOR.$PATCH (version code: $NEW_VERSION)"
else
    echo "❌ New Boost version.hpp not found"
fi

echo

# 测试CMake配置
echo "4. Testing CMake Boost detection:"
cd /data/wow/azerothcore-wotlk

# 清理之前的构建
if [ -d "build" ]; then
    echo "   Cleaning previous build directory..."
    rm -rf build
fi

mkdir -p build
cd build

echo "   Running CMake configuration..."
cmake .. -DBOOST_ROOT=/usr/local/boost-1.66 -DBoost_ROOT=/usr/local/boost-1.66 2>&1 | grep -i boost

echo
echo "=== Test Complete ==="