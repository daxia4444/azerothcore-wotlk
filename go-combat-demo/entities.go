package main

import (
	"fmt"
	"math/rand"
)

// 玩家结构
type Player struct {
	*Unit
	class uint8
}

// 创建玩家
func NewPlayer(name string, level uint8, class uint8) *Player {
	player := &Player{
		Unit:  NewUnit(generateGUID(), name, level, UNIT_TYPE_PLAYER),
		class: class,
	}

	// 设置基础AI
	player.SetAI(NewPlayerAI(player))

	// 根据职业设置能量类型
	switch class {
	case CLASS_WARRIOR:
		player.SetMaxPower(POWER_RAGE, 100)
		player.SetPower(POWER_RAGE, 0)
	case CLASS_MAGE, CLASS_PRIEST, CLASS_WARLOCK:
		player.SetMaxPower(POWER_MANA, 5000)
		player.SetPower(POWER_MANA, 5000)
	case CLASS_ROGUE:
		player.SetMaxPower(POWER_ENERGY, 100)
		player.SetPower(POWER_ENERGY, 100)
	case CLASS_HUNTER:
		player.SetMaxPower(POWER_MANA, 3000)
		player.SetPower(POWER_MANA, 3000)
	}

	return player
}

func (p *Player) GetClass() uint8 {
	return p.class
}

// 生物结构
type Creature struct {
	*Unit
	creatureType uint8
	ai           IAI
}

// 创建生物
func NewCreature(name string, level uint8, creatureType uint8) *Creature {
	creature := &Creature{
		Unit:         NewUnit(generateGUID(), name, level, UNIT_TYPE_CREATURE),
		creatureType: creatureType,
	}

	// 设置基础AI
	creature.SetAI(NewCreatureAI(creature))

	return creature
}

func (c *Creature) GetCreatureType() uint8 {
	return c.creatureType
}

// === Player网络支持方法 ===

// SetGUID 设置GUID
func (p *Player) SetGUID(guid uint64) {
	p.Unit.SetGUID(guid)
}

// IsValidAttackTarget 检查是否为有效攻击目标
func (p *Player) IsValidAttackTarget(target IUnit) bool {
	return p.Unit.IsValidAttackTarget(target)
}

// SetTarget 设置目标
func (p *Player) SetTarget(target IUnit) {
	p.Unit.SetTarget(target)
}

// GetTarget 获取目标
func (p *Player) GetTarget() IUnit {
	return p.Unit.GetTarget()
}

// CastSpell 施放法术
func (p *Player) CastSpell(target IUnit, spellId uint32) {
	p.Unit.CastSpell(target, spellId)
}

// Heal 治疗
func (p *Player) Heal(caster IUnit, amount uint32) {
	p.Unit.Heal(caster, amount)
}

// === Creature网络支持方法 ===

// SetGUID 设置GUID
func (c *Creature) SetGUID(guid uint64) {
	c.Unit.SetGUID(guid)
}

// IsValidAttackTarget 检查是否为有效攻击目标
func (c *Creature) IsValidAttackTarget(target IUnit) bool {
	return c.Unit.IsValidAttackTarget(target)
}

// SetTarget 设置目标
func (c *Creature) SetTarget(target IUnit) {
	c.Unit.SetTarget(target)
}

// GetTarget 获取目标
func (c *Creature) GetTarget() IUnit {
	return c.Unit.GetTarget()
}

// CastSpell 施放法术
func (c *Creature) CastSpell(target IUnit, spellId uint32) {
	c.Unit.CastSpell(target, spellId)
}

// Heal 治疗
func (c *Creature) Heal(caster IUnit, amount uint32) {
	c.Unit.Heal(caster, amount)
}

// 玩家AI
type PlayerAI struct {
	owner           *Player
	lastActionTime  uint32
	lastHealTime    uint32
	lastSpecialTime uint32
}

func NewPlayerAI(owner *Player) *PlayerAI {
	return &PlayerAI{owner: owner}
}

