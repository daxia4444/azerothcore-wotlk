package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// 常量定义 - 游戏战斗系统核心常量
const (
	// 单位类型 - 区分不同的游戏实体
	UNIT_TYPE_PLAYER   = 1 // 玩家角色
	UNIT_TYPE_CREATURE = 2 // NPC生物(怪物、商人等)

	// 职业类型 - 魔兽世界经典职业系统
	CLASS_WARRIOR = 1  // 战士 - 近战坦克职业，使用怒气
	CLASS_PALADIN = 2  // 圣骑士 - 坦克/治疗混合职业，使用法力
	CLASS_HUNTER  = 3  // 猎人 - 远程物理DPS，使用集中值
	CLASS_ROGUE   = 4  // 盗贼 - 近战敏捷DPS，使用能量
	CLASS_PRIEST  = 5  // 牧师 - 治疗/法术DPS，使用法力
	CLASS_MAGE    = 8  // 法师 - 远程法术DPS，使用法力
	CLASS_WARLOCK = 9  // 术士 - 远程法术DPS，使用法力
	CLASS_DRUID   = 11 // 德鲁伊 - 多形态混合职业，使用法力

	// 生物类型 - NPC怪物的种族分类，影响法术效果
	CREATURE_TYPE_BEAST      = 1 // 野兽 - 动物类生物
	CREATURE_TYPE_DRAGONKIN  = 2 // 龙类 - 龙族生物
	CREATURE_TYPE_DEMON      = 3 // 恶魔 - 来自扭曲虚空的邪恶生物
	CREATURE_TYPE_ELEMENTAL  = 4 // 元素 - 火水土气元素生物
	CREATURE_TYPE_GIANT      = 5 // 巨人 - 大型人形生物
	CREATURE_TYPE_UNDEAD     = 6 // 亡灵 - 不死生物，免疫某些效果
	CREATURE_TYPE_HUMANOID   = 7 // 人形 - 人类、精灵、兽人等智慧种族
	CREATURE_TYPE_CRITTER    = 8 // 小动物 - 无害的小生物
	CREATURE_TYPE_MECHANICAL = 9 // 机械 - 工程学造物，免疫心智控制

	// 能量类型 - 不同职业使用的资源系统
	POWER_MANA   = 0 // 法力值 - 法师、牧师、术士、德鲁伊使用
	POWER_RAGE   = 1 // 怒气值 - 战士使用，战斗中积累
	POWER_FOCUS  = 2 // 集中值 - 猎人使用，缓慢恢复
	POWER_ENERGY = 3 // 能量值 - 盗贼使用，快速恢复

	// 攻击类型 - 不同的攻击方式
	BASE_ATTACK   = 0 // 主手攻击 - 主武器攻击
	OFF_ATTACK    = 1 // 副手攻击 - 副手武器攻击(双持)
	RANGED_ATTACK = 2 // 远程攻击 - 弓箭、枪械、法杖攻击

	// 伤害类型 - 伤害的来源和性质（主要定义在damage.go中）
	DOT  = 3 // 持续伤害 - 毒素、燃烧等持续效果
	HEAL = 4 // 治疗效果 - 恢复生命值

	// 法术学派 - 魔法伤害的类型，影响抗性计算(位掩码)
	SPELL_SCHOOL_NORMAL = 1  // 物理伤害 - 武器攻击等物理伤害
	SPELL_SCHOOL_HOLY   = 2  // 神圣伤害 - 圣骑士、牧师的神圣法术
	SPELL_SCHOOL_FIRE   = 4  // 火焰伤害 - 火球术、燃烧等火系法术
	SPELL_SCHOOL_NATURE = 8  // 自然伤害 - 闪电、毒素等自然法术
	SPELL_SCHOOL_FROST  = 16 // 冰霜伤害 - 暴风雪、冰箭等冰系法术
	SPELL_SCHOOL_SHADOW = 32 // 暗影伤害 - 术士、牧师的暗影法术
	SPELL_SCHOOL_ARCANE = 64 // 奥术伤害 - 法师的奥术法术

	// 单位状态 - 使用位掩码表示各种状态，可以同时拥有多个状态
	UNIT_STATE_DIED            = 0x00000001 // 死亡状态 - 单位已死亡
	UNIT_STATE_MELEE_ATTACKING = 0x00000002 // 近战攻击中 - 正在进行近战攻击
	UNIT_STATE_CHARMED         = 0x00000400 // 魅惑状态 - 被敌人控制
	UNIT_STATE_STUNNED         = 0x00000800 // 昏迷状态 - 无法行动
	UNIT_STATE_ROOTED          = 0x00001000 // 定身状态 - 无法移动但可以攻击
	UNIT_STATE_CONFUSED        = 0x00002000 // 混乱状态 - 随机移动攻击
	UNIT_STATE_DISTRACTED      = 0x00004000 // 分心状态 - 注意力被分散
	UNIT_STATE_ISOLATED        = 0x00008000 // 孤立状态 - 与队友分离
	UNIT_STATE_ATTACK_PLAYER   = 0x00010000 // 攻击玩家 - 正在攻击玩家角色
	UNIT_STATE_CASTING         = 0x00020000 // 施法状态 - 正在施放法术
	UNIT_STATE_POSSESSED       = 0x00040000 // 被附身 - 被其他意识控制
	UNIT_STATE_CHARGING        = 0x00080000 // 冲锋状态 - 正在冲向目标
	UNIT_STATE_JUMPING         = 0x00100000 // 跳跃状态 - 正在跳跃
	UNIT_STATE_MOVE            = 0x00200000 // 移动状态 - 正在移动
	UNIT_STATE_ROTATING        = 0x00400000 // 转向状态 - 正在转向
	UNIT_STATE_EVADE           = 0x00800000 // 闪避状态 - 回避攻击
	UNIT_STATE_ROAMING         = 0x01000000 // 漫游状态 - 自由移动巡逻
	UNIT_STATE_IN_FLIGHT       = 0x02000000 // 飞行状态 - 在空中飞行
	UNIT_STATE_FOLLOW          = 0x04000000 // 跟随状态 - 跟随其他单位
	UNIT_STATE_ROOT            = 0x08000000 // 束缚状态 - 被法术束缚无法移动
	UNIT_STATE_FLEEING         = 0x10000000 // 逃跑状态 - 正在逃离战斗
	UNIT_STATE_IN_COMBAT       = 0x20000000 // 战斗状态 - 处于战斗中

	// 单位标志 - 额外的状态标记
	UNIT_FLAG_IN_COMBAT     = 0x00080000 // 战斗标志 - 标记单位处于战斗状态
	UNIT_FLAG_PET_IN_COMBAT = 0x00100000 // 宠物战斗标志 - 宠物处于战斗状态

	// 战斗计时器 - 控制战斗状态的持续时间(毫秒)
	COMBAT_TIMER_PVP = 5500 // PvP战斗计时器 (5.5秒) - 玩家对战后的战斗状态持续时间
	COMBAT_TIMER_PVE = 5000 // PvE战斗计时器 (5秒) - 对怪物战斗后的战斗状态持续时间

	// 攻击计时器 - 控制攻击频率(毫秒)
	BASE_ATTACK_TIME = 2000 // 基础攻击间隔 (2秒) - 默认的攻击速度

	// 战斗距离 - 近战攻击的有效范围(码)
	MIN_MELEE_REACH = 1.5 // 最小近战范围 - 近战攻击的最小距离

	// 命中结果 - 攻击的各种可能结果（主要定义在damage.go中）
	MELEE_HIT_CRUSHING = 7 // 碾压 - 高等级对低等级的强力攻击

)

