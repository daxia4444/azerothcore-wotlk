# MySQL Configuration Fix Summary

## 问题描述
CMake在第115行报错，无法找到MySQL配置文件：
```
CMake Error at CMakeLists.txt:115 (find_package):By not providing "FindMySQL.cmake" in CMAKE_MODULE_PATH this project has
asked CMake to find a package configuration file provided by "MySQL", but
CMake did not find one.
```

## 问题原因分析
1. **FindMySQL.cmake语法错误**：`src/cmake/macros/FindMySQL.cmake`文件第53行有语法错误
   - 原代码：`set(MYSQL_MINIMUM_VERSION "5.5)`
   - 缺少了一个引号

2. **MySQL路径配置问题**：CMake无法正确定位MySQL的头文件和库文件

## 解决方案

### 1. 修复FindMySQL.cmake语法错误
**文件**：`src/cmake/macros/FindMySQL.cmake`
**修改**：第53行
```cmake
# 修复前
set(MYSQL_MINIMUM_VERSION "5.5)

# 修复后  
set(MYSQL_MINIMUM_VERSION "5.5")
```

### 2. 更新CMakeLists.txt中的MySQL配置
**文件**：`CMakeLists.txt`
**位置**：第113-127行
**修改内容**：
```cmake
# 手动设置MySQL路径（避免mysql_config问题）
set(MYSQL_INCLUDE_DIR "/usr/include/mysql" CACHE PATH "MySQL include directory")
set(MYSQL_LIBRARY "/usr/lib64/mysql/libmysqlclient.so" CACHE FILEPATH "MySQL client library")
set(MYSQL_CONFIG_PREFER_PATH "/usr/lib64/mysql" CACHE FILEPATH "preferred path to MySQL (mysql_config)")
set(MYSQL_ROOT_DIR "/usr" CACHE PATH "MySQL root directory")

# 验证MySQL文件存在
if(NOT EXISTS "${MYSQL_INCLUDE_DIR}/mysql.h")
    message(FATAL_ERROR "MySQL header file not found at ${MYSQL_INCLUDE_DIR}/mysql.h")
endif()

if(NOT EXISTS "${MYSQL_LIBRARY}")
    message(FATAL_ERROR "MySQL library not found at ${MYSQL_LIBRARY}")
endif()

# 查找MySQL包
find_package(MySQL REQUIRED)
```

## 验证结果
- ✅ MySQL头文件存在：`/usr/include/mysql/mysql.h`
- ✅ MySQL库文件存在：`/usr/lib64/mysql/libmysqlclient.so`
- ✅ FindMySQL.cmake语法错误已修复
- ✅ CMakeLists.txt配置已更新

## 下一步操作
1. 清理之前的构建目录：`rm -rf build`
2. 重新创建构建目录：`mkdir build && cd build`
3. 运行CMake配置：`cmake ..`
4. 编译项目：`make -j$(nproc)`

## 技术说明
- 使用了直接路径设置而不是依赖`mysql_config`脚本，避免了执行权限问题
- 添加了文件存在性验证，确保在配置阶段就能发现路径问题
- 保持了与原有FindMySQL.cmake模块的兼容性

## 修改的文件列表
1. `/data/wow/azerothcore-wotlk/src/cmake/macros/FindMySQL.cmake` - 修复语法错误
2. `/data/wow/azerothcore-wotlk/CMakeLists.txt` - 更新MySQL配置