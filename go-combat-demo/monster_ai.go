package main

import (
	"fmt"
	"math/rand"
)

// 矿工AI - 基础近战攻击
type MinerAI struct {
	owner          *Creature
	lastAttackTime uint32
}

func NewMinerAI(owner *Creature) *MinerAI {
	return &MinerAI{owner: owner}
}

func (ai *MinerAI) UpdateAI(diff uint32) {
	if !ai.owner.IsAlive() {
		return
	}

	ai.lastAttackTime += diff

	// 如果没有目标，寻找攻击者
	if ai.owner.GetVictim() == nil {
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() {
				ai.owner.Attack(attacker)
				fmt.Printf("[矿工AI] %s 开始攻击 %s\n", ai.owner.GetName(), attacker.GetName())
				break
			}
		}
	}

	// 简单的攻击逻辑
	if ai.owner.GetVictim() != nil && ai.lastAttackTime >= 2000 { // 2秒攻击间隔
		ai.lastAttackTime = 0
		// 矿工的普通攻击
	}
}

func (ai *MinerAI) AttackStart(target IUnit) {
	fmt.Printf("[矿工AI] %s 愤怒地挥舞着镐子攻击 %s\n", ai.owner.GetName(), target.GetName())
}

func (ai *MinerAI) EnterCombat(target IUnit) {
	fmt.Printf("[矿工AI] %s 大喊：\"别想破坏我们的工作！\"\n", ai.owner.GetName())
}

func (ai *MinerAI) JustDied(killer IUnit) {
	if killer != nil {
		fmt.Printf("[矿工AI] %s 临死前喊道：\"迪菲亚兄弟会...永不屈服...\"\n", ai.owner.GetName())
	}
}

func (ai *MinerAI) DamageTaken(attacker IUnit, damage uint32) {
	if damage > 300 {
		fmt.Printf("[矿工AI] %s 痛苦地呻吟：\"啊！我的背！\"\n", ai.owner.GetName())
	}
}

func (ai *MinerAI) DamageDealt(victim IUnit, damage uint32) {
	if damage > 200 {
		fmt.Printf("[矿工AI] %s 得意地说：\"尝尝我镐子的厉害！\"\n", ai.owner.GetName())
	}
}

// 监工AI - 有指挥能力
type OverseerAI struct {
	owner           *Creature
	lastCommandTime uint32
	lastAttackTime  uint32
}

func NewOverseerAI(owner *Creature) *OverseerAI {
	return &OverseerAI{owner: owner}
}

func (ai *OverseerAI) UpdateAI(diff uint32) {
	if !ai.owner.IsAlive() {
		return
	}

	ai.lastCommandTime += diff
	ai.lastAttackTime += diff

	// 寻找目标
	if ai.owner.GetVictim() == nil {
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() {
				ai.owner.Attack(attacker)
				break
			}
		}
	}

	// 指挥技能 - 每10秒鼓舞周围的小弟
	if ai.lastCommandTime >= 10000 {
		ai.lastCommandTime = 0
		fmt.Printf("[监工AI] %s 大喊：\"给我狠狠地打！\"\n", ai.owner.GetName())
		// 这里可以添加增益效果给周围的小怪
	}
}

func (ai *OverseerAI) AttackStart(target IUnit) {
	fmt.Printf("[监工AI] %s 指着 %s 大喊：\"就是这个入侵者！\"\n", ai.owner.GetName(), target.GetName())
}

func (ai *OverseerAI) EnterCombat(target IUnit) {
	fmt.Printf("[监工AI] %s 怒吼：\"保卫我们的矿井！\"\n", ai.owner.GetName())
}

func (ai *OverseerAI) JustDied(killer IUnit) {
	fmt.Printf("[监工AI] %s 临死前咒骂：\"范克里夫大人...会为我报仇的...\"\n", ai.owner.GetName())
}

