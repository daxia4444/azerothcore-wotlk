package main

import (
	"fmt"
	"math/rand"
)

// 伤害处理实现 - 对应AzerothCore的Unit::DealDamage函数
func (u *Unit) DealDamage(attacker IUnit, damage uint32, damageType int, schoolMask int) uint32 {
	// 脚本钩子 - 允许修改伤害
	damage = u.scriptHookDamage(attacker, damage, damageType)

	// 保存用于怒气计算的伤害值
	rageDamage := damage

	// 通知AI系统
	if u.ai != nil {
		u.ai.DamageTaken(attacker, damage)
	}
	if attacker != nil && attacker.GetAI() != nil {
		attacker.GetAI().DamageDealt(u, damage)
	}

	// GM无敌检查（简化版）
	if u.isGMGodMode() {
		return 0
	}

	// 如果伤害为0，仍然处理怒气奖励
	if damage == 0 {
		if unitSelf, ok := IUnit(u).(*Unit); ok {
			unitSelf.rewardRageFromAbsorbedDamage(rageDamage)
		}
		return 0
	}

	// 决斗特殊处理（简化版）
	if u.isDueling() && damage >= u.health {
		damage = u.health - 1 // 决斗中不会真正死亡
		fmt.Printf("决斗中 %s 的生命值被限制为1点\n", u.name)
	}

	// 切磋系统处理
	if u.isSparring(attacker) && damage >= u.health {
		damage = 0
		fmt.Printf("切磋中 %s 避免了致命伤害\n", u.name)
		return 0
	}

	// 处理死亡
	if u.health <= damage {
		fmt.Printf("致命伤害: %s 即将死亡\n", u.name)
		u.handleDeath(attacker, damageType, schoolMask)
		return damage
	}

	// 存活时的伤害处理
	u.ModifyHealth(-int32(damage))

	// 移除因直接伤害中断的光环（简化版）
	if damageType == DIRECT_DAMAGE || damageType == SPELL_DIRECT_DAMAGE {
		u.removeDirectDamageAuras()
	}

	// 网络广播伤害信息 - 基于AzerothCore的网络同步
	if u.world != nil && attacker != nil {
		hitResult := MELEE_HIT_NORMAL // 简化处理，实际应该根据攻击类型确定
		u.world.BroadcastAttackerStateUpdate(attacker, u, damage, hitResult, schoolMask)
		// 注意：血量更新已经在ModifyHealth中自动广播了
	}

	// 更新威胁值
	if attacker != nil {
		u.threatManager.AddThreat(attacker, float32(damage))

		// 开始战斗状态
		if !u.IsInCombat() {
			u.CombatStart(attacker)
		}
	}

	// 装备耐久度损失（简化版）
	u.handleDurabilityLoss()

	// 受伤怒气奖励
	if unitSelf, ok := IUnit(u).(*Unit); ok {
		unitSelf.rewardRage(damage, false)
	}

	// 法术推迟处理（简化版）
	if damageType != NODAMAGE && damage > 0 {
		u.handleSpellPushback(damage)
	}

	return damage
}

// 脚本钩子 - 允许脚本修改伤害
func (u *Unit) scriptHookDamage(attacker IUnit, damage uint32, damageType int) uint32 {
	// 这里可以添加各种脚本逻辑
	// 例如：训练假人设置伤害为0
	if u.name == "TrainingDummy" {
		return 0
	}

	// 其他脚本修改...
	return damage
}

// GM无敌模式检查
func (u *Unit) isGMGodMode() bool {
	// 简化版：假设没有GM无敌
	return false
}

// 决斗检查
func (u *Unit) isDueling() bool {
	// 简化版：假设没有决斗
	return false
}

// 切磋检查
func (u *Unit) isSparring(attacker IUnit) bool {
	// 简化版：假设没有切磋
	return false
}

// 处理死亡
func (u *Unit) handleDeath(killer IUnit, damageType int, schoolMask int) {
	fmt.Printf("%s 死于 %s 的攻击\n", u.name, killer.GetName())

	// 设置生命值为0
	u.health = 0

	// 调用死亡处理
	u.setDeathState()

	// 停止所有攻击
	u.AttackStop()

	// 清除所有攻击者的目标
	for _, attacker := range u.attackers {
		if attacker.GetVictim() == u {
			attacker.SetVictim(nil)
		}
	}

	// 清空攻击者列表
	u.attackers = make(map[uint64]IUnit)
}

// 移除直接伤害光环
func (u *Unit) removeDirectDamageAuras() {
	// 简化版：假设移除了一些光环
	// 在真实实现中，这里会移除具有AURA_INTERRUPT_FLAG_TAKE_DAMAGE标志的光环
}

// 处理装备耐久度损失
func (u *Unit) handleDurabilityLoss() {
	// 简化版：随机耐久度损失
	if rand.Float32() < 0.1 { // 10%概率
		fmt.Printf("%s 的装备耐久度下降\n", u.name)
	}
}