func (ai *PlayerAI) UpdateAI(diff uint32) {
	if !ai.owner.IsAlive() {
		return
	}

	ai.lastActionTime += diff
	ai.lastHealTime += diff
	ai.lastSpecialTime += diff

	// 根据职业执行不同的AI逻辑
	switch ai.owner.GetClass() {
	case CLASS_WARRIOR:
		ai.updateWarriorAI()
	case CLASS_MAGE:
		ai.updateMageAI()
	case CLASS_PRIEST:
		ai.updatePriestAI()
	case CLASS_HUNTER:
		ai.updateHunterAI()
	case CLASS_WARLOCK:
		ai.updateWarlockAI()
	}

	// 通用目标选择逻辑
	if ai.owner.GetVictim() == nil && ai.owner.IsInCombat() {
		ai.selectTarget()
	}
}

// 战士AI - 坦克职责
func (ai *PlayerAI) updateWarriorAI() {
	// 坦克优先攻击威胁最高的目标
	if ai.lastActionTime >= 1500 { // 1.5秒攻击间隔
		ai.lastActionTime = 0
		if ai.owner.GetVictim() != nil && !ai.owner.isCurrentlySpellCasting() {
			// 使用英勇打击
			ai.owner.CastSpell(ai.owner.GetVictim(), SPELL_HEROIC_STRIKE)
		}
	}

	// 嘲讽技能 - 每8秒
	if ai.lastSpecialTime >= 8000 {
		ai.lastSpecialTime = 0
		if ai.owner.IsInCombat() && ai.owner.GetVictim() != nil {
			ai.owner.CastSpell(ai.owner.GetVictim(), SPELL_TAUNT)
		}
	}
}

// 法师AI - DPS职责
func (ai *PlayerAI) updateMageAI() {
	// 法师优先攻击生命值最低的目标
	if ai.lastActionTime >= 2500 { // 2.5秒施法间隔
		ai.lastActionTime = 0
		if ai.owner.GetVictim() != nil && !ai.owner.isCurrentlySpellCasting() {
			// 随机选择法术
			spells := []uint32{SPELL_FROSTBOLT, SPELL_FIREBALL}
			selectedSpell := spells[rand.Intn(len(spells))]
			ai.owner.CastSpell(ai.owner.GetVictim(), selectedSpell)
		}
	}

	// 群体法术 - 当敌人数量>=3时
	if ai.lastSpecialTime >= 10000 {
		ai.lastSpecialTime = 0
		enemyCount := ai.countNearbyEnemies()
		if enemyCount >= 3 && !ai.owner.isCurrentlySpellCasting() {
			// 施放暴风雪
			ai.owner.CastSpell(ai.owner.GetVictim(), SPELL_BLIZZARD)
		} else if ai.owner.GetVictim() != nil && !ai.owner.isCurrentlySpellCasting() {
			// 施放冰霜新星
			ai.owner.CastSpell(ai.owner, SPELL_FROST_NOVA)
		}
	}
}

// 牧师AI - 治疗职责
func (ai *PlayerAI) updatePriestAI() {
	// 优先治疗生命值最低的队友
	if ai.lastHealTime >= 2000 { // 2秒治疗间隔
		ai.lastHealTime = 0
		target := ai.findHealTarget()
		if target != nil && !ai.owner.isCurrentlySpellCasting() {
			// 根据伤势选择治疗法术
			healthPct := float32(target.GetHealth()) / float32(target.GetMaxHealth())
			if healthPct < 0.3 {
				// 生命值低于30%，使用快速治疗
				ai.owner.CastSpell(target, SPELL_FLASH_HEAL)
			} else {
				// 使用普通治疗术
				ai.owner.CastSpell(target, SPELL_HEAL)
			}
		}
	}

	// 预防性护盾
	if ai.lastSpecialTime >= 8000 {
		ai.lastSpecialTime = 0
		target := ai.findShieldTarget()
		if target != nil && !ai.owner.isCurrentlySpellCasting() {
			ai.owner.CastSpell(target, SPELL_POWER_WORD_SHIELD)
		}
	}

	// 攻击性法术
	if ai.lastActionTime >= 3000 && ai.owner.GetVictim() != nil && !ai.owner.isCurrentlySpellCasting() {
		ai.lastActionTime = 0
		ai.owner.CastSpell(ai.owner.GetVictim(), SPELL_SMITE)
	}
}