func (ai *OverseerAI) DamageTaken(attacker IUnit, damage uint32) {
	if damage > ai.owner.GetMaxHealth()/3 {
		fmt.Printf("[监工AI] %s 愤怒地咆哮：\"你们会为此付出代价！\"\n", ai.owner.GetName())
	}
}

func (ai *OverseerAI) DamageDealt(victim IUnit, damage uint32) {
	if damage > 400 {
		fmt.Printf("[监工AI] %s 冷笑：\"这就是反抗的下场！\"\n", ai.owner.GetName())
	}
}

// 暴徒AI - 敏捷型攻击
type ThugAI struct {
	owner           *Creature
	lastStealthTime uint32
	lastAttackTime  uint32
	isStealthed     bool
}

func NewThugAI(owner *Creature) *ThugAI {
	return &ThugAI{owner: owner}
}

func (ai *ThugAI) UpdateAI(diff uint32) {
	if !ai.owner.IsAlive() {
		return
	}

	ai.lastStealthTime += diff
	ai.lastAttackTime += diff

	// 寻找目标
	if ai.owner.GetVictim() == nil {
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() {
				ai.owner.Attack(attacker)
				break
			}
		}
	}

	// 潜行技能 - 每15秒尝试潜行
	if !ai.isStealthed && ai.lastStealthTime >= 15000 {
		ai.lastStealthTime = 0
		ai.isStealthed = true
		fmt.Printf("[暴徒AI] %s 消失在阴影中...\n", ai.owner.GetName())
		// 下次攻击会是偷袭
	}

	// 偷袭攻击
	if ai.isStealthed && ai.owner.GetVictim() != nil {
		ai.isStealthed = false
		fmt.Printf("[暴徒AI] %s 从阴影中突然出现，发动偷袭！\n", ai.owner.GetName())
		// 造成额外伤害
	}
}

func (ai *ThugAI) AttackStart(target IUnit) {
	fmt.Printf("[暴徒AI] %s 狡猾地笑着：\"又有新的猎物了...\"\n", ai.owner.GetName())
}

func (ai *ThugAI) EnterCombat(target IUnit) {
	fmt.Printf("[暴徒AI] %s 抽出匕首：\"让我来解决你！\"\n", ai.owner.GetName())
}

func (ai *ThugAI) JustDied(killer IUnit) {
	fmt.Printf("[暴徒AI] %s 不甘地说：\"我...还没...展示真正的实力...\"\n", ai.owner.GetName())
}

func (ai *ThugAI) DamageTaken(attacker IUnit, damage uint32) {
	if damage > 400 {
		fmt.Printf("[暴徒AI] %s 惊讶地说：\"什么？！这不可能！\"\n", ai.owner.GetName())
	}
}

func (ai *ThugAI) DamageDealt(victim IUnit, damage uint32) {
	if damage > 500 {
		fmt.Printf("[暴徒AI] %s 得意地说：\"感受毒刃的威力吧！\"\n", ai.owner.GetName())
	}
}

// 咒术师AI - 法术攻击
type ConjurerAI struct {
	owner         *Creature
	lastSpellTime uint32
	lastHealTime  uint32
}

func NewConjurerAI(owner *Creature) *ConjurerAI {
	return &ConjurerAI{owner: owner}
}

func (ai *ConjurerAI) UpdateAI(diff uint32) {
	if !ai.owner.IsAlive() {
		return
	}

	ai.lastSpellTime += diff
	ai.lastHealTime += diff

	// 寻找目标
	if ai.owner.GetVictim() == nil {
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() {
				ai.owner.Attack(attacker)
				break
			}
		}
	}

	// 治疗技能 - 生命值低于50%时治疗自己
	if ai.owner.GetHealth() < ai.owner.GetMaxHealth()/2 && ai.lastHealTime >= 8000 {
		ai.lastHealTime = 0
		healAmount := uint32(500 + rand.Intn(300))
		newHealth := ai.owner.GetHealth() + healAmount
		if newHealth > ai.owner.GetMaxHealth() {
			newHealth = ai.owner.GetMaxHealth()
		}
		ai.owner.SetHealth(newHealth)
		fmt.Printf("[咒术师AI] %s 施放治疗术，恢复 %d 点生命值\n", ai.owner.GetName(), healAmount)
	}

	// 法术攻击 - 每3秒施放火球术
	if ai.owner.GetVictim() != nil && ai.lastSpellTime >= 3000 {
		ai.lastSpellTime = 0
		fmt.Printf("[咒术师AI] %s 开始吟唱火球术...\n", ai.owner.GetName())
		// 这里可以添加法术伤害逻辑
	}
}

