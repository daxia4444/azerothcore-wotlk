# AzerothCore 地图数据提取指南

## 📋 概述

AzerothCore 是一个开源的魔兽世界服务器核心，但由于版权原因，地图数据不包含在源码中。本指南将详细说明如何从魔兽世界客户端提取地图数据并正确配置服务器。

## ⚠️ 重要声明

- **版权声明**：地图数据属于暴雪娱乐公司版权，只能从合法拥有的魔兽世界客户端提取
- **法律责任**：用户需确保拥有合法的魔兽世界客户端授权
- **禁止分发**：不得分发或共享提取的地图数据

## 🎯 系统要求

### 硬件要求
- **磁盘空间**：至少 15GB 可用空间
- **内存**：建议 8GB 以上
- **CPU**：多核处理器（提取过程较耗时）

### 软件要求
- **魔兽世界客户端**：3.3.5a (Build 12340)
- **操作系统**：Linux/Windows/macOS
- **编译工具**：CMake 3.16+, GCC 8+ 或 Visual Studio 2019+

## 📂 数据类型说明

| 数据类型 | 文件夹 | 必需性 | 说明 |
|---------|--------|--------|------|
| **DBC** | `dbc/` | ✅ 必需 | 数据库缓存文件，服务器启动必需 |
| **Maps** | `maps/` | ✅ 必需 | 基础地形数据，服务器启动必需 |
| **VMaps** | `vmaps/` | 🔶 推荐 | 视觉地图数据，用于碰撞检测和视线判断 |
| **MMaps** | `mmaps/` | 🔶 推荐 | 移动地图数据，用于NPC寻路系统 |
| **Cameras** | `Cameras/` | 🔹 可选 | 摄像机数据，用于过场动画 |

## 🛠️ 提取工具编译

### 步骤 1：编译 AzerothCore 提取工具

```bash
# 进入 AzerothCore 项目目录
cd /data/wow/azerothcore-wotlk

# 创建编译目录
mkdir -p build && cd build

# 配置编译选项（启用工具编译）
cmake .. \
    -DCMAKE_BUILD_TYPE=Release \
    -DTOOLS=1 \
    -DSCRIPTS=static \
    -DMYSQL_ADD_INCLUDE_PATH=/usr/include/mysql \
    -DMYSQL_ADD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu \
    -DCMAKE_INSTALL_PREFIX=/data/wow/azerothcore-wotlk/env/dist

# 编译（使用所有CPU核心）
make -j$(nproc)

# 安装到指定目录
make install
```

### 步骤 2：验证编译结果

```bash
# 检查提取工具是否编译成功
ls -la /data/wow/azerothcore-wotlk/env/dist/bin/

# 应该看到以下文件：
# map_extractor      - 基础地图提取器
# vmap4_extractor    - 视觉地图提取器  
# vmap4_assembler    - 视觉地图组装器
# mmaps_generator    - 移动地图生成器
```

## 📥 地图数据提取流程

### 步骤 1：准备魔兽世界客户端

```bash
# 确保客户端目录结构正确
WoW_Directory/
├── WoW.exe                    # 游戏主程序
├── Data/                      # 数据文件夹
│   ├── common.MPQ            # 通用数据包
│   ├── common-2.MPQ          # 通用数据包2
│   ├── expansion.MPQ         # 燃烧的远征数据包
│   ├── lichking.MPQ          # 巫妖王之怒数据包
│   ├── patch.MPQ             # 补丁数据包
│   ├── patch-2.MPQ           # 补丁数据包2
│   ├── patch-3.MPQ           # 补丁数据包3
│   └── ...
├── Interface/                 # 界面文件
├── WTF/                      # 配置文件
└── ...
```

### 步骤 2：复制提取工具

