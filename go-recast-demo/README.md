# AzerothCore Recast Navigation Go Demo

## 项目概述

这是一个基于真实 Recast Navigation 库的 Go 语言演示程序，完全符合 AzerothCore 项目的导航实现逻辑。

## 文件说明

### 核心文件
- **`real_azerothcore_demo.go`** - 主要演示程序，使用真实的 Recast Navigation C++ 库
- **`build_real_demo.sh`** - 构建脚本，用于编译 Recast Navigation 库和 Go 程序
- **`REAL_IMPLEMENTATION.md`** - 详细的实现文档和使用说明

### 备份文件
- **`high_precision_demo.go.bak`** - 已移除的纯 Go 模拟实现备份

## 为什么只保留真实实现？

1. **技术一致性** - 使用与 AzerothCore 完全相同的 Recast Navigation C++ 库
2. **架构一致性** - 采用相同的类名、方法名和数据结构
3. **性能一致性** - 具有与真实系统相同的性能特征
4. **避免混淆** - 防止开发者误解 AzerothCore 的实际实现方式

## 主要特性

### 🏰 AzerothCore 兼容性
- 完全兼容 AzerothCore 的 `PathGenerator` 类
- 支持 `MMapManager` 地图管理
- 使用真实的 `.mmtile` 文件格式
- 支持多实例查询器

### 🧭 导航功能
- **路径查找** - 基于 A* 算法的多边形路径查找
- **障碍物检测** - 射线检测和碰撞检测
- **可行走性检查** - 地形和坡度验证
- **路径优化** - 直线路径和平滑路径

### ⚡ 性能优化
- **并发安全** - 支持多线程访问
- **内存管理** - 自动资源清理
- **性能监控** - 内置统计和性能分析
- **错误处理** - 完善的错误处理机制

## 快速开始

### 1. 编译依赖
```bash
# 运行构建脚本
./build_real_demo.sh
```

### 2. 运行演示
```bash
# 编译并运行
go run real_azerothcore_demo.go
```

### 3. 查看结果
程序将演示以下功能：
- 暴风城内部寻路
- 可行走性检查
- 障碍物检测
- 坡度检查
- 性能测试

## 技术架构

```
AzerothCore Go Demo
├── MMapManager (地图管理器)
│   ├── 地图数据加载
│   ├── 瓦片管理
│   └── 查询器管理
├── PathGenerator (路径生成器)
│   ├── 多边形路径构建
│   ├── 点路径构建
│   ├── 障碍物检测
│   └── 可行走性检查
└── Recast Navigation (C++ 库)
    ├── dtNavMesh
    ├── dtNavMeshQuery
    └── dtQueryFilter
```

## 与 AzerothCore 的对应关系

| Go 实现 | AzerothCore 对应 |
|---------|------------------|
| `MMapManager` | `MMapMgr` |
| `PathGenerator` | `PathGenerator` |
| `CalculatePath()` | `CalculatePath()` |
| `BuildPolyPath()` | `BuildPolyPath()` |
| `BuildPointPath()` | `BuildPointPath()` |
| `IsWalkable()` | 可行走性检查 |
| `Raycast()` | 射线检测 |

## 开发说明

### 代码风格
- 遵循 Go 语言标准格式
- 使用英文注释和文档
- 保持与 AzerothCore 命名一致

### 错误处理
- 完善的错误检查和处理
- 详细的日志记录
- 资源自动清理

### 性能考虑
- 并发安全的设计
- 内存池和对象复用
- 性能统计和监控

## 注意事项

1. **依赖要求** - 需要编译 Recast Navigation C++ 库
2. **平台兼容** - 主要支持 Linux 平台
3. **内存管理** - 注意 C/Go 内存边界
4. **线程安全** - 支持多线程并发访问

## 更新日志

### v1.0 (当前版本)
- ✅ 移除纯 Go 模拟实现
- ✅ 专注于真实 Recast Navigation 实现
- ✅ 优化错误处理和日志记录
- ✅ 添加性能监控和统计
- ✅ 完善代码注释和文档

## 贡献指南

1. 保持与 AzerothCore 实现的一致性
2. 添加充分的测试用例
3. 遵循项目的代码风格
4. 更新相关文档

## 许可证

本项目遵循与 AzerothCore 相同的开源许可证。