// 猎人AI - 远程DPS
func (ai *PlayerAI) updateHunterAI() {
	// 猎人保持距离攻击
	if ai.lastActionTime >= 1800 { // 1.8秒射击间隔
		ai.lastActionTime = 0
		if ai.owner.GetVictim() != nil && !ai.owner.isCurrentlySpellCasting() {
			// 使用瞄准射击
			ai.owner.CastSpell(ai.owner.GetVictim(), SPELL_AIMED_SHOT)
		}
	}

	// 多重射击 - 对付多个敌人
	if ai.lastSpecialTime >= 12000 {
		ai.lastSpecialTime = 0
		enemyCount := ai.countNearbyEnemies()
		if enemyCount >= 2 && !ai.owner.isCurrentlySpellCasting() {
			ai.owner.CastSpell(ai.owner.GetVictim(), SPELL_MULTI_SHOT)
		} else if ai.owner.GetVictim() != nil && !ai.owner.isCurrentlySpellCasting() {
			// 标记目标
			ai.owner.CastSpell(ai.owner.GetVictim(), SPELL_HUNTER_MARK)
		}
	}
}

// 术士AI - DOT和召唤
func (ai *PlayerAI) updateWarlockAI() {
	// 术士施放持续伤害法术
	if ai.lastActionTime >= 2000 {
		ai.lastActionTime = 0
		if ai.owner.GetVictim() != nil && !ai.owner.isCurrentlySpellCasting() {
			// 优先施放DOT法术
			spells := []uint32{SPELL_SHADOW_BOLT, SPELL_IMMOLATE, SPELL_CORRUPTION}
			selectedSpell := spells[rand.Intn(len(spells))]
			ai.owner.CastSpell(ai.owner.GetVictim(), selectedSpell)
		}
	}

	// 恐惧术 - 紧急情况
	if ai.lastSpecialTime >= 15000 && ai.owner.GetHealth() < ai.owner.GetMaxHealth()/3 {
		ai.lastSpecialTime = 0
		if ai.owner.GetVictim() != nil && !ai.owner.isCurrentlySpellCasting() {
			ai.owner.CastSpell(ai.owner.GetVictim(), SPELL_FEAR)
		}
	}
}

// 选择攻击目标
func (ai *PlayerAI) selectTarget() {
	var bestTarget IUnit

	// 战士优先攻击威胁最高的
	if ai.owner.GetClass() == CLASS_WARRIOR {
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() {
				bestTarget = attacker
				break
			}
		}
	} else {
		// 其他职业攻击生命值最低的
		var lowestHealth uint32 = ^uint32(0) // 最大值
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() && attacker.GetHealth() < lowestHealth {
				lowestHealth = attacker.GetHealth()
				bestTarget = attacker
			}
		}
	}

	if bestTarget != nil {
		ai.owner.Attack(bestTarget)
	}
}

// 寻找治疗目标
func (ai *PlayerAI) findHealTarget() IUnit {
	// 这里需要访问团队成员，简化处理
	if ai.owner.GetHealth() < ai.owner.GetMaxHealth()*2/3 {
		return ai.owner
	}
	return nil
}

// 寻找护盾目标
func (ai *PlayerAI) findShieldTarget() IUnit {
	// 简化处理：为自己或生命值较低的目标施加护盾
	if ai.owner.GetHealth() < ai.owner.GetMaxHealth()*4/5 {
		return ai.owner
	}
	return nil
}