```bash
# 设置变量
WOW_CLIENT_PATH="/path/to/your/wow/client"
AZEROTHCORE_PATH="/data/wow/azerothcore-wotlk"

# 复制提取工具到WoW客户端目录
cp ${AZEROTHCORE_PATH}/env/dist/bin/map_extractor ${WOW_CLIENT_PATH}/
cp ${AZEROTHCORE_PATH}/env/dist/bin/vmap4_extractor ${WOW_CLIENT_PATH}/
cp ${AZEROTHCORE_PATH}/env/dist/bin/vmap4_assembler ${WOW_CLIENT_PATH}/
cp ${AZEROTHCORE_PATH}/env/dist/bin/mmaps_generator ${WOW_CLIENT_PATH}/

# 复制提取脚本
cp ${AZEROTHCORE_PATH}/apps/extractor/extractor.sh ${WOW_CLIENT_PATH}/

# 设置执行权限
chmod +x ${WOW_CLIENT_PATH}/extractor.sh
chmod +x ${WOW_CLIENT_PATH}/map_extractor
chmod +x ${WOW_CLIENT_PATH}/vmap4_extractor
chmod +x ${WOW_CLIENT_PATH}/vmap4_assembler
chmod +x ${WOW_CLIENT_PATH}/mmaps_generator
```

### 步骤 3：运行提取脚本

```bash
# 进入WoW客户端目录
cd ${WOW_CLIENT_PATH}

# 运行提取脚本
./extractor.sh
```

### 步骤 4：选择提取选项

脚本运行后会显示菜单：

```
AzerothCore Map & DBC Extractor
===============================
1) Extract DBC & Maps (Required - ~5 minutes)
2) Extract VMaps (Recommended - ~30 minutes)  
3) Extract MMaps (Recommended - ~3-6 hours)
4) Extract All (Recommended - ~4-7 hours)
5) Exit

Please select an option [1-5]:
```

**推荐选择：**
- **选项 4**：提取全部数据（推荐，但耗时较长）
- **选项 1**：仅提取基础数据（最快，但功能受限）

### 步骤 5：等待提取完成

```bash
# 提取过程中会显示进度信息
# DBC & Maps 提取：约 5-10 分钟
# VMaps 提取：约 30-60 分钟  
# MMaps 提取：约 3-6 小时（取决于CPU性能）

# 提取完成后，客户端目录会生成以下文件夹：
# dbc/      - 数据库缓存文件
# maps/     - 基础地形数据
# vmaps/    - 视觉地图数据
# mmaps/    - 移动地图数据
# Cameras/  - 摄像机数据
```

## 📋 数据验证与安装

### 步骤 1：验证提取结果

```bash
# 检查提取的数据完整性
cd ${WOW_CLIENT_PATH}

# 检查DBC文件
echo "DBC files count: $(find dbc/ -name "*.dbc" | wc -l)"
# 应该有 100+ 个 .dbc 文件

# 检查地图文件  
echo "Map files count: $(find maps/ -name "*.map" | wc -l)"
# 应该有 3000+ 个 .map 文件

# 检查VMaps文件（如果提取了）
if [ -d "vmaps" ]; then
    echo "VMap files count: $(find vmaps/ -name "*.vmtree" -o -name "*.vmtile" | wc -l)"
fi

# 检查MMaps文件（如果提取了）
if [ -d "mmaps" ]; then
    echo "MMap files count: $(find mmaps/ -name "*.mmap" | wc -l)"
fi
```

### 步骤 2：安装数据到服务器

```bash
# 创建服务器数据目录
mkdir -p ${AZEROTHCORE_PATH}/data

# 复制提取的数据
cp -r ${WOW_CLIENT_PATH}/dbc ${AZEROTHCORE_PATH}/data/
cp -r ${WOW_CLIENT_PATH}/maps ${AZEROTHCORE_PATH}/data/

# 复制可选数据（如果存在）
if [ -d "${WOW_CLIENT_PATH}/vmaps" ]; then
    cp -r ${WOW_CLIENT_PATH}/vmaps ${AZEROTHCORE_PATH}/data/
fi

if [ -d "${WOW_CLIENT_PATH}/mmaps" ]; then
    cp -r ${WOW_CLIENT_PATH}/mmaps ${AZEROTHCORE_PATH}/data/
fi

if [ -d "${WOW_CLIENT_PATH}/Cameras" ]; then
    cp -r ${WOW_CLIENT_PATH}/Cameras ${AZEROTHCORE_PATH}/data/
fi

# 设置正确的权限
chmod -R 755 ${AZEROTHCORE_PATH}/data/
```

