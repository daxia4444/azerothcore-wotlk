# Boost 1.66.0 安装指南

## 快速安装

### 方法一：使用提供的安装脚本（推荐）

```bash
# 1. 切换到项目目录
cd /data/wow/azerothcore-wotlk

# 2. 以root权限运行安装脚本
sudo ./install_boost_1.66.sh

# 3. 安装完成后，加载环境变量
source /etc/profile.d/boost-1.66.sh

# 4. 重新编译项目
rm -rf build
mkdir build && cd build
cmake ..
make -j$(nproc)
```

### 方法二：手动安装

```bash
# 1. 下载并解压Boost 1.66.0
cd /tmp
wget https://boostorg.jfrog.io/artifactory/main/release/1.66.0/source/boost_1_66_0.tar.gz
tar -xzf boost_1_66_0.tar.gz
cd boost_1_66_0

# 2. 激活devtoolset-8环境
source /opt/rh/devtoolset-8/enable
export CC=/opt/rh/devtoolset-8/root/usr/bin/gcc
export CXX=/opt/rh/devtoolset-8/root/usr/bin/g++

# 3. 配置构建系统
./bootstrap.sh --prefix=/usr/local/boost-1.66 --with-toolset=gcc

# 4. 编译安装（需要30-60分钟）
./b2 --toolset=gcc variant=release threading=multi link=shared runtime-link=shared -j$(nproc)
sudo ./b2 install --toolset=gcc variant=release threading=multi link=shared runtime-link=shared

# 5. 配置环境变量
sudo bash -c 'cat > /etc/profile.d/boost-1.66.sh << EOF
export BOOST_ROOT=/usr/local/boost-1.66
export Boost_ROOT=/usr/local/boost-1.66
export LD_LIBRARY_PATH=/usr/local/boost-1.66/lib:\$LD_LIBRARY_PATH
EOF'

# 6. 更新动态链接库
echo "/usr/local/boost-1.66/lib" | sudo tee /etc/ld.so.conf.d/boost-1.66.conf
sudo ldconfig

# 7. 加载环境变量
source /etc/profile.d/boost-1.66.sh
```

## 验证安装

```bash
# 检查Boost版本
ls -la /usr/local/boost-1.66/include/boost/version.hpp
grep "BOOST_VERSION " /usr/local/boost-1.66/include/boost/version.hpp

# 检查库文件
ls -la /usr/local/boost-1.66/lib/libboost_*.so
```

## 重新编译项目

```bash
cd /data/wow/azerothcore-wotlk
rm -rf build
mkdir build && cd build

# 确保环境变量已设置
source /etc/profile.d/boost-1.66.sh

# 配置和编译
cmake ..
make -j$(nproc)
```

## 故障排除

### 如果CMake仍然找不到Boost：

```bash
# 手动指定Boost路径
cmake -DBOOST_ROOT=/usr/local/boost-1.66 -DBoost_ROOT=/usr/local/boost-1.66 ..
```

### 如果VSCode仍然报错：

1. 重启VSCode
2. 按 `Ctrl+Shift+P`，执行 "C/C++: Reload IntelliSense Database"
3. 按 `Ctrl+Shift+P`，执行 "C/C++: Select IntelliSense Configuration"，选择 "Linux"

## 已完成的配置更改

1. ✅ 更新了 `.vscode/c_cpp_properties.json` - 添加了Boost 1.66头文件路径
2. ✅ 更新了 `.vscode/settings.json` - 添加了Boost 1.66头文件路径  
3. ✅ 修改了 `deps/boost/CMakeLists.txt` - 降低了Boost版本要求到1.66
4. ✅ 创建了自动安装脚本 `install_boost_1.66.sh`

## 注意事项

- 编译过程可能需要30-60分钟，请耐心等待
- 确保有足够的磁盘空间（至少2GB）
- 编译过程中会使用所有CPU核心，系统可能会比较繁忙
- 安装完成后需要重新登录或手动加载环境变量