// 处理法术推迟
func (u *Unit) handleSpellPushback(damage uint32) {
	// 简化版：如果正在施法，可能被推迟
	if u.HasUnitState(UNIT_STATE_CASTING) {
		if damage > 100 {
			fmt.Printf("%s 的法术施放被推迟\n", u.name)
		}
	}
}

// 从被吸收的伤害中奖励怒气
func (u *Unit) rewardRageFromAbsorbedDamage(absorbedDamage uint32) {
	if player, ok := IUnit(u).(*Player); ok && player.class == CLASS_WARRIOR {
		rageGain := absorbedDamage / 200 // 被吸收伤害的怒气奖励较少
		if rageGain > 0 {
			u.ModifyPower(POWER_RAGE, int32(rageGain))
		}
	}
}

// 威胁管理器
type ThreatManager struct {
	threatList map[uint64]*ThreatInfo
}

type ThreatInfo struct {
	unit   IUnit
	threat float32
}

func NewThreatManager() *ThreatManager {
	return &ThreatManager{
		threatList: make(map[uint64]*ThreatInfo),
	}
}

func (tm *ThreatManager) AddThreat(unit IUnit, threat float32) {
	guid := unit.GetGUID()
	if info, exists := tm.threatList[guid]; exists {
		info.threat += threat
	} else {
		tm.threatList[guid] = &ThreatInfo{
			unit:   unit,
			threat: threat,
		}
	}

	if threat > 50 {
		fmt.Printf("威胁值更新: %s 对目标的威胁值增加 %.1f\n", unit.GetName(), threat)
	}
}

func (tm *ThreatManager) GetHighestThreatTarget() IUnit {
	var highestThreat float32 = 0
	var target IUnit = nil

	for _, info := range tm.threatList {
		if info.unit.IsAlive() && info.threat > highestThreat {
			highestThreat = info.threat
			target = info.unit
		}
	}

	return target
}

func (tm *ThreatManager) RemoveThreat(unit IUnit) {
	delete(tm.threatList, unit.GetGUID())
}

func (tm *ThreatManager) ClearAllThreat() {
	tm.threatList = make(map[uint64]*ThreatInfo)
}

func (tm *ThreatManager) IsEmpty() bool {
	for _, info := range tm.threatList {
		if info.unit.IsAlive() {
			return false
		}
	}
	return true
}

// 伤害计算辅助函数
func calculateArmorReduction(armor uint32, attackerLevel uint8) float32 {
	// 简化的护甲减伤计算
	// 真实公式更复杂，涉及攻击者等级和目标护甲值
	if armor == 0 {
		return 1.0
	}

	// 基础公式：减伤 = armor / (armor + 400 + 85 * attackerLevel)
	reduction := float32(armor) / (float32(armor) + 400 + 85*float32(attackerLevel))

	// 限制最大减伤为75%
	if reduction > 0.75 {
		reduction = 0.75
	}

	return 1.0 - reduction
}

// 抗性计算
func calculateResistance(resistance uint32, spellLevel uint8) float32 {
	// 简化的抗性计算
	if resistance == 0 {
		return 1.0
	}

	// 基础抗性减免
	resistChance := float32(resistance) / (float32(resistance) + float32(spellLevel)*5)

	// 限制最大抗性为75%
	if resistChance > 0.75 {
		resistChance = 0.75
	}

	return 1.0 - resistChance
}

// 暴击伤害加成
func calculateCriticalDamage(baseDamage uint32, critMultiplier float32) uint32 {
	return uint32(float32(baseDamage) * critMultiplier)
}

// 伤害吸收处理
func (u *Unit) absorbDamage(damage uint32, schoolMask int) (uint32, uint32) {
	// 简化版：假设有一些伤害吸收
	absorbed := uint32(0)

	// 模拟护盾吸收
	if rand.Float32() < 0.2 { // 20%概率有护盾
		absorbed = damage / 4 // 吸收25%伤害
		if absorbed > 0 {
			fmt.Printf("%s 的护盾吸收了 %d 点伤害\n", u.name, absorbed)
		}
	}

	finalDamage := damage - absorbed
	return finalDamage, absorbed
}

// 获取单位所在的世界引用 - 辅助函数
func getWorldFromUnit(unit IUnit) *World {
	if u, ok := unit.(*Unit); ok {
		return u.world
	}
	if p, ok := unit.(*Player); ok {
		return p.world
	}
	return nil
}

// 获取法术学派名称
func getSchoolName(schoolMask int) string {
	switch schoolMask {
	case SPELL_SCHOOL_NORMAL:
		return "物理"
	case SPELL_SCHOOL_HOLY:
		return "神圣"
	case SPELL_SCHOOL_FIRE:
		return "火焰"
	case SPELL_SCHOOL_NATURE:
		return "自然"
	case SPELL_SCHOOL_FROST:
		return "冰霜"
	case SPELL_SCHOOL_SHADOW:
		return "暗影"
	case SPELL_SCHOOL_ARCANE:
		return "奥术"
	default:
		return "未知"
	}
}