### 步骤 3：配置服务器

```bash
# 编辑世界服务器配置文件
vim ${AZEROTHCORE_PATH}/env/dist/etc/worldserver.conf

# 找到并修改以下配置项：
# DataDir = "/data/wow/azerothcore-wotlk/data"
# 
# 其他相关配置：
# vmap.enableLOS = 1                    # 启用视线检查（需要vmaps）
# vmap.enableHeight = 1                 # 启用高度检查（需要vmaps）  
# mmap.enablePathFinding = 1            # 启用寻路系统（需要mmaps）
```

## 🗂️ 文件结构说明

### 最终数据目录结构

```
/data/wow/azerothcore-wotlk/data/
├── dbc/                              # 数据库缓存文件
│   ├── AreaTable.dbc                # 区域表
│   ├── Map.dbc                      # 地图表
│   ├── GameObjectDisplayInfo.dbc    # 游戏对象显示信息
│   ├── CreatureDisplayInfo.dbc      # 生物显示信息
│   └── ... (100+ files)
├── maps/                            # 基础地形数据
│   ├── 00000000.map                # 东部王国 (0,0) 网格
│   ├── 00000001.map                # 东部王国 (0,1) 网格
│   ├── 00100000.map                # 卡利姆多 (0,0) 网格
│   └── ... (3000+ files)
├── vmaps/                           # 视觉地图数据（可选）
│   ├── 000.vmtree                  # 东部王国视觉地图树
│   ├── 000/                        # 东部王国视觉地图瓦片
│   │   ├── 000_28_28.vmtile
│   │   └── ...
│   └── ...
├── mmaps/                           # 移动地图数据（可选）
│   ├── 000.mmap                    # 东部王国移动地图
│   ├── 000_28_28.mmtile           # 东部王国移动地图瓦片
│   └── ...
└── Cameras/                         # 摄像机数据（可选）
    ├── CinematicCamera.dbc
    └── ...
```

### 地图文件命名规则

```cpp
// 地图文件格式：{MapID:3位}{GridX:2位}{GridY:2位}.map
// 示例：
// 00000000.map = 地图ID:000 (东部王国) + 网格X:00 + 网格Y:00
// 00100515.map = 地图ID:001 (卡利姆多) + 网格X:05 + 网格Y:15
// 53200000.map = 地图ID:532 (卡拉赞) + 网格X:00 + 网格Y:00
```

### 常见地图ID

| 地图ID | 地图名称 | 说明 |
|--------|----------|------|
| 0 | 东部王国 (Eastern Kingdoms) | 主大陆 |
| 1 | 卡利姆多 (Kalimdor) | 主大陆 |
| 530 | 外域 (Outland) | 燃烧的远征 |
| 571 | 诺森德 (Northrend) | 巫妖王之怒 |
| 532 | 卡拉赞 (Karazhan) | 副本 |
| 533 | 纳克萨玛斯 (Naxxramas) | 副本 |

## 🚀 启动服务器

### 步骤 1：验证配置

```bash
# 检查数据目录
ls -la ${AZEROTHCORE_PATH}/data/

# 检查配置文件
grep "DataDir" ${AZEROTHCORE_PATH}/env/dist/etc/worldserver.conf
```

### 步骤 2：启动服务器

```bash
# 进入服务器目录
cd ${AZEROTHCORE_PATH}/env/dist/bin

# 启动认证服务器
./authserver

# 在另一个终端启动世界服务器
./worldserver
```

### 步骤 3：检查启动日志

服务器启动时会显示地图加载信息：

```
Loading DBC files...
Loading Maps...
Loading VMaps... (如果启用)
Loading MMaps... (如果启用)
World initialized in X seconds.
```

## 🔧 故障排除

### 常见问题

#### 1. 提取工具编译失败

```bash
# 检查依赖包
sudo apt-get install build-essential cmake libmysqlclient-dev \
    libssl-dev libbz2-dev libreadline-dev libncurses-dev \
    mysql-server p7zip-full

# 清理并重新编译
cd ${AZEROTHCORE_PATH}/build
make clean
cmake .. -DTOOLS=1
make -j$(nproc)
```

