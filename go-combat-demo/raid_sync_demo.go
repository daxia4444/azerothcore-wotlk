package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 简化的40人团队副本同步演示
func main() {
	fmt.Println("🎮 AzerothCore风格40人团队副本同步演示")
	fmt.Println("============================================================")

	// 创建世界
	world := NewWorld()

	// 初始化法术管理器
	GlobalSpellManager = &SpellManager{
		spells: make(map[uint32]*SpellInfo),
	}
	GlobalSpellManager.LoadSpells()

	// 创建BOSS
	boss := NewCreature("团队副本BOSS", 60, CREATURE_TYPE_DEMON)
	boss.SetMaxHealth(50000)
	boss.SetHealth(50000)
	world.AddUnit(boss)

	// 创建40个玩家模拟团队副本
	players := make([]*Player, 40)

	fmt.Println("📋 创建40人团队...")
	for i := 0; i < 40; i++ {
		// 创建不同职业的玩家
		var class uint8
		var name string
		switch i % 4 {
		case 0:
			class = CLASS_WARRIOR
			name = fmt.Sprintf("战士%d", i+1)
		case 1:
			class = CLASS_MAGE
			name = fmt.Sprintf("法师%d", i+1)
		case 2:
			class = CLASS_PRIEST
			name = fmt.Sprintf("牧师%d", i+1)
		case 3:
			class = CLASS_HUNTER
			name = fmt.Sprintf("猎人%d", i+1)
		}

		player := NewPlayer(name, 60, class)
		player.SetGUID(uint64(1000 + i))
		players[i] = player
		world.AddUnit(player)

		fmt.Printf("  ✅ %s 加入团队 (血量: %d/%d)\n",
			name, player.GetHealth(), player.GetMaxHealth())
	}

	fmt.Println("\n🎯 开始团队副本战斗演示...")
	fmt.Println("============================================================")

	// 简化的战斗演示
	demonstrateRaidSync(world, players, boss)

	fmt.Println("\n📊 AzerothCore状态同步机制总结:")
	fmt.Println("✅ 即时同步: 法术施放、伤害、治疗立即广播给所有团队成员")
	fmt.Println("✅ 状态同步: 血量、法力值变化实时更新")
	fmt.Println("✅ 视野优化: 只向相关玩家发送更新（团队副本中所有人都相关）")
	fmt.Println("✅ 队列处理: 数据包队列化处理，防止网络阻塞")
	fmt.Println("✅ 服务器权威: 所有计算在服务器端完成，防止作弊")
}

