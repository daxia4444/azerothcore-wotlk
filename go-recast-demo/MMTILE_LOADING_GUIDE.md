# 📋 .mmtile 文件加载到 Recast 库完整指南

## 🎯 概述

本文档详细解释了 AzerothCore 项目中如何将 `.mmtile` 文件加载到 Recast Navigation 库，以及整个参数传递过程。

## 🔄 完整的数据流转过程

```
WoW 原始地图数据 → 地图提取工具 → .mmtile 文件 → Go 程序 → CGO → Recast 库 → 寻路查询
```

## 📁 文件结构和命名规则

### AzerothCore .mmtile 文件命名规则：
```
{dataPath}/mmaps/{mapID:03d}{y:02d}{x:02d}.mmtile
```

**示例：**
- 地图 0 (东部王国)，瓦片 (32, 32): `000/mmaps/0003232.mmtile`
- 地图 1 (卡利姆多)，瓦片 (28, 35): `001/mmaps/0012835.mmtile`

## 🔧 关键代码实现分析

### 1. C 代码部分 - Recast 库接口

```c
// 添加瓦片到导航网格 (核心函数)
dtTileRef addTileToNavMesh(dtNavMesh* navMesh, unsigned char* data, int dataSize, int flags) {
    if (!navMesh || !data || dataSize <= 0) {
        return 0;
    }

    dtTileRef tileRef = 0;
    // 🔑 关键调用：将二进制数据传递给 Recast 库
    dtStatus status = navMesh->addTile(data, dataSize, flags, 0, &tileRef);

    if (dtStatusFailed(status)) {
        return 0;
    }

    return tileRef;
}

// 初始化导航网格参数
dtStatus initNavMesh(dtNavMesh* navMesh, dtNavMeshParams* params) {
    if (!navMesh || !params) {
        return DT_FAILURE;
    }
    return navMesh->init(params);
}

// 创建导航网格参数
void createNavMeshParams(dtNavMeshParams* params, float* bmin, float* bmax,
                        float tileWidth, float tileHeight, int maxTiles, int maxPolys) {
    if (!params || !bmin || !bmax) {
        return;
    }

    memset(params, 0, sizeof(dtNavMeshParams));
    dtVcopy(params->orig, bmin);           // 地图原点
    params->tileWidth = tileWidth;         // 瓦片宽度
    params->tileHeight = tileHeight;       // 瓦片高度
    params->maxTiles = maxTiles;           // 最大瓦片数
    params->maxPolys = maxPolys;           // 最大多边形数
}
```

### 2. Go 代码部分 - 参数传递流程

#### 2.1 主加载函数 `LoadMap`

```go
func (mgr *MMapManager) LoadMap(mapID uint32, x, y int32) bool {
    // 步骤 1: 创建导航网格容器
    mapData = &MapData{
        navMesh: C.createNavMesh(), // 创建空的导航网格
        // ...
    }

    // 步骤 2: 🔑 初始化导航网格参数 (告诉 Recast 地图的基本信息)
    if !mgr.initializeNavMeshParams(mapData.navMesh, mapID) {
        return false
    }

    // 步骤 3: 加载 .mmtile 文件数据
    tile := mgr.loadNavMeshTile(mapID, x, y)

    // 步骤 4: 🔑 将瓦片数据传递给 Recast 库
    tileRef := mgr.addTileToNavMesh(mapData.navMesh, tile)
    
    return tileRef != 0
}
```

#### 2.2 导航网格参数初始化

```go
func (mgr *MMapManager) initializeNavMeshParams(navMesh *C.dtNavMesh, mapID uint32) bool {
    // 🔑 关键参数设置
    var params C.dtNavMeshParams
    
    // AzerothCore 地图参数
    mapSize := float32(MAP_SIZE * TILE_SIZE) // 64 * 533.33 = 34133.33 码
    halfMapSize := mapSize / 2.0

    // 地图边界
    bmin := [3]C.float{-halfMapSize, -500.0, -halfMapSize}
    bmax := [3]C.float{halfMapSize, 500.0, halfMapSize}

    // 🔑 调用 C 函数创建参数
    C.createNavMeshParams(&params, &bmin[0], &bmax[0],
        C.float(TILE_SIZE),           // 瓦片宽度: 533.33 码
        C.float(TILE_SIZE),           // 瓦片高度: 533.33 码
        C.int(MAP_SIZE*MAP_SIZE),     // 最大瓦片数: 64*64 = 4096
        C.int(65536))                 // 每瓦片最大多边形数

    // 🔑 初始化导航网格
    status := C.initNavMesh(navMesh, &params)
    return !C.dtStatusFailed(status)
}
```

#### 2.3 瓦片文件加载