// 基础单位接口
type IUnit interface {
	// 基础属性
	GetGUID() uint64
	GetName() string
	GetLevel() uint8
	SetLevel(level uint8)

	// 生命值
	GetHealth() uint32
	GetMaxHealth() uint32
	SetHealth(health uint32)
	SetMaxHealth(maxHealth uint32)
	ModifyHealth(delta int32) int32
	IsAlive() bool

	// 能量值
	GetPower(powerType uint8) uint32
	GetMaxPower(powerType uint8) uint32
	SetPower(powerType uint8, power uint32)
	SetMaxPower(powerType uint8, maxPower uint32)
	ModifyPower(powerType uint8, delta int32) int32

	// 战斗相关
	IsInCombat() bool
	SetInCombat(inCombat bool)
	Attack(target IUnit) bool
	AttackStop()
	GetVictim() IUnit
	SetVictim(victim IUnit)

	// 更新
	Update(diff uint32)

	// 伤害处理
	DealDamage(attacker IUnit, damage uint32, damageType int, schoolMask int) uint32

	// 位置和距离
	GetX() float32
	GetY() float32
	GetZ() float32
	GetPosition() (float32, float32, float32) // 新增：获取位置坐标
	GetDistanceTo(target IUnit) float32
	IsWithinMeleeRange(target IUnit) bool

	// 状态
	HasUnitState(state uint32) bool
	AddUnitState(state uint32)
	ClearUnitState(state uint32)

	// AI
	GetAI() IAI
	SetAI(ai IAI)

	// 网络支持
	SetGUID(guid uint64)
	IsValidAttackTarget(target IUnit) bool
	SetTarget(target IUnit)
	GetTarget() IUnit
	CastSpell(target IUnit, spellId uint32)
	Heal(caster IUnit, amount uint32)
}