func (ai *ConjurerAI) AttackStart(target IUnit) {
	fmt.Printf("[咒术师AI] %s 举起法杖：\"感受魔法的力量！\"\n", ai.owner.GetName())
}

func (ai *ConjurerAI) EnterCombat(target IUnit) {
	fmt.Printf("[咒术师AI] %s 念咒：\"黑暗之力，听从我的召唤！\"\n", ai.owner.GetName())
}

func (ai *ConjurerAI) JustDied(killer IUnit) {
	fmt.Printf("[咒术师AI] %s 临死前预言：\"黑暗...将会...降临...\"\n", ai.owner.GetName())
}

func (ai *ConjurerAI) DamageTaken(attacker IUnit, damage uint32) {
	if damage > 300 {
		fmt.Printf("[咒术师AI] %s 痛苦地说：\"我的魔法护盾！\"\n", ai.owner.GetName())
	}
}

func (ai *ConjurerAI) DamageDealt(victim IUnit, damage uint32) {
	if damage > 600 {
		fmt.Printf("[咒术师AI] %s 狂笑：\"烈焰吞噬一切！\"\n", ai.owner.GetName())
	}
}

// 精英AI - 强化版战士
type EliteAI struct {
	owner          *Creature
	lastChargeTime uint32
	lastShoutTime  uint32
}

func NewEliteAI(owner *Creature) *EliteAI {
	return &EliteAI{owner: owner}
}

func (ai *EliteAI) UpdateAI(diff uint32) {
	if !ai.owner.IsAlive() {
		return
	}

	ai.lastChargeTime += diff
	ai.lastShoutTime += diff

	// 寻找目标
	if ai.owner.GetVictim() == nil {
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() {
				ai.owner.Attack(attacker)
				break
			}
		}
	}

	// 冲锋技能 - 每12秒冲锋最远的敌人
	if ai.lastChargeTime >= 12000 && ai.owner.GetVictim() != nil {
		ai.lastChargeTime = 0
		fmt.Printf("[精英AI] %s 发动冲锋攻击 %s！\n", ai.owner.GetName(), ai.owner.GetVictim().GetName())
		// 造成额外伤害和眩晕效果
	}

	// 战吼技能 - 每20秒降低敌人攻击力
	if ai.lastShoutTime >= 20000 {
		ai.lastShoutTime = 0
		fmt.Printf("[精英AI] %s 发出震天战吼！\n", ai.owner.GetName())
		// 降低周围敌人的攻击力
	}
}

func (ai *EliteAI) AttackStart(target IUnit) {
	fmt.Printf("[精英AI] %s 拔出巨剑：\"准备受死吧，入侵者！\"\n", ai.owner.GetName())
}

func (ai *EliteAI) EnterCombat(target IUnit) {
	fmt.Printf("[精英AI] %s 怒吼：\"为了迪菲亚兄弟会的荣耀！\"\n", ai.owner.GetName())
}

func (ai *EliteAI) JustDied(killer IUnit) {
	fmt.Printf("[精英AI] %s 倒下时说：\"我...已经...尽力了...\"\n", ai.owner.GetName())
}

func (ai *EliteAI) DamageTaken(attacker IUnit, damage uint32) {
	if damage > ai.owner.GetMaxHealth()/4 {
		fmt.Printf("[精英AI] %s 愤怒地咆哮：\"你激怒我了！\"\n", ai.owner.GetName())
	}
}