```go
func (mgr *MMapManager) loadNavMeshTile(mapID uint32, x, y int32) *NavMeshTile {
    // 构建文件路径
    tileFileName := fmt.Sprintf("%s/mmaps/%03d%02d%02d.mmtile", 
                               mgr.dataPath, mapID, y, x)
    
    // 🔑 从文件系统读取二进制数据
    tileData, err := mgr.loadTileFromFile(tileFileName)
    if err != nil {
        // 如果真实文件不存在，创建模拟数据
        return mgr.createSimulatedTile(mapID, x, y)
    }

    // 解析和验证数据
    header, err := mgr.parseTileHeader(tileData)
    if err != nil || !mgr.validateTileData(header, tileData) {
        return nil
    }

    return &NavMeshTile{
        TileX:  x,
        TileY:  y,
        Header: header,
        Data:   tileData, // 🔑 这是传递给 Recast 的关键数据
    }
}
```

#### 2.4 瓦片数据传递给 Recast

```go
func (mgr *MMapManager) addTileToNavMesh(navMesh *C.dtNavMesh, tile *NavMeshTile) C.dtTileRef {
    // 🔑 关键步骤：Go []byte 转换为 C unsigned char*
    dataPtr := (*C.uchar)(unsafe.Pointer(&tile.Data[0]))
    dataSize := C.int(len(tile.Data))
    flags := C.int(0x01) // DT_TILE_FREE_DATA

    // 🔑 调用 C 函数，将数据传递给 Recast Navigation 库
    tileRef := C.addTileToNavMesh(navMesh, dataPtr, dataSize, flags)

    return tileRef
}
```

## 📊 参数传递的关键环节

### 1. **导航网格初始化参数**

| 参数 | 值 | 说明 |
|------|----|----|
| `orig` | `(-17066, -500, -17066)` | 地图原点坐标 |
| `tileWidth` | `533.33` | 瓦片宽度 (码) |
| `tileHeight` | `533.33` | 瓦片高度 (码) |
| `maxTiles` | `4096` | 最大瓦片数 (64×64) |
| `maxPolys` | `65536` | 每瓦片最大多边形数 |

### 2. **瓦片数据参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| `navMesh` | `*C.dtNavMesh` | 已初始化的导航网格 |
| `data` | `*C.uchar` | .mmtile 文件的二进制数据 |
| `dataSize` | `C.int` | 数据大小 (字节) |
| `flags` | `C.int` | 瓦片标志 (如内存管理) |

## 🔍 .mmtile 文件格式

### 文件结构：
```
[文件头] [多边形数据] [顶点数据] [连接数据] [细节网格] [BVH树]
```

### 关键字段：
- **Magic**: `0x4E415654` ("NAVT")
- **Version**: 通常为 7
- **坐标**: 瓦片的 X, Y 坐标
- **多边形数量**: 该瓦片包含的多边形数
- **顶点数量**: 该瓦片包含的顶点数
- **边界框**: 瓦片的 3D 边界

## ⚡ 性能优化要点

### 1. **内存管理**
```go
// 使用 DT_TILE_FREE_DATA 让 Recast 管理内存
flags := C.int(0x01) // DT_TILE_FREE_DATA
```

### 2. **缓存机制**
```go
// 缓存已加载的瓦片引用
mapData.loadedTileRefs[tileID] = tileRef

// 缓存查询器实例
mapData.navMeshQueries[instanceID] = query
```

### 3. **延迟加载**
```go
// 只在需要时加载瓦片
if _, exists := mapData.tiles[tileKey]; exists {
    return true // 瓦片已加载
}
```

## 🚨 常见问题和解决方案

### 1. **文件不存在**
```go
if err != nil {
    // 创建模拟数据用于演示
    return mgr.createSimulatedTile(mapID, x, y)
}
```

### 2. **数据格式错误**
```go
if !mgr.validateTileData(header, tileData) {
    mgr.logger.Printf("❌ 瓦片数据验证失败")
    return nil
}
```

### 3. **Recast 库拒绝数据**
```go
if tileRef == 0 {
    mgr.logger.Printf("❌ Recast 库拒绝瓦片数据: 可能格式不正确")
    return 0
}
```

## 🎯 总结

整个流程的核心是：

1. **初始化阶段**: 告诉 Recast 库地图的基本参数 (边界、瓦片大小等)
2. **数据加载阶段**: 从 `.mmtile` 文件读取二进制数据
3. **数据传递阶段**: 通过 CGO 将 Go 的 `[]byte` 转换为 C 的 `unsigned char*`
4. **库调用阶段**: 调用 `navMesh->addTile()` 将数据传递给 Recast Navigation 库

这样，Recast 库就"知道"了地图的结构，可以进行寻路计算了！