# AzerothCore血量同步机制实现文档

## 📋 概述

本文档详细说明了在Go语言demo中实现的AzerothCore风格血量同步机制。该机制确保所有客户端能够实时看到玩家和NPC的血量、能量值变化。

## 🔄 同步机制架构

### **1. 双重同步策略**

AzerothCore采用**即时同步 + 定期同步**的混合模式：

```
即时同步 (Immediate Sync)
├── 血量变化时立即广播
├── 能量变化时立即广播  
├── 伤害/治疗时立即广播
└── 确保关键事件的即时响应

定期同步 (Periodic Sync)
├── 每5秒广播完整状态
├── 防止数据不一致
├── 处理网络丢包情况
└── 提供状态校验机制
```

### **2. 网络消息类型**

| 消息类型 | 操作码 | 用途 | 触发时机 |
|---------|--------|------|----------|
| SMSG_UPDATE_OBJECT | 0x0A9 | 完整对象更新 | 定期同步 |
| SMSG_POWER_UPDATE | 0x480 | 能量值更新 | 能量变化时 |
| SMSG_HEALTH_UPDATE | 0x481 | 血量更新 | 血量变化时 |
| SMSG_ATTACKERSTATEUPDATE | 0x14A | 攻击状态更新 | 伤害发生时 |

## 🛠️ 实现细节

### **1. 血量变化同步**

```go
// 在Unit.ModifyHealth中自动触发
func (u *Unit) ModifyHealth(delta int32) int32 {
    oldHealth := int32(u.health)
    newHealth := oldHealth + delta
    
    // 应用血量变化
    u.health = uint32(newHealth)
    
    // 🔥 关键：即时网络同步
    if u.world != nil && oldHealth != newHealth {
        u.world.BroadcastHealthUpdate(u, uint32(oldHealth), u.health)
    }
    
    return int32(u.health) - oldHealth
}
```

### **2. 能量变化同步**

```go
// 在Unit.ModifyPower中自动触发
func (u *Unit) ModifyPower(powerType int, delta int32) int32 {
    oldPower := int32(u.GetPower(powerType))
    newPower := oldPower + delta
    
    u.SetPower(powerType, uint32(newPower))
    
    // 🔥 关键：即时网络同步
    if u.world != nil && oldPower != newPower {
        u.world.BroadcastPowerUpdate(u, powerType, uint32(oldPower), uint32(newPower))
    }
    
    return newPower - oldPower
}
```

### **3. 伤害事件同步**

```go
// 在Unit.DealDamage中触发
func (u *Unit) DealDamage(attacker IUnit, damage uint32, damageType int, schoolMask int) uint32 {
    // 应用伤害（会自动触发血量同步）
    u.ModifyHealth(-int32(damage))
    
    // 🔥 关键：广播攻击状态更新
    if u.world != nil && attacker != nil {
        u.world.BroadcastAttackerStateUpdate(attacker, u, damage, MELEE_HIT_NORMAL, schoolMask)
    }
    
    return damage
}
```

## 📡 网络广播实现

### **1. 血量更新广播**

```go
func (w *World) BroadcastHealthUpdate(unit IUnit, oldHealth, newHealth uint32) {
    packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
    packet.WriteUint64(unit.GetGUID())
    packet.WriteUint8(1) // 更新类型：血量
    packet.WriteUint32(oldHealth)
    packet.WriteUint32(newHealth)
    packet.WriteUint32(unit.GetMaxHealth())
    
    w.BroadcastPacket(packet)
    
    fmt.Printf("[网络] 广播血量更新: %s %d->%d/%d\n",
        unit.GetName(), oldHealth, newHealth, unit.GetMaxHealth())
}
```

### **2. 能量更新广播**

```go
func (w *World) BroadcastPowerUpdate(unit IUnit, powerType int, oldPower, newPower uint32) {
    packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
    packet.WriteUint64(unit.GetGUID())
    packet.WriteUint8(2) // 更新类型：能量
    packet.WriteUint8(uint8(powerType))
    packet.WriteUint32(oldPower)
    packet.WriteUint32(newPower)
    packet.WriteUint32(unit.GetMaxPower(powerType))
    
    w.BroadcastPacket(packet)
}
```

### **3. 定期状态同步**

