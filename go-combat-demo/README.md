# AzerothCore 战斗系统 Go 语言实现

这是一个用 Go 语言实现的 AzerothCore 战斗系统简化版本，旨在解决 C++ 版本中复杂的虚函数继承体系和庞大代码量的问题。

## 🎯 项目目标

- **简化复杂性**: 用 Go 的接口系统替代 C++ 的虚函数继承
- **提高可读性**: 清晰的代码结构，易于理解和维护
- **保持核心逻辑**: 完整实现 AzerothCore 的核心战斗机制
- **学习友好**: 适合学习和理解 MMO 游戏战斗系统

## 📁 项目结构

```
go-combat-demo/
├── main.go        # 主程序入口，演示战斗系统
├── unit.go        # 基础单位系统和战斗逻辑
├── entities.go    # 玩家和生物类型定义
├── damage.go      # 伤害处理系统（对应 Unit::DealDamage）
├── world.go       # 世界管理器和辅助系统
├── go.mod         # Go 模块配置
└── README.md      # 项目说明
```

## 🔧 核心功能

### 1. 基础单位系统
- **IUnit 接口**: 定义所有单位的基本行为
- **Unit 结构**: 实现基础单位功能
- **生命值/能量管理**: 完整的 HP/MP/怒气系统

### 2. 战斗系统
- **攻击机制**: 近战攻击、命中判定、伤害计算
- **战斗状态**: 进入/退出战斗、战斗计时器
- **命中结果**: 未命中、闪避、招架、格挡、暴击

### 3. 伤害处理
- **DealDamage 函数**: 完整对应 AzerothCore 的核心伤害处理逻辑
- **威胁管理**: 威胁值计算和目标选择
- **护甲减伤**: 基于等级和护甲值的伤害减免
- **伤害吸收**: 护盾和伤害吸收机制

### 4. AI 系统
- **IAI 接口**: 统一的 AI 行为定义
- **PlayerAI**: 玩家 AI（通常由玩家控制）
- **CreatureAI**: 生物 AI（自动战斗逻辑）

### 5. 辅助系统
- **世界管理器**: 管理所有游戏单位
- **战斗日志**: 记录战斗事件
- **统计系统**: 伤害、命中率、暴击率统计

## 🚀 运行方法

1. **安装 Go 环境** (版本 1.19 或更高)
   ```bash
   # 检查 Go 版本
   go version
   ```

2. **运行演示程序**
   ```bash
   cd go-combat-demo
   go run .
   ```

3. **观察战斗过程**
   程序会自动创建一个玩家和一个怪物，并模拟完整的战斗过程。

## 📊 演示输出示例

```
=== AzerothCore Combat System Demo (Go Implementation) ===
单位 TestPlayer 加入世界
单位 TestMonster 加入世界
玩家: TestPlayer (等级80, 生命值: 25000/25000)
怪物: TestMonster (等级82, 生命值: 30000/30000)

=== 开始战斗 ===
TestPlayer 开始攻击 TestMonster
TestPlayer 进入战斗状态
TestMonster 进入战斗状态
[PlayerAI] TestPlayer 开始攻击 TestMonster
[CreatureAI] TestMonster 与 TestPlayer 进入战斗

--- 第 1 轮战斗 ---
TestPlayer 对 TestMonster 造成 856 点伤害
TestPlayer 获得 8 点怒气
威胁值更新: TestPlayer 对目标的威胁值增加 856.0
TestMonster 对 TestPlayer 造成暴击！
TestMonster 对 TestPlayer 造成 1640 点伤害
...
```

## 🎮 核心概念对比

### C++ vs Go 实现对比

| 特性 | AzerothCore (C++) | Go 实现 |
|------|------------------|---------|
| 继承体系 | 复杂的虚函数继承 | 简单的接口组合 |
| 代码量 | 21,000+ 行 | ~1,500 行 |
| 可读性 | 需要深入理解继承关系 | 直观的接口定义 |
| 调试难度 | 虚函数调用难以追踪 | 明确的函数调用 |
| 学习曲线 | 陡峭 | 平缓 |

### 保持的核心逻辑

1. **Unit::DealDamage 函数逻辑**
   - 脚本钩子处理
   - AI 通知系统
   - GM 无敌检查
   - 决斗/切磋处理
   - 威胁值管理
   - 怒气奖励系统

2. **战斗状态管理**
   - 战斗开始/结束条件
   - 战斗计时器机制
   - 攻击者列表维护

3. **伤害计算系统**
   - 护甲减伤公式
   - 命中判定算法
   - 暴击伤害计算

## 🔍 学习建议

### 1. 从接口开始
```go
type IUnit interface {
    GetHealth() uint32
    IsAlive() bool
    Attack(target IUnit) bool
    DealDamage(attacker IUnit, damage uint32, damageType int, schoolMask int) uint32
}
```

### 2. 理解核心循环
```go
func (u *Unit) Update(diff uint32) {
    // 更新攻击计时器
    // 更新战斗计时器  
    // 执行攻击
    // 更新AI
}
```

### 3. 掌握伤害处理
```go
func (u *Unit) DealDamage(attacker IUnit, damage uint32, damageType int, schoolMask int) uint32 {
    // 对应 AzerothCore 的完整 DealDamage 逻辑
}
```

## 🛠️ 扩展建议

1. **添加法术系统**: 实现法术施放、法术伤害、法术抗性
2. **完善光环系统**: 添加 Buff/Debuff 效果
3. **实现装备系统**: 武器、护甲对战斗的影响
4. **添加技能系统**: 不同职业的特殊技能
5. **网络支持**: 多玩家在线战斗

## 📚 相关资源

- [AzerothCore 官方文档](https://www.azerothcore.org/)
- [Go 语言官方文档](https://golang.org/doc/)
- [游戏服务器架构设计](https://github.com/topics/game-server)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目！

## 📄 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。

---

**注意**: 这是一个教育性质的简化实现，主要用于学习和理解 AzerothCore 战斗系统的核心机制。在生产环境中使用需要进一步的优化和完善。