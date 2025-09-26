package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	// 初始化随机种子
	rand.Seed(time.Now().UnixNano())

	fmt.Println("🌐 === AzerothCore风格的客户端-服务器战斗系统演示 === 🌐")
	fmt.Println("基于WorldSession和网络通信的Go语言实现")
	fmt.Println("🔮 包含完整的法术系统：即时法术、施法法术、引导法术")

	// 首先运行血量同步演示
	fmt.Println("\n💓 === 血量同步机制演示 === 💓")
	RunHealthSyncDemo()

	fmt.Println("\n\n🎮 === 开始网络服务器演示 === 🎮")

	// 初始化法术系统
	fmt.Println("⚡ 初始化法术管理器...")
	InitSpellManager()

	// 创建世界管理器
	world := NewWorld()

	// 启动游戏服务器
	server := NewGameServer(world)
	err := server.Start("localhost:8080")
	if err != nil {
		fmt.Printf("启动服务器失败: %v\n", err)
		return
	}
	defer server.Stop()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	fmt.Println("\n🎮 === 创建5个客户端连接 === 🎮")

	// 创建5个游戏客户端
	clients := make([]*GameClient, 5)
	players := make([]*Player, 5)
	classNames := []string{"战士", "法师", "牧师", "猎人", "术士"}
	playerNames := []string{"钢铁卫士", "烈焰法师", "圣光牧师", "神射手", "暗影术士"}
	classes := []uint8{CLASS_WARRIOR, CLASS_MAGE, CLASS_PRIEST, CLASS_HUNTER, CLASS_WARLOCK}

	// 连接客户端到服务器
	for i := 0; i < 5; i++ {
		clients[i] = NewGameClient(uint32(i+1), playerNames[i], world)
		err := clients[i].Connect("localhost:8080")
		if err != nil {
			fmt.Printf("客户端 %s 连接失败: %v\n", playerNames[i], err)
			continue
		}

		// 创建玩家角色
		players[i] = NewPlayer(playerNames[i], 22, classes[i])
		players[i].SetMaxHealth([]uint32{8000, 4500, 5000, 5500, 4800}[i])
		players[i].SetHealth(players[i].GetMaxHealth())

		// 设置能量值
		switch classes[i] {
		case CLASS_WARRIOR:
			players[i].SetMaxPower(POWER_RAGE, 100)
			players[i].SetPower(POWER_RAGE, 0)
		case CLASS_MAGE, CLASS_PRIEST, CLASS_WARLOCK:
			maxMana := []uint32{0, 6000, 7000, 0, 6500}[i]
			players[i].SetMaxPower(POWER_MANA, maxMana)
			players[i].SetPower(POWER_MANA, maxMana)
		case CLASS_HUNTER:
			players[i].SetMaxPower(POWER_MANA, 4000)
			players[i].SetPower(POWER_MANA, 4000)
		}

		// 设置世界引用（法术系统需要）
		players[i].SetWorld(world)

		// 登录玩家到客户端
		clients[i].Login(players[i])
		world.AddUnit(players[i])

		fmt.Printf("  ✅ %s (%s) 已连接并登录\n", playerNames[i], classNames[i])
		time.Sleep(200 * time.Millisecond) // 避免连接过快
	}

	fmt.Printf("\n📊 服务器状态: %d 个活跃会话\n", server.GetSessionCount())

	// 等待所有连接稳定
	time.Sleep(2 * time.Second)

	fmt.Println("\n⚔️ === 开始网络战斗演示 === ⚔️")

	// === 第一阶段：目标选择演示 ===
	fmt.Println("\n--- 🎯 第一阶段：客户端目标选择 ---")

	// 创建一些敌人
	enemies := make([]*Creature, 3)
	enemyNames := []string{"迪菲亚矿工", "迪菲亚精英", "埃德温·范克里夫"}
	enemyLevels := []uint8{18, 22, 26}
	enemyHealth := []uint32{1200, 2500, 8000}

	for i := 0; i < 3; i++ {
		enemies[i] = NewCreature(enemyNames[i], enemyLevels[i], CREATURE_TYPE_HUMANOID)
		enemies[i].SetMaxHealth(enemyHealth[i])
		enemies[i].SetHealth(enemyHealth[i])
		enemies[i].SetAI(NewCreatureAI(enemies[i]))
		enemies[i].SetWorld(world) // 设置世界引用
		world.AddUnit(enemies[i])
	}

	// 客户端选择目标
	for i, client := range clients {
		if client.IsRunning() {
			targetIndex := i % len(enemies)
			client.SetTarget(enemies[targetIndex])
			fmt.Printf("📡 客户端 %s 选择目标: %s\n", playerNames[i], enemyNames[targetIndex])
		}
	}

	time.Sleep(1 * time.Second)

	// === 第二阶段：攻击指令演示 ===
	fmt.Println("\n--- ⚔️ 第二阶段：客户端发起攻击 ---")

	for i, client := range clients {
		if client.IsRunning() && client.GetTarget() != nil {
			client.Attack(client.GetTarget())
			fmt.Printf("📡 客户端 %s 发送攻击指令\n", playerNames[i])
		}
	}

	// 运行战斗循环
	fmt.Println("\n🔥 网络战斗进行中...")
	combatTime := 0
	maxCombatTime := 30000 // 30秒

	for combatTime < maxCombatTime {
		// 更新所有客户端
		for _, client := range clients {
			if client.IsRunning() {
				client.Update(200)
			}
		}

		// 检查敌人状态
		aliveEnemies := 0
		for _, enemy := range enemies {
			if enemy.IsAlive() {
				aliveEnemies++
			}
		}

		if aliveEnemies == 0 {
			fmt.Println("🎉 所有敌人被击败！")
			break
		}

		// 检查玩家状态
		alivePlayers := 0
		for _, player := range players {
			if player.IsAlive() {
				alivePlayers++
			}
		}

		if alivePlayers == 0 {
			fmt.Println("💀 所有玩家阵亡！")
			break
		}

		combatTime += 200
		time.Sleep(200 * time.Millisecond)

		// 每5秒显示状态
		if combatTime%5000 == 0 {
			fmt.Printf("⚔️ 战斗进行中... 存活敌人: %d, 存活玩家: %d\n", aliveEnemies, alivePlayers)
		}
	}

	// === 第三阶段：法术施放演示 ===
	fmt.Println("\n--- 🔮 第三阶段：客户端法术施放演示 ---")

	// 法师施放寒冰箭（施法法术）
	if clients[1].IsRunning() && len(enemies) > 0 && enemies[0].IsAlive() {
		clients[1].CastSpell(SPELL_FROSTBOLT, enemies[0])
		fmt.Printf("📡 客户端 %s 施放寒冰箭（2.5秒施法时间）\n", playerNames[1])
	}

	time.Sleep(1 * time.Second)

	// 牧师施放快速治疗（施法法术）
	if clients[2].IsRunning() && players[0].IsAlive() {
		clients[2].CastSpell(SPELL_FLASH_HEAL, players[0])
		fmt.Printf("📡 客户端 %s 施放快速治疗（1.5秒施法时间）\n", playerNames[2])
	}

	time.Sleep(1 * time.Second)

	// 法师施放冰霜新星（即时法术）
	if clients[1].IsRunning() {
		clients[1].CastSpell(SPELL_FROST_NOVA, clients[1].GetPlayer())
		fmt.Printf("📡 客户端 %s 施放冰霜新星（即时法术）\n", playerNames[1])
	}

	time.Sleep(1 * time.Second)

	// 战士使用嘲讽（即时技能）
	if clients[0].IsRunning() && len(enemies) > 1 && enemies[1].IsAlive() {
		clients[0].CastSpell(SPELL_TAUNT, enemies[1])
		fmt.Printf("📡 客户端 %s 使用嘲讽（即时技能）\n", playerNames[0])
	}

	time.Sleep(1 * time.Second)

	// 术士施放暗影箭（施法法术）
	if clients[4].IsRunning() && len(enemies) > 0 && enemies[0].IsAlive() {
		clients[4].CastSpell(SPELL_SHADOW_BOLT, enemies[0])
		fmt.Printf("📡 客户端 %s 施放暗影箭（2.5秒施法时间）\n", playerNames[4])
	}

	time.Sleep(1 * time.Second)

	// 猎人使用瞄准射击（施法技能）
	if clients[3].IsRunning() && len(enemies) > 1 && enemies[1].IsAlive() {
		clients[3].CastSpell(SPELL_AIMED_SHOT, enemies[1])
		fmt.Printf("📡 客户端 %s 使用瞄准射击（3秒施法时间）\n", playerNames[3])
	}

	time.Sleep(1 * time.Second)

	// 牧师施放真言术：盾（即时法术）
	if clients[2].IsRunning() && players[1].IsAlive() {
		clients[2].CastSpell(SPELL_POWER_WORD_SHIELD, players[1])
		fmt.Printf("📡 客户端 %s 为法师施放真言术：盾（即时法术）\n", playerNames[2])
	}

	time.Sleep(1 * time.Second)

	// 法师施放暴风雪（引导法术）
	if clients[1].IsRunning() && len(enemies) > 0 {
		clients[1].CastSpell(SPELL_BLIZZARD, enemies[0])
		fmt.Printf("📡 客户端 %s 施放暴风雪（8秒引导法术）\n", playerNames[1])
	}

	// 等待法术施放完成
	fmt.Println("⏳ 等待法术施放完成...")
	time.Sleep(5 * time.Second)

	// === 第四阶段：保持连接演示 ===
	fmt.Println("\n--- 💓 第四阶段：保持连接心跳 ---")

	for i, client := range clients {
		if client.IsRunning() {
			client.SendKeepAlive()
			fmt.Printf("📡 客户端 %s 发送心跳包\n", playerNames[i])
		}
	}

	time.Sleep(1 * time.Second)

	// === 显示最终结果 ===
	fmt.Println("\n📊 === 最终战斗结果 === 📊")

	fmt.Println("\n🛡️ 玩家状态:")
	for i, player := range players {
		if player.IsAlive() {
			fmt.Printf("  ✅ %s (%s) - 生命值: %d/%d\n",
				player.GetName(), classNames[i], player.GetHealth(), player.GetMaxHealth())
		} else {
			fmt.Printf("  💀 %s (%s) - 已阵亡\n",
				player.GetName(), classNames[i])
		}
	}

	fmt.Println("\n👹 敌人状态:")
	for _, enemy := range enemies {
		if enemy.IsAlive() {
			fmt.Printf("  ⚠️  %s - 生命值: %d/%d\n",
				enemy.GetName(), enemy.GetHealth(), enemy.GetMaxHealth())
		} else {
			fmt.Printf("  💀 %s - 已被击败\n", enemy.GetName())
		}
	}

	// 断开所有客户端
	fmt.Println("\n🔌 === 断开客户端连接 === 🔌")
	for i, client := range clients {
		if client.IsRunning() {
			client.Disconnect()
			fmt.Printf("  📡 客户端 %s 已断开\n", playerNames[i])
		}
	}

	time.Sleep(1 * time.Second)

	fmt.Printf("\n📊 服务器最终状态: %d 个活跃会话\n", server.GetSessionCount())

	fmt.Println("\n🎊 === 网络战斗系统演示完成！=== 🎊")
	fmt.Println("✨ 成功演示了基于AzerothCore架构的客户端-服务器通信")
	fmt.Println("📋 包含功能:")
	fmt.Println("   • WorldSession会话管理")
	fmt.Println("   • 操作码(Opcode)处理")
	fmt.Println("   • 数据包(WorldPacket)通信")
	fmt.Println("   • 客户端操作指令")
	fmt.Println("   • 服务器响应处理")
	fmt.Println("   • 多玩家并发支持")
	fmt.Println("   • 完整法术系统:")
	fmt.Println("     - 即时法术（冰霜新星、真言术：盾）")
	fmt.Println("     - 施法法术（寒冰箭、火球术、治疗术）")
	fmt.Println("     - 引导法术（暴风雪）")
	fmt.Println("     - 法术冷却系统")
	fmt.Println("     - 伤害计算与等级加成")
	fmt.Println("     - 法力消耗与能量管理")
	fmt.Println("     - 法术打断机制")
	fmt.Println("   • 职业技能系统:")
	fmt.Println("     - 战士（英勇打击、嘲讽）")
	fmt.Println("     - 法师（寒冰箭、火球术、暴风雪、冰霜新星）")
	fmt.Println("     - 牧师（治疗术、快速治疗、真言术：盾）")
	fmt.Println("     - 猎人（瞄准射击、多重射击）")
	fmt.Println("     - 术士（暗影箭、献祭、恐惧术）")
	fmt.Println("   • 网络同步:")
	fmt.Println("     - 法术开始广播（SMSG_SPELL_START）")
	fmt.Println("     - 法术生效广播（SMSG_SPELL_GO）")
	fmt.Println("     - 伤害/治疗同步")
	fmt.Println("     - 状态更新广播")
}