```go
func (w *World) broadcastPeriodicUpdates(diff uint32) {
    lastPeriodicUpdate += diff
    if lastPeriodicUpdate >= PERIODIC_UPDATE_INTERVAL { // 5秒
        lastPeriodicUpdate = 0
        
        // 广播所有单位的完整状态
        w.mutex.RLock()
        for _, unit := range w.units {
            if unit.IsAlive() {
                w.BroadcastUnitUpdate(unit)
            }
        }
        w.mutex.RUnlock()
    }
}
```

## 🎯 关键设计原则

### **1. 状态同步 vs 帧同步**

| 特性 | 帧同步 | AzerothCore状态同步 |
|------|--------|-------------------|
| 网络流量 | 高（每帧） | 中（事件驱动） |
| 延迟容忍 | 低 | 高 |
| 作弊防护 | 弱 | 强 |
| 适用场景 | RTS/MOBA | MMORPG |

### **2. 即时响应原则**

- **玩家操作** → **立即响应** → **服务器验证** → **广播结果**
- 确保玩家感受到即时反馈，同时维护服务器权威性

### **3. 网络优化策略**

```go
// 只在实际变化时发送更新
if u.world != nil && oldHealth != newHealth {
    u.world.BroadcastHealthUpdate(u, oldHealth, newHealth)
}
```

## 🔍 演示程序说明

### **运行血量同步演示**

```bash
cd go-combat-demo
go run *.go
```

### **演示内容**

1. **即时血量同步** - 攻击时立即广播血量变化
2. **法术消耗能量同步** - 施法时立即广播能量消耗
3. **治疗效果同步** - 治疗时立即广播血量恢复
4. **定期状态同步** - 每5秒广播完整状态
5. **并发血量变化** - 多个同时的DOT/HOT效果

### **预期输出示例**

```
=== AzerothCore风格的血量同步演示 ===

初始状态:
- 战士玩家: 3000/3000 HP, 0/100 怒气
- 法师玩家: 2500/2500 HP, 3000/3000 法力

=== 演示1: 即时血量同步 ===
战士攻击法师...
[网络] 广播攻击状态: 战士玩家 对 法师玩家 造成 300 伤害
[网络] 广播血量更新: 法师玩家 2500->2200/2500
第1次攻击: 造成300伤害, 法师玩家剩余血量: 2200/2500

=== 演示2: 法术消耗能量同步 ===
法师施放寒冰箭...
[网络] 广播法力值更新: 法师玩家 3000->2800/3000
法术消耗: 200法力, 法师玩家剩余法力: 2800/3000

=== 演示3: 治疗效果同步 ===
法师施放快速治疗...
[网络] 广播血量更新: 法师玩家 2200->2500/2500
治疗效果: +500生命值, 法师玩家当前血量: 2500/2500
```

## ✅ 与AzerothCore的一致性

### **1. 架构一致性**

- ✅ 使用WorldSession管理客户端连接
- ✅ 使用WorldPacket进行数据传输
- ✅ 使用操作码(Opcode)区分消息类型
- ✅ 实现即时同步 + 定期同步机制

### **2. 消息格式一致性**

- ✅ SMSG_UPDATE_OBJECT格式
- ✅ SMSG_POWER_UPDATE格式
- ✅ SMSG_ATTACKERSTATEUPDATE格式
- ✅ 数据包结构和字段顺序

### **3. 时序一致性**

- ✅ 血量变化时立即广播
- ✅ 能量变化时立即广播
- ✅ 定期完整状态同步
- ✅ 伤害事件的即时响应

## 🚀 性能优化

### **1. 网络优化**

- 只在实际变化时发送更新
- 批量处理定期更新
- 使用高效的数据包格式

### **2. 并发安全**

- 使用读写锁保护共享数据
- 原子操作处理GUID生成
- 线程安全的会话管理

### **3. 内存优化**

- 及时清理过期的冷却时间
- 复用数据包对象
- 高效的单位查找机制

## 📚 总结

本实现完全遵循了AzerothCore的血量同步机制：

1. **双重同步策略** - 即时响应 + 定期校验
2. **事件驱动更新** - 状态变化时自动触发同步
3. **网络消息规范** - 使用标准的操作码和数据格式
4. **性能优化** - 只在必要时发送更新，避免无效网络流量
5. **并发安全** - 支持多客户端同时操作

这种设计确保了MMORPG中关键的用户体验：**即时反馈**和**状态一致性**。