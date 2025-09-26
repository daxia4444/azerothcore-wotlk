# AzerothCore Go战斗系统 - 网络版本

## 🌟 项目概述

这是一个基于AzerothCore架构的Go语言战斗系统实现，完整模拟了魔兽世界的客户端-服务器网络通信机制。项目包含了WorldSession会话管理、操作码处理、数据包通信等核心功能。

## 🏗️ 系统架构

### 核心组件

1. **网络通信层**
   - `WorldSocket`: 网络套接字管理
   - `WorldPacket`: 数据包封装
   - `OpcodeTable`: 操作码路由表

2. **会话管理层**
   - `WorldSession`: 玩家会话管理
   - `GameClient`: 客户端模拟器
   - `GameServer`: 游戏服务器

3. **游戏逻辑层**
   - `World`: 世界管理器
   - `Unit`: 基础单位系统
   - `Player/Creature`: 玩家和生物实体

4. **战斗系统**
   - 伤害计算和处理
   - 威胁值管理
   - AI行为系统

## 📁 文件结构

```
go-combat-demo/
├── main_network.go     # 网络版主程序 (推荐使用)
├── network.go          # 网络通信核心
├── client_server.go    # 客户端和服务器实现
├── world.go           # 世界管理器
├── unit.go            # 基础单位系统
├── entities.go        # 玩家和生物实体
├── damage.go          # 伤害处理系统
├── dungeon.go         # 副本系统
└── monster_ai.go      # 怪物AI系统
```

## 🚀 运行方式

### 网络版本(推荐)
```bash
go run main_network.go network.go client_server.go world.go unit.go entities.go damage.go
```

## 🎮 功能特性

### 网络通信
- ✅ TCP套接字连接
- ✅ 数据包序列化/反序列化
- ✅ 操作码路由系统
- ✅ 会话超时管理
- ✅ 心跳包机制

### 战斗系统
- ✅ 近战攻击系统
- ✅ 法术施放系统
- ✅ 命中判定(暴击、格挡、闪避等)
- ✅ 威胁值管理
- ✅ AI行为系统

### 职业系统
- ✅ 战士(坦克) - 怒气系统
- ✅ 法师(DPS) - 法力系统
- ✅ 牧师(治疗) - 法力系统
- ✅ 猎人(远程DPS) - 法力系统
- ✅ 术士(法术DPS) - 法力系统

### 操作码支持
- `CMSG_ATTACKSWING`: 攻击指令
- `CMSG_ATTACKSTOP`: 停止攻击
- `CMSG_SET_SELECTION`: 选择目标
- `CMSG_CAST_SPELL`: 施放法术
- `CMSG_KEEP_ALIVE`: 保持连接

## 🔧 技术实现

### 基于AzerothCore的设计模式

1. **WorldSession模式**
   ```go
   type WorldSession struct {
       id           uint32
       player       *Player
       socket       *WorldSocket
       opcodeTable  *OpcodeTable
   }
   ```

2. **操作码处理**
   ```go
   type ClientOpcodeHandler struct {
       name       string
       status     int
       processing int
       handler    func(*WorldSession, *WorldPacket)
   }
   ```

3. **数据包结构**
   ```go
   type WorldPacket struct {
       opcode uint16
       data   []byte
   }
   ```

### 核心算法

1. **伤害计算**
   - 基础伤害 + 随机浮动
   - 护甲减免计算
   - 暴击倍率处理

2. **命中判定**
   - 命中率检查
   - 暴击率计算
   - 格挡/闪避判定

3. **威胁值系统**
   - 伤害威胁值
   - 治疗威胁值
   - 嘲讽机制

## 📊 演示结果

运行网络版本后，你将看到：

1. **服务器启动**
   ```
   🌐 === AzerothCore风格的客户端-服务器战斗系统演示 === 🌐
   游戏服务器已启动，监听地址: localhost:8080
   ```

2. **客户端连接**
   ```
   🎮 === 创建5个客户端连接 === 🎮
   ✅ 钢铁卫士 (战士) 已连接并登录
   ✅ 烈焰法师 (法师) 已连接并登录
   ...
   ```

3. **网络通信**
   ```
   📡 客户端 钢铁卫士 选择目标: 迪菲亚矿工
   📡 客户端 烈焰法师 发送攻击指令
   📡 客户端 圣光牧师 施放治疗术
   ```

## 🎯 项目亮点

### 1. 完整的网络架构
- 模拟了真实的MMO客户端-服务器通信
- 支持多客户端并发连接
- 实现了完整的会话生命周期管理

### 2. 忠实的AzerothCore实现
- 保持了原版的类结构和方法命名
- 实现了核心的操作码处理机制
- 模拟了真实的数据包通信流程

### 3. 可扩展的设计
- 模块化的代码结构
- 清晰的接口定义
- 易于添加新功能

### 4. 教育价值
- 详细的中文注释
- 清晰的代码结构
- 完整的演示流程

## 🔮 未来扩展

可以进一步添加的功能：
- 更多职业和技能
- 装备系统
- 组队系统
- 公会系统
- 拍卖行系统
- 任务系统

## 📝 总结

这个项目成功地将AzerothCore的复杂C++战斗系统简化为易于理解的Go语言实现，同时保持了核心架构的完整性。通过网络通信的方式，完美展示了现代MMO游戏的客户端-服务器交互模式。

项目不仅具有教育价值，也为理解大型游戏服务器架构提供了很好的参考。代码结构清晰，注释详细，是学习游戏服务器开发的优秀案例。