// AI接口
type IAI interface {
	UpdateAI(diff uint32)
	AttackStart(target IUnit)
	EnterCombat(target IUnit)
	JustDied(killer IUnit)
	DamageTaken(attacker IUnit, damage uint32)
	DamageDealt(victim IUnit, damage uint32)
}

// 基础单位结构
type Unit struct {
	// 基础标识信息
	guid     uint64 // 全局唯一标识符，用于区分不同的单位实例
	name     string // 单位名称，显示给玩家看的名字
	level    uint8  // 单位等级，影响属性和战斗力
	unitType int    // 单位类型，区分玩家(UNIT_TYPE_PLAYER)和生物(UNIT_TYPE_CREATURE)

	// 生命值系统
	health    uint32 // 当前生命值，降到0时单位死亡
	maxHealth uint32 // 最大生命值，生命值上限

	// 能量值系统 - 不同职业使用不同的能量类型(法力、怒气、能量、集中值)
	powers    map[uint8]uint32 // 当前各种能量值(法力、怒气、能量、集中值)
	maxPowers map[uint8]uint32 // 各种能量值的上限

	// 3D世界坐标位置
	x, y, z     float32 // 单位在游戏世界中的三维坐标，用于距离计算和移动
	orientation float32 // 单位朝向，用于移动和攻击方向

	// 战斗状态管理
	inCombat    bool             // 是否处于战斗状态，影响回血回蓝和其他机制
	combatTimer uint32           // 战斗计时器，用于判断何时退出战斗状态
	victim      IUnit            // 当前攻击目标，近战攻击的对象
	attackers   map[uint64]IUnit // 正在攻击此单位的敌人列表，key为攻击者GUID
	attackTimer map[int]int32    // 各种攻击类型的冷却计时器(主手、副手、远程)

	// 单位状态标志位
	unitState uint32 // 位掩码，记录各种状态(死亡、眩晕、定身、施法等)

	// 人工智能系统
	ai IAI // AI控制器，负责单位的自动行为(攻击、移动、技能使用等)

	// 仇恨威胁系统
	threatManager *ThreatManager // 威胁值管理器，用于确定攻击优先级和仇恨列表

	// 目标系统
	target IUnit // 当前选择的目标，用于技能施放和交互

	// 法术系统 - 基于AzerothCore的法术管理
	currentSpells  map[int]*Spell       // 当前施法中的法术，key为法术类型(CURRENT_GENERIC_SPELL等)
	spellCooldowns map[uint32]time.Time // 法术冷却时间，key为法术ID，value为冷却结束时间
	world          *World               // 世界引用，用于法术系统
}

// 创建基础单位
func NewUnit(guid uint64, name string, level uint8, unitType int) *Unit {
	unit := &Unit{
		guid:           guid,
		name:           name,
		level:          level,
		unitType:       unitType,
		powers:         make(map[uint8]uint32),
		maxPowers:      make(map[uint8]uint32),
		attackers:      make(map[uint64]IUnit),
		attackTimer:    make(map[int]int32),
		threatManager:  NewThreatManager(),
		currentSpells:  make(map[int]*Spell),
		spellCooldowns: make(map[uint32]time.Time),
	}

	// 初始化攻击计时器
	unit.attackTimer[BASE_ATTACK] = 0
	unit.attackTimer[OFF_ATTACK] = 0
	unit.attackTimer[RANGED_ATTACK] = 0

	return unit
}

// 实现IUnit接口
func (u *Unit) GetGUID() uint64      { return u.guid }
func (u *Unit) GetName() string      { return u.name }
func (u *Unit) GetLevel() uint8      { return u.level }
func (u *Unit) SetLevel(level uint8) { u.level = level }