// 计算附近敌人数量
func (ai *PlayerAI) countNearbyEnemies() int {
	count := 0
	for _, attacker := range ai.owner.attackers {
		if attacker.IsAlive() {
			count++
		}
	}
	return count
}

func (ai *PlayerAI) AttackStart(target IUnit) {
	fmt.Printf("[PlayerAI] %s 开始攻击 %s\n", ai.owner.GetName(), target.GetName())
}

func (ai *PlayerAI) EnterCombat(target IUnit) {
	fmt.Printf("[PlayerAI] %s 与 %s 进入战斗\n", ai.owner.GetName(), target.GetName())
}

func (ai *PlayerAI) JustDied(killer IUnit) {
	if killer != nil {
		fmt.Printf("[PlayerAI] %s 被 %s 杀死\n", ai.owner.GetName(), killer.GetName())
	} else {
		fmt.Printf("[PlayerAI] %s 死亡\n", ai.owner.GetName())
	}
}

func (ai *PlayerAI) DamageTaken(attacker IUnit, damage uint32) {
	if damage > 500 {
		fmt.Printf("[PlayerAI] %s 受到来自 %s 的大量伤害: %d\n",
			ai.owner.GetName(), attacker.GetName(), damage)
	}
}

func (ai *PlayerAI) DamageDealt(victim IUnit, damage uint32) {
	if damage > 500 {
		fmt.Printf("[PlayerAI] %s 对 %s 造成大量伤害: %d\n",
			ai.owner.GetName(), victim.GetName(), damage)
	}
}

// 生物AI
type CreatureAI struct {
	owner *Creature
}

func NewCreatureAI(owner *Creature) *CreatureAI {
	return &CreatureAI{owner: owner}
}

func (ai *CreatureAI) UpdateAI(diff uint32) {
	if !ai.owner.IsAlive() {
		return
	}

	// 如果没有目标，寻找攻击者
	if ai.owner.GetVictim() == nil {
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() {
				ai.owner.Attack(attacker)
				fmt.Printf("[CreatureAI] %s 开始反击 %s\n", ai.owner.GetName(), attacker.GetName())
				break
			}
		}
	}

	// 如果有目标但目标死亡，清除目标
	if ai.owner.GetVictim() != nil && !ai.owner.GetVictim().IsAlive() {
		ai.owner.SetVictim(nil)
	}
}

func (ai *CreatureAI) AttackStart(target IUnit) {
	fmt.Printf("[CreatureAI] %s 开始攻击 %s\n", ai.owner.GetName(), target.GetName())
}

func (ai *CreatureAI) EnterCombat(target IUnit) {
	fmt.Printf("[CreatureAI] %s 与 %s 进入战斗\n", ai.owner.GetName(), target.GetName())
	// 立即开始攻击
	if ai.owner.GetVictim() == nil {
		ai.owner.Attack(target)
		fmt.Printf("[CreatureAI] %s 立即开始反击 %s\n", ai.owner.GetName(), target.GetName())
	}
}

func (ai *CreatureAI) JustDied(killer IUnit) {
	if killer != nil {
		fmt.Printf("[CreatureAI] %s 被 %s 杀死\n", ai.owner.GetName(), killer.GetName())
	} else {
		fmt.Printf("[CreatureAI] %s 死亡\n", ai.owner.GetName())
	}
}

func (ai *CreatureAI) DamageTaken(attacker IUnit, damage uint32) {
	// 生物受到伤害时的反应
	if damage > ai.owner.GetMaxHealth()/4 { // 超过25%最大生命值
		fmt.Printf("[CreatureAI] %s 受到重创，伤害: %d\n", ai.owner.GetName(), damage)
	}
}

func (ai *CreatureAI) DamageDealt(victim IUnit, damage uint32) {
	// 生物造成伤害时的反应
	if damage > 800 {
		fmt.Printf("[CreatureAI] %s 造成致命一击: %d\n", ai.owner.GetName(), damage)
	}
}