// 简化的团队副本战斗演示
func demonstrateRaidSync(world *World, players []*Player, boss *Creature) {

	// 第一阶段：坦克开怪
	fmt.Println("\n🛡️  第一阶段：坦克开怪")
	tank := players[0] // 第一个战士作为坦克

	// 坦克攻击BOSS
	fmt.Printf("⚔️  %s 对 %s 发起攻击\n", tank.GetName(), boss.GetName())
	damage := uint32(800 + rand.Intn(400))              // 800-1200伤害
	actualDamage := boss.DealDamage(tank, damage, 0, 1) // 直接伤害，物理系

	// 广播攻击状态更新（AzerothCore的SMSG_ATTACKERSTATEUPDATE）
	broadcastAttackUpdate(world, tank, boss, actualDamage)

	time.Sleep(500 * time.Millisecond)

	// 第二阶段：法师DPS输出
	fmt.Println("\n🔥 第二阶段：法师DPS输出")

	// 选择3个法师进行演示
	for i := 1; i < 4; i++ {
		if players[i].GetClass() == CLASS_MAGE {
			mage := players[i]

			// 施放寒冰箭
			spell := GlobalSpellManager.GetSpell(SPELL_FROSTBOLT)
			if spell != nil {
				fmt.Printf("❄️  %s 开始施放寒冰箭\n", mage.GetName())

				// 广播法术开始（SMSG_SPELL_START）
				broadcastSpellStart(world, mage, boss, spell)

				// 模拟施法时间
				time.Sleep(100 * time.Millisecond)

				// 法术完成，计算伤害
				damage := uint32(1200 + rand.Intn(800))              // 1200-2000伤害
				actualDamage := boss.DealDamage(mage, damage, 0, 16) // 直接伤害，冰霜系

				// 广播法术生效（SMSG_SPELL_GO）
				broadcastSpellGo(world, mage, boss, spell, actualDamage)
			}
		}
	}

	time.Sleep(500 * time.Millisecond)

	// 第三阶段：治疗阶段
	fmt.Println("\n💚 第三阶段：牧师治疗团队")

	// 模拟坦克受到伤害
	tankDamage := uint32(3000)
	actualTankDamage := tank.DealDamage(boss, tankDamage, 0, 1) // 直接伤害，物理系
	fmt.Printf("💥 %s 受到 %s 的攻击，损失 %d 血量\n",
		tank.GetName(), boss.GetName(), actualTankDamage)

	// 牧师治疗坦克
	priest := players[2] // 第一个牧师
	if priest.GetClass() == CLASS_PRIEST {
		// 施放快速治疗
		spell := GlobalSpellManager.GetSpell(SPELL_FLASH_HEAL)
		if spell != nil {
			fmt.Printf("✨ %s 对 %s 施放快速治疗\n",
				priest.GetName(), tank.GetName())

			// 广播治疗法术
			broadcastSpellStart(world, priest, tank, spell)

			// 模拟施法时间
			time.Sleep(100 * time.Millisecond)

			// 治疗生效
			healAmount := uint32(2500 + rand.Intn(1000)) // 2500-3500治疗
			tank.Heal(priest, healAmount)

			// 广播治疗效果
			broadcastHealUpdate(world, priest, tank, healAmount)
		}
	}

	time.Sleep(500 * time.Millisecond)

	// 第四阶段：AOE阶段
	fmt.Println("\n💥 第四阶段：BOSS释放AOE技能")

	// BOSS对前5个玩家造成AOE伤害（简化演示）
	aoeDamage := uint32(1500)
	fmt.Printf("🌪️  %s 释放AOE技能，对团队成员造成 %d 伤害\n",
		boss.GetName(), aoeDamage)

	// 同时更新前5个玩家的血量
	for i := 0; i < 5; i++ {
		player := players[i]
		if player.IsAlive() {
			oldHealth := player.GetHealth()
			actualDamage := player.DealDamage(boss, aoeDamage, 0, 32) // 直接伤害，暗影系
			// 广播血量更新给所有团队成员
			world.BroadcastHealthUpdate(player, oldHealth, player.GetHealth())
			fmt.Printf("💥 %s 受到AOE伤害 %d 点\n", player.GetName(), actualDamage)
		}
	}

	time.Sleep(500 * time.Millisecond)

	// 最终阶段：团队协作击败BOSS
	fmt.Println("\n🏆 最终阶段：团队协作击败BOSS")

	// 前10个玩家一起攻击
	totalDamage := uint32(0)
	for i := 0; i < 10; i++ {
		player := players[i]
		if player.IsAlive() {
			damage := uint32(500 + rand.Intn(300))                // 500-800伤害
			actualDamage := boss.DealDamage(player, damage, 0, 1) // 直接伤害，物理系
			totalDamage += actualDamage

			// 广播攻击更新
			broadcastAttackUpdate(world, player, boss, actualDamage)
		}
	}

	fmt.Printf("⚡ 团队总伤害: %d，%s 剩余血量: %d/%d\n",
		totalDamage, boss.GetName(), boss.GetHealth(), boss.GetMaxHealth())

	if boss.GetHealth() == 0 {
		fmt.Printf("🎉 恭喜！团队成功击败了 %s！\n", boss.GetName())

		// 广播BOSS死亡
		broadcastUnitDeath(world, boss)
	} else {
		fmt.Printf("💪 %s 还剩余 %d 血量，战斗继续！\n", boss.GetName(), boss.GetHealth())
	}
}

// 团队副本相关的广播方法
func broadcastAttackUpdate(w *World, attacker, target IUnit, damage uint32) {
	packet := NewWorldPacket(SMSG_ATTACKERSTATEUPDATE)
	packet.WriteUint64(attacker.GetGUID())
	packet.WriteUint64(target.GetGUID())
	packet.WriteUint32(damage)

	w.BroadcastPacket(packet)

	fmt.Printf("[网络] 广播攻击状态: %s 对 %s 造成 %d 伤害\n",
		attacker.GetName(), target.GetName(), damage)
}

func broadcastSpellStart(w *World, caster, target IUnit, spell *SpellInfo) {
	packet := NewWorldPacket(SMSG_SPELL_START)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint64(target.GetGUID())
	packet.WriteUint32(spell.ID)

	w.BroadcastPacket(packet)

	fmt.Printf("[网络] 广播法术开始: %s 对 %s 施放 %s\n",
		caster.GetName(), target.GetName(), spell.Name)
}

func broadcastSpellGo(w *World, caster, target IUnit, spell *SpellInfo, damage uint32) {
	packet := NewWorldPacket(SMSG_SPELL_START)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint64(target.GetGUID())
	packet.WriteUint32(spell.ID)
	packet.WriteUint32(damage)

	w.BroadcastPacket(packet)

	fmt.Printf("[网络] 广播法术生效: %s 的 %s 对 %s 造成 %d 伤害\n",
		caster.GetName(), spell.Name, target.GetName(), damage)
}

func broadcastHealUpdate(w *World, healer, target IUnit, healAmount uint32) {
	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint64(healer.GetGUID())
	packet.WriteUint64(target.GetGUID())
	packet.WriteUint32(3) // 更新类型：治疗
	packet.WriteUint32(healAmount)

	w.BroadcastPacket(packet)

	fmt.Printf("[网络] 广播治疗更新: %s 治疗了 %s %d 点生命值\n",
		healer.GetName(), target.GetName(), healAmount)
}

func broadcastUnitDeath(w *World, unit IUnit) {
	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint64(unit.GetGUID())
	packet.WriteUint32(1) // 死亡标记

	w.BroadcastPacket(packet)

	fmt.Printf("[网络] 广播单位死亡: %s 已死亡\n", unit.GetName())
}