func (u *Unit) GetHealth() uint32    { return u.health }
func (u *Unit) GetMaxHealth() uint32 { return u.maxHealth }
func (u *Unit) SetHealth(health uint32) {
	oldHealth := u.health
	if health > u.maxHealth {
		u.health = u.maxHealth
	} else {
		u.health = health
	}

	// 网络同步 - 基于AzerothCore的即时血量同步
	if u.world != nil && oldHealth != u.health {
		u.world.BroadcastHealthUpdate(u, oldHealth, u.health)

		// 添加到批量更新 - 基于AzerothCore的UpdateData机制
		updateBlock := u.buildHealthUpdateBlock(oldHealth, u.health)
		players := u.world.GetPlayersInRange(u.x, u.y, u.z, 100.0)
		for _, player := range players {
			u.world.AddBatchUpdate(u, player.id, updateBlock)
		}
	}
}

func (u *Unit) SetMaxHealth(maxHealth uint32) { u.maxHealth = maxHealth }

func (u *Unit) ModifyHealth(delta int32) int32 {
	oldHealth := int32(u.health)
	newHealth := oldHealth + delta

	if newHealth < 0 {
		newHealth = 0
	} else if newHealth > int32(u.maxHealth) {
		newHealth = int32(u.maxHealth)
	}

	u.health = uint32(newHealth)

	// 网络同步 - 基于AzerothCore的即时血量同步
	if u.world != nil && oldHealth != newHealth {
		u.world.BroadcastHealthUpdate(u, uint32(oldHealth), u.health)
	}

	// 如果生命值降到0，处理死亡
	if u.health == 0 && oldHealth > 0 {
		u.setDeathState()
	}

	return int32(u.health) - oldHealth
}

func (u *Unit) IsAlive() bool {
	return u.health > 0
}

func (u *Unit) GetPower(powerType uint8) uint32 {
	if power, exists := u.powers[powerType]; exists {
		return power
	}
	return 0
}

func (u *Unit) GetMaxPower(powerType uint8) uint32 {
	if maxPower, exists := u.maxPowers[powerType]; exists {
		return maxPower
	}
	return 0
}

func (u *Unit) SetPower(powerType uint8, power uint32) {
	oldPower := u.GetPower(powerType)
	maxPower := u.GetMaxPower(powerType)
	if power > maxPower {
		power = maxPower
	}
	u.powers[powerType] = power

	// 网络同步 - 基于AzerothCore的SMSG_POWER_UPDATE
	if u.world != nil && oldPower != power {
		u.world.BroadcastPowerUpdate(u, powerType, oldPower, power)

		// 添加到批量更新 - 基于AzerothCore的UpdateData机制
		updateBlock := u.buildPowerUpdateBlock(powerType, oldPower, power)
		players := u.world.GetPlayersInRange(u.x, u.y, u.z, 100.0)
		for _, player := range players {
			u.world.AddBatchUpdate(u, player.id, updateBlock)
		}
	}
}

func (u *Unit) SetMaxPower(powerType uint8, maxPower uint32) {
	u.maxPowers[powerType] = maxPower
}

func (u *Unit) ModifyPower(powerType uint8, delta int32) int32 {
	oldPower := int32(u.GetPower(powerType))
	newPower := oldPower + delta
	maxPower := int32(u.GetMaxPower(powerType))

	if newPower < 0 {
		newPower = 0
	} else if newPower > maxPower {
		newPower = maxPower
	}

	u.SetPower(powerType, uint32(newPower))

	// 网络同步 - 基于AzerothCore的SMSG_POWER_UPDATE
	if u.world != nil && oldPower != newPower {
		u.world.BroadcastPowerUpdate(u, powerType, uint32(oldPower), uint32(newPower))
	}

	return newPower - oldPower
}

func (u *Unit) IsInCombat() bool {
	return u.inCombat
}

func (u *Unit) SetInCombat(inCombat bool) {
	u.inCombat = inCombat
	if inCombat {
		u.combatTimer = COMBAT_TIMER_PVE
		u.AddUnitState(UNIT_STATE_IN_COMBAT)
	} else {
		u.combatTimer = 0
		u.ClearUnitState(UNIT_STATE_IN_COMBAT)
	}
}

func (u *Unit) GetVictim() IUnit {
	return u.victim
}

func (u *Unit) SetVictim(victim IUnit) {
	u.victim = victim
}

func (u *Unit) GetX() float32 { return u.x }
func (u *Unit) GetY() float32 { return u.y }
func (u *Unit) GetZ() float32 { return u.z }

func (u *Unit) GetPosition() (float32, float32, float32) {
	return u.x, u.y, u.z
}

// SetPosition 设置单位位置
func (u *Unit) SetPosition(x, y, z float32) {
	u.x = x
	u.y = y
	u.z = z
}