#### 2. 提取过程中断

```bash
# 检查磁盘空间
df -h

# 检查客户端文件完整性
ls -la ${WOW_CLIENT_PATH}/Data/

# 重新运行提取脚本
cd ${WOW_CLIENT_PATH}
./extractor.sh
```

#### 3. 服务器启动失败

```bash
# 检查数据目录权限
chmod -R 755 ${AZEROTHCORE_PATH}/data/

# 检查配置文件
grep -E "DataDir|vmap|mmap" ${AZEROTHCORE_PATH}/env/dist/etc/worldserver.conf

# 查看错误日志
tail -f ${AZEROTHCORE_PATH}/env/dist/logs/Server.log
```

#### 4. 地图文件缺失

```bash
# 检查关键地图文件
ls ${AZEROTHCORE_PATH}/data/maps/000*.map | head -10

# 如果文件缺失，重新提取
cd ${WOW_CLIENT_PATH}
./map_extractor
```

### 性能优化建议

#### 1. 磁盘I/O优化

```bash
# 使用SSD存储地图数据
# 或者将数据目录挂载到内存盘（如果内存充足）
sudo mkdir /mnt/ramdisk
sudo mount -t tmpfs -o size=8G tmpfs /mnt/ramdisk
cp -r ${AZEROTHCORE_PATH}/data/* /mnt/ramdisk/
```

#### 2. 配置优化

```ini
# worldserver.conf 性能相关配置
GridUnload = 1                        # 启用网格卸载
GridCleanUpDelay = 300000            # 网格清理延迟（毫秒）
MapUpdateInterval = 100              # 地图更新间隔（毫秒）
vmap.enableLOS = 1                   # 启用视线检查
vmap.enableHeight = 1                # 启用高度检查
mmap.enablePathFinding = 1           # 启用寻路
```

## 📊 数据统计

### 典型数据大小

| 数据类型 | 文件数量 | 磁盘占用 | 提取时间 |
|---------|----------|----------|----------|
| DBC | ~150 | ~50MB | 1-2分钟 |
| Maps | ~3000 | ~2GB | 3-5分钟 |
| VMaps | ~5000 | ~1.5GB | 30-60分钟 |
| MMaps | ~8000 | ~4GB | 3-6小时 |
| **总计** | **~16000** | **~7.5GB** | **4-7小时** |

### 系统资源使用

- **CPU使用率**：提取期间 80-100%
- **内存使用**：峰值约 2-4GB
- **磁盘I/O**：大量随机读写
- **网络**：无需网络连接

## 📚 参考资料

### 官方文档
- [AzerothCore Wiki](https://www.azerothcore.org/wiki/)
- [Installation Guide](https://www.azerothcore.org/wiki/installation)
- [Database Setup](https://www.azerothcore.org/wiki/database-installation)

### 社区资源
- [AzerothCore Discord](https://discord.gg/gkt4y2x)
- [GitHub Issues](https://github.com/azerothcore/azerothcore-wotlk/issues)
- [Community Forum](https://github.com/azerothcore/azerothcore-wotlk/discussions)

### 技术文档
- [Map System Architecture](https://www.azerothcore.org/wiki/map-system)
- [Grid System](https://www.azerothcore.org/wiki/grid-system)
- [Collision Detection](https://www.azerothcore.org/wiki/collision-detection)

## 📝 更新日志

- **2024-01-15**：初始版本创建
- **2024-01-20**：添加故障排除章节
- **2024-01-25**：完善性能优化建议
- **2024-02-01**：更新数据统计信息

---

## 💡 小贴士

1. **首次提取建议**：先选择选项1（仅DBC和Maps），确保服务器能正常启动后，再提取VMaps和MMaps
2. **磁盘空间**：确保有足够的磁盘空间，建议预留15GB以上
3. **备份数据**：提取完成后建议备份数据，避免重复提取
4. **版本匹配**：确保客户端版本为3.3.5a (12340)，其他版本可能导致提取失败
5. **并行处理**：MMaps提取支持多线程，可以通过参数调整线程数量

---

**祝你游戏愉快！** 🎮