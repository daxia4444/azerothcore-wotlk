#!/bin/bash

echo "=== MySQL Configuration Test ==="
echo

echo "1. Checking MySQL config binary:"
which mysql_config
if [ $? -eq 0 ]; then
    echo "✓ mysql_config found"
    echo "MySQL version: $(/usr/lib64/mysql/mysql_config --version 2>/dev/null || echo 'Unable to get version')"
    echo "MySQL include path: $(/usr/lib64/mysql/mysql_config --include 2>/dev/null || echo 'Unable to get include path')"
    echo "MySQL library path: $(/usr/lib64/mysql/mysql_config --libs 2>/dev/null || echo 'Unable to get library path')"
else
    echo "✗ mysql_config not found in PATH"
fi

echo
echo "2. Checking MySQL header files:"
if [ -f "/usr/include/mysql/mysql.h" ]; then
    echo "✓ MySQL headers found at /usr/include/mysql/mysql.h"
else
    echo "✗ MySQL headers not found"
fi

echo
echo "3. Checking MySQL library files:"
if [ -f "/usr/lib64/mysql/libmysqlclient.so" ]; then
    echo "✓ MySQL client library found at /usr/lib64/mysql/libmysqlclient.so"
    ls -la /usr/lib64/mysql/libmysqlclient*
else
    echo "✗ MySQL client library not found"
fi

echo
echo "4. Testing CMake MySQL detection:"
cd /data/wow/azerothcore-wotlk
rm -rf build_test
mkdir build_test
cd build_test

cat > test_mysql.cmake << 'EOF'
cmake_minimum_required(VERSION 3.16)
project(TestMySQL)

# 添加模块路径
list(APPEND CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/../src/cmake/macros")

# 设置MySQL环境变量
set(MYSQL_CONFIG_PREFER_PATH "/usr/lib64/mysql" CACHE FILEPATH "preferred path to MySQL (mysql_config)")
set(MYSQL_ROOT_DIR "/usr" CACHE PATH "MySQL root directory")

# 查找MySQL
find_package(MySQL REQUIRED)

if(MYSQL_FOUND)
    message(STATUS "✓ MySQL found successfully!")
    message(STATUS "MySQL include dir: ${MYSQL_INCLUDE_DIR}")
    message(STATUS "MySQL library: ${MYSQL_LIBRARY}")
    if(MYSQL_EXECUTABLE)
        message(STATUS "MySQL executable: ${MYSQL_EXECUTABLE}")
    endif()
else()
    message(FATAL_ERROR "✗ MySQL not found!")
endif()
EOF

echo "Running CMake test..."
cmake -f test_mysql.cmake .

echo
echo "=== Test Complete ==="