func (u *Unit) GetDistanceTo(target IUnit) float32 {
	dx := u.x - target.GetX()
	dy := u.y - target.GetY()
	dz := u.z - target.GetZ()
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

func (u *Unit) IsWithinMeleeRange(target IUnit) bool {
	distance := u.GetDistanceTo(target)
	return distance <= MIN_MELEE_REACH+2.0 // 加上一些容错范围
}

func (u *Unit) HasUnitState(state uint32) bool {
	return (u.unitState & state) != 0
}

func (u *Unit) AddUnitState(state uint32) {
	u.unitState |= state
}

func (u *Unit) ClearUnitState(state uint32) {
	u.unitState &= ^state
}

func (u *Unit) GetAI() IAI {
	return u.ai
}

func (u *Unit) SetAI(ai IAI) {
	u.ai = ai
}

// 攻击方法
func (u *Unit) Attack(target IUnit) bool {
	if target == nil || !target.IsAlive() || !u.IsAlive() {
		return false
	}

	// 检查是否在近战范围内
	if !u.IsWithinMeleeRange(target) {
		fmt.Printf("%s 距离 %s 太远，无法攻击\n", u.name, target.GetName())
		return false
	}

	// 设置受害者
	u.SetVictim(target)

	// 开始战斗
	u.CombatStart(target)

	// 通知AI
	if u.ai != nil {
		u.ai.AttackStart(target)
	}

	return true
}

func (u *Unit) AttackStop() {
	u.SetVictim(nil)
	u.ClearUnitState(UNIT_STATE_MELEE_ATTACKING)
}

// 开始战斗
func (u *Unit) CombatStart(target IUnit) {
	if !u.IsInCombat() {
		u.SetInCombat(true)
		fmt.Printf("%s 进入战斗状态\n", u.name)
	}

	// 目标也进入战斗
	if !target.IsInCombat() {
		target.SetInCombat(true)
		fmt.Printf("%s 进入战斗状态\n", target.GetName())
	}

	// 添加到攻击者列表
	if targetUnit, ok := target.(*Unit); ok {
		targetUnit.attackers[u.guid] = IUnit(u)
	}

	// 同时将目标添加到自己的攻击者列表（用于互相攻击）
	if selfUnit, ok := IUnit(u).(*Unit); ok {
		selfUnit.attackers[target.GetGUID()] = target
	}

	// 通知AI进入战斗
	if u.ai != nil {
		u.ai.EnterCombat(target)
	}
	if target.GetAI() != nil {
		target.GetAI().EnterCombat(u)
	}
}

// 处理死亡
func (u *Unit) setDeathState() {
	u.AddUnitState(UNIT_STATE_DIED)
	u.AttackStop()

	// 清除战斗状态
	u.SetInCombat(false)

	// 通知AI死亡
	if u.ai != nil {
		// 找到杀死者
		var killer IUnit
		for _, attacker := range u.attackers {
			if attacker.IsAlive() {
				killer = attacker
				break
			}
		}
		u.ai.JustDied(killer)
	}

	fmt.Printf("%s 死亡了\n", u.name)
}

// 更新方法
func (u *Unit) Update(diff uint32) {
	// 更新攻击计时器
	for attackType := range u.attackTimer {
		if u.attackTimer[attackType] > 0 {
			u.attackTimer[attackType] -= int32(diff)
		}
	}

	// 更新战斗计时器
	if u.inCombat && u.combatTimer > 0 {
		if u.combatTimer > diff {
			u.combatTimer -= diff
		} else {
			u.combatTimer = 0
			// 检查是否应该退出战斗
			if len(u.attackers) == 0 && u.victim == nil {
				u.SetInCombat(false)
				fmt.Printf("%s 退出战斗状态\n", u.name)
			}
		}
	}

	// 更新法术系统 - 基于AzerothCore的法术更新逻辑
	u.updateSpells(diff)

	// 执行攻击
	if u.victim != nil && u.IsAlive() && u.victim.IsAlive() {
		if u.attackTimer[BASE_ATTACK] <= 0 {
			u.performMeleeAttack(u.victim)
			u.attackTimer[BASE_ATTACK] = BASE_ATTACK_TIME
		}
	}

	// 更新AI
	if u.ai != nil {
		u.ai.UpdateAI(diff)
	}
}

// 执行近战攻击
func (u *Unit) performMeleeAttack(target IUnit) {
	if !u.IsWithinMeleeRange(target) {
		return
	}

	// 计算伤害
	damage := u.calculateMeleeDamage(target)

	// 计算命中结果
	hitResult := u.rollMeleeHitResult(target)

	switch hitResult {
	case MELEE_HIT_MISS:
		fmt.Printf("%s 攻击 %s 未命中\n", u.name, target.GetName())
		return
	case MELEE_HIT_DODGE:
		fmt.Printf("%s 攻击 %s 被闪避\n", u.name, target.GetName())
		return
	case MELEE_HIT_PARRY:
		fmt.Printf("%s 攻击 %s 被招架\n", u.name, target.GetName())
		return
	case MELEE_HIT_BLOCK:
		damage = damage / 2 // 格挡减少50%伤害
		fmt.Printf("%s 攻击 %s 被格挡，伤害减少\n", u.name, target.GetName())
	case MELEE_HIT_CRITICAL:

		damage = damage * 2 // 暴击双倍伤害
		fmt.Printf("%s 对 %s 造成暴击！\n", u.name, target.GetName())
	}

	// 造成伤害
	actualDamage := target.DealDamage(u, damage, DIRECT_DAMAGE, SPELL_SCHOOL_NORMAL)

	if actualDamage > 0 {
		fmt.Printf("%s 对 %s 造成 %d 点伤害\n", u.name, target.GetName(), actualDamage)

		// 奖励怒气（如果是战士）
		if unitSelf, ok := IUnit(u).(*Unit); ok {
			unitSelf.rewardRage(actualDamage, true)
		}
		if targetUnit, ok := target.(*Unit); ok {
			targetUnit.rewardRage(actualDamage, false)
		}
	}
}

// 计算近战伤害
func (u *Unit) calculateMeleeDamage(target IUnit) uint32 {
	// 基础伤害基于等级
	baseDamage := float32(u.level) * 10.0

	// 添加一些随机性
	variance := baseDamage * 0.3 // 30%的变化范围
	damage := baseDamage + (rand.Float32()-0.5)*2*variance

	if damage < 1 {
		damage = 1
	}

	return uint32(damage)
}

// 计算命中结果
func (u *Unit) rollMeleeHitResult(target IUnit) int {
	roll := rand.Float32() * 100

	// 简化的命中计算
	missChance := float32(5.0)  // 5%未命中
	dodgeChance := float32(5.0) // 5%闪避
	parryChance := float32(5.0) // 5%招架
	blockChance := float32(5.0) // 5%格挡
	critChance := float32(5.0)  // 5%暴击

	if roll < missChance {
		return MELEE_HIT_MISS
	}
	roll -= missChance

	if roll < dodgeChance {
		return MELEE_HIT_DODGE
	}
	roll -= dodgeChance

	if roll < parryChance {
		return MELEE_HIT_PARRY
	}
	roll -= parryChance

	if roll < blockChance {
		return MELEE_HIT_BLOCK
	}
	roll -= blockChance

	if roll < critChance {
		return MELEE_HIT_CRITICAL

	}

	return MELEE_HIT_NORMAL
}

// 奖励怒气
func (u *Unit) rewardRage(damage uint32, isAttacker bool) {
	// 只有战士职业才有怒气
	if player, ok := IUnit(u).(*Player); ok && player.class == CLASS_WARRIOR {
		rageGain := uint32(0)
		if isAttacker {
			// 攻击者获得更多怒气
			rageGain = damage / 100
		} else {
			// 受害者获得较少怒气
			rageGain = damage / 200
		}

		if rageGain > 0 {
			u.ModifyPower(uint8(POWER_RAGE), int32(rageGain))
			if rageGain > 1 {
				fmt.Printf("%s 获得 %d 点怒气\n", u.name, rageGain)
			}
		}
	}
}

// === 网络支持方法 ===

// SetGUID 设置GUID
func (u *Unit) SetGUID(guid uint64) {
	u.guid = guid
}

// IsValidAttackTarget 检查是否为有效攻击目标
func (u *Unit) IsValidAttackTarget(target IUnit) bool {
	if target == nil {
		return false
	}

	// 不能攻击自己
	if target.GetGUID() == u.GetGUID() {
		return false
	}

	// 目标必须存活
	if !target.IsAlive() {
		return false
	}

	// 简化版：玩家可以攻击生物，生物可以攻击玩家
	return true
}

// SetTarget 设置目标
func (u *Unit) SetTarget(target IUnit) {
	u.target = target
}

// GetTarget 获取目标
func (u *Unit) GetTarget() IUnit {
	return u.target
}

// CastSpell 施放法术 - 基于AzerothCore的Unit::CastSpell
func (u *Unit) CastSpell(target IUnit, spellId uint32) {
	// 获取法术信息
	spellInfo := GlobalSpellManager.GetSpell(spellId)
	if spellInfo == nil {
		fmt.Printf("未知法术ID: %d\n", spellId)
		return
	}

	// 检查是否在冷却中
	if u.isSpellOnCooldown(spellId) {
		fmt.Printf("%s 的 %s 还在冷却中\n", u.GetName(), spellInfo.Name)
		return
	}

	// 检查是否已在施法
	if u.isCurrentlySpellCasting() {
		fmt.Printf("%s 正在施法，无法施放新法术\n", u.GetName())
		return
	}

	// 创建法术实例
	spell := NewSpell(u, spellInfo, u.world)

	// 准备施法
	if spell.Prepare(target) {
		// 根据法术类型添加到对应槽位
		spellType := CURRENT_GENERIC_SPELL
		if spellInfo.IsChanneled {
			spellType = CURRENT_CHANNELED_SPELL
		}

		u.currentSpells[spellType] = spell

		// 设置冷却时间
		if spellInfo.Cooldown > 0 {
			u.spellCooldowns[spellId] = time.Now().Add(spellInfo.Cooldown)
		}
	}
}

// updateSpells 更新法术状态 - 基于AzerothCore的Unit::_UpdateSpells
func (u *Unit) updateSpells(diff uint32) {
	// 更新所有当前施法中的法术
	for spellType, spell := range u.currentSpells {
		if spell != nil {
			// 更新法术
			if !spell.Update(diff) {
				// 法术完成或被打断，移除
				delete(u.currentSpells, spellType)
			}
		}
	}

	// 清理过期的冷却时间
	now := time.Now()
	for spellId, cooldownEnd := range u.spellCooldowns {
		if now.After(cooldownEnd) {
			delete(u.spellCooldowns, spellId)
		}
	}
}

// isSpellOnCooldown 检查法术是否在冷却中
func (u *Unit) isSpellOnCooldown(spellId uint32) bool {
	if cooldownEnd, exists := u.spellCooldowns[spellId]; exists {
		return time.Now().Before(cooldownEnd)
	}
	return false
}

// isCurrentlySpellCasting 检查是否正在施法
func (u *Unit) isCurrentlySpellCasting() bool {
	return len(u.currentSpells) > 0
}

// InterruptSpell 打断法术 - 基于AzerothCore的Unit::InterruptSpell
func (u *Unit) InterruptSpell(spellType int) {
	if spell, exists := u.currentSpells[spellType]; exists && spell != nil {
		spell.Interrupt()
		delete(u.currentSpells, spellType)
	}
}

// InterruptNonMeleeSpells 打断非近战法术 - 基于AzerothCore的逻辑
func (u *Unit) InterruptNonMeleeSpells(withDelayed bool) {
	// 打断通用法术
	u.InterruptSpell(CURRENT_GENERIC_SPELL)

	// 打断引导法术
	u.InterruptSpell(CURRENT_CHANNELED_SPELL)

	// 打断自动重复法术
	u.InterruptSpell(CURRENT_AUTOREPEAT_SPELL)
}

// SetWorld 设置世界引用
func (u *Unit) SetWorld(world *World) {
	u.world = world
}

// GetCurrentSpell 获取当前施法中的法术
func (u *Unit) GetCurrentSpell(spellType int) *Spell {
	return u.currentSpells[spellType]
}

// HasSpellCooldown 检查法术冷却
func (u *Unit) HasSpellCooldown(spellId uint32) bool {
	return u.isSpellOnCooldown(spellId)
}

// GetSpellCooldownDelay 获取法术冷却剩余时间
func (u *Unit) GetSpellCooldownDelay(spellId uint32) time.Duration {
	if cooldownEnd, exists := u.spellCooldowns[spellId]; exists {
		remaining := time.Until(cooldownEnd)
		if remaining > 0 {
			return remaining
		}
	}
	return 0
}

// Heal 治疗
func (u *Unit) Heal(caster IUnit, amount uint32) {
	if !u.IsAlive() {
		return
	}

	oldHealth := u.GetHealth()
	newHealth := min(oldHealth+amount, u.GetMaxHealth())
	u.SetHealth(newHealth)

	actualHealing := newHealth - oldHealth
	if actualHealing > 0 {
		fmt.Printf("%s 被治疗了 %d 点生命值 (%d/%d)\n",
			u.GetName(), actualHealing, newHealth, u.GetMaxHealth())

		// 网络广播治疗信息 - 基于AzerothCore的网络同步
		if u.world != nil && caster != nil {
			u.world.BroadcastUnitUpdate(u) // 广播生命值更新
		}
	}
}

// buildHealthUpdateBlock 构建血量更新数据块 - 基于AzerothCore的UpdateData机制
func (u *Unit) buildHealthUpdateBlock(oldHealth, newHealth uint32) []byte {
	// 创建一个临时的WorldPacket来构建数据
	packet := NewWorldPacket(SMSG_HEALTH_UPDATE)
	packet.WriteUint64(u.guid)      // 单位GUID
	packet.WriteUint32(oldHealth)   // 旧血量
	packet.WriteUint32(newHealth)   // 新血量
	packet.WriteUint32(u.maxHealth) // 最大血量

	// 返回数据块
	return packet.data
}

// buildPowerUpdateBlock 构建能量更新数据块 - 基于AzerothCore的UpdateData机制
func (u *Unit) buildPowerUpdateBlock(powerType uint8, oldPower, newPower uint32) []byte {
	// 创建一个临时的WorldPacket来构建数据
	packet := NewWorldPacket(SMSG_POWER_UPDATE)
	packet.WriteUint64(u.guid)                   // 单位GUID
	packet.WriteUint8(powerType)                 // 能量类型
	packet.WriteUint32(oldPower)                 // 旧能量值
	packet.WriteUint32(newPower)                 // 新能量值
	packet.WriteUint32(u.GetMaxPower(powerType)) // 最大能量值

	// 返回数据块
	return packet.data
}

// buildPositionUpdateBlock 构建位置更新数据块 - 基于AzerothCore的移动同步
func (u *Unit) buildPositionUpdateBlock() []byte {
	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint64(u.guid)         // 单位GUID
	packet.WriteFloat32(u.x)           // X坐标
	packet.WriteFloat32(u.y)           // Y坐标
	packet.WriteFloat32(u.z)           // Z坐标
	packet.WriteFloat32(u.orientation) // 朝向

	return packet.data
}

// buildFullUpdateBlock 构建完整状态更新数据块 - 基于AzerothCore的完整对象更新
func (u *Unit) buildFullUpdateBlock() []byte {
	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint64(u.guid)      // 单位GUID
	packet.WriteUint32(u.health)    // 当前血量
	packet.WriteUint32(u.maxHealth) // 最大血量

	// 写入所有能量类型
	for powerType := uint8(0); powerType < 4; powerType++ {
		packet.WriteUint32(u.GetPower(powerType))    // 当前能量
		packet.WriteUint32(u.GetMaxPower(powerType)) // 最大能量
	}

	// 写入位置信息
	packet.WriteFloat32(u.x)
	packet.WriteFloat32(u.y)
	packet.WriteFloat32(u.z)
	packet.WriteFloat32(u.orientation)

	// 写入状态标志
	packet.WriteUint32(u.getUnitFlags())

	return packet.data
}

// getUnitFlags 获取单位状态标志 - 基于AzerothCore的UnitFlags
func (u *Unit) getUnitFlags() uint32 {
	flags := uint32(0)

	if !u.IsAlive() {
		flags |= 0x00000001 // UNIT_FLAG_DEAD
	}

	if u.IsInCombat() {
		flags |= 0x00080000 // UNIT_FLAG_IN_COMBAT
	}

	// 可以根据需要添加更多标志
	return flags
}

// AddBatchUpdateForMovement 为移动添加批量更新 - 基于AzerothCore的移动同步
func (u *Unit) AddBatchUpdateForMovement() {
	if u.world == nil {
		return
	}

	// 构建位置更新数据块
	updateBlock := u.buildPositionUpdateBlock()

	// 获取范围内的玩家会话
	players := u.world.GetPlayersInRange(u.x, u.y, u.z, 100.0)
	for _, player := range players {
		u.world.AddBatchUpdate(u, player.id, updateBlock)
	}
}

// AddBatchUpdateForFullState 为完整状态添加批量更新 - 基于AzerothCore的完整对象同步
func (u *Unit) AddBatchUpdateForFullState() {
	if u.world == nil {
		return
	}

	// 构建完整状态更新数据块
	updateBlock := u.buildFullUpdateBlock()

	// 获取范围内的玩家会话
	players := u.world.GetPlayersInRange(u.x, u.y, u.z, 100.0)
	for _, player := range players {
		u.world.AddBatchUpdate(u, player.id, updateBlock)
	}

	fmt.Printf("[BatchUpdate] 完整状态更新: %s (范围: %d玩家)\n", u.name, len(players))
}