func (ai *EliteAI) DamageDealt(victim IUnit, damage uint32) {
	if damage > 800 {
		fmt.Printf("[精英AI] %s 狂笑：\"这就是精英的实力！\"\n", ai.owner.GetName())
	}
}

// 范克里夫BOSS AI - 多阶段BOSS
type VanCleefAI struct {
	owner           *Creature
	lastSpecialTime uint32
	phase           uint8
	enrageTime      uint32
}

func NewVanCleefAI(owner *Creature) *VanCleefAI {
	return &VanCleefAI{owner: owner, phase: 1}
}

func (ai *VanCleefAI) UpdateAI(diff uint32) {
	if !ai.owner.IsAlive() {
		return
	}

	ai.lastSpecialTime += diff
	ai.enrageTime += diff

	// 寻找目标
	if ai.owner.GetVictim() == nil {
		for _, attacker := range ai.owner.attackers {
			if attacker.IsAlive() {
				ai.owner.Attack(attacker)
				break
			}
		}
	}

	// 根据阶段执行不同技能
	switch ai.phase {
	case 1: // 第一阶段：基础攻击
		if ai.lastSpecialTime >= 8000 {
			ai.lastSpecialTime = 0
			fmt.Printf("[范克里夫AI] %s 使用致命打击！\n", ai.owner.GetName())
			// 造成高额伤害
		}
	case 2: // 第二阶段：召唤小弟后
		if ai.lastSpecialTime >= 6000 {
			ai.lastSpecialTime = 0
			fmt.Printf("[范克里夫AI] %s 使用旋风斩！\n", ai.owner.GetName())
			// 对周围所有敌人造成伤害
		}
	case 3: // 第三阶段：狂暴
		if ai.lastSpecialTime >= 4000 {
			ai.lastSpecialTime = 0
			fmt.Printf("[范克里夫AI] %s 在狂暴状态下疯狂攻击！\n", ai.owner.GetName())
			// 攻击速度和伤害大幅提升
		}
	}

	// 狂暴计时器 - 战斗10分钟后狂暴
	if ai.enrageTime >= 600000 { // 10分钟
		fmt.Printf("[范克里夫AI] %s 进入最终狂暴状态！\n", ai.owner.GetName())
		// 大幅提升所有属性
	}
}

func (ai *VanCleefAI) SetPhase(phase uint8) {
	ai.phase = phase
}

func (ai *VanCleefAI) AttackStart(target IUnit) {
	fmt.Printf("[范克里夫AI] %s 冷笑：\"又有不知死活的冒险者来送死了...\"\n", ai.owner.GetName())
}

func (ai *VanCleefAI) EnterCombat(target IUnit) {
	fmt.Printf("[范克里夫AI] %s 拔出双刀：\"迪菲亚兄弟会的力量，你们永远不会理解！\"\n", ai.owner.GetName())
}

func (ai *VanCleefAI) JustDied(killer IUnit) {
	fmt.Printf("[范克里夫AI] %s 临死前说：\"这...不可能...我的梦想...我的兄弟会...\"\n", ai.owner.GetName())
	fmt.Printf("[范克里夫AI] %s 最后的话：\"但是...这只是开始...迪菲亚兄弟会...永远不会消失...\"\n", ai.owner.GetName())
}

func (ai *VanCleefAI) DamageTaken(attacker IUnit, damage uint32) {
	if damage > 1000 {
		fmt.Printf("[范克里夫AI] %s 愤怒地说：\"你们会为此付出血的代价！\"\n", ai.owner.GetName())
	}
}

func (ai *VanCleefAI) DamageDealt(victim IUnit, damage uint32) {
	if damage > 1200 {
		fmt.Printf("[范克里夫AI] %s 狂笑：\"感受绝望吧！这就是背叛暴风城的下场！\"\n", ai.owner.GetName())
	}
}
