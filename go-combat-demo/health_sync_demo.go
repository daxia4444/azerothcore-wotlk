package main

import (
	"fmt"
	"time"
)

// 血量同步演示 - 基于AzerothCore的网络同步机制
func RunHealthSyncDemo() {
	fmt.Println("=== AzerothCore风格的血量同步演示 ===")

	// 初始化法术管理器
	InitSpellManager()

	// 创建世界
	world := NewWorld()

	// 创建玩家
	player1 := NewPlayer("战士玩家", 60, CLASS_WARRIOR)
	player1.SetGUID(generateGUID())
	player1.SetMaxHealth(3000)
	player1.SetHealth(3000)
	player1.SetMaxPower(POWER_RAGE, 100)
	player1.SetPower(POWER_RAGE, 0)
	player1.SetWorld(world)
	player1.x, player1.y, player1.z = 0, 0, 0

	player2 := NewPlayer("法师玩家", 60, CLASS_MAGE)
	player2.SetGUID(generateGUID())
	player2.SetMaxHealth(2500)
	player2.SetHealth(2500)
	player2.SetMaxPower(POWER_MANA, 3000)
	player2.SetPower(POWER_MANA, 3000)
	player2.SetWorld(world)
	player2.x, player2.y, player2.z = 5, 0, 0

	// 添加到世界
	world.AddUnit(player1)
	world.AddUnit(player2)

	// 模拟网络会话
	session1 := &WorldSession{id: 1}
	session2 := &WorldSession{id: 2}
	world.AddSession(session1)
	world.AddSession(session2)

	fmt.Printf("初始状态:\n")
	fmt.Printf("- %s: %d/%d HP, %d/%d 怒气\n",
		player1.GetName(), player1.GetHealth(), player1.GetMaxHealth(),
		player1.GetPower(POWER_RAGE), player1.GetMaxPower(POWER_RAGE))
	fmt.Printf("- %s: %d/%d HP, %d/%d 法力\n",
		player2.GetName(), player2.GetHealth(), player2.GetMaxHealth(),
		player2.GetPower(POWER_MANA), player2.GetMaxPower(POWER_MANA))

	fmt.Println("\n=== 演示1: 即时血量同步 ===")

	// 战士攻击法师
	fmt.Println("战士攻击法师...")
	player1.Attack(player2)

	// 模拟几次攻击
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)

		// 造成伤害 - 这会触发即时血量同步
		damage := uint32(300 + i*50)
		actualDamage := player2.DealDamage(player1, damage, DIRECT_DAMAGE, SPELL_SCHOOL_NORMAL)

		fmt.Printf("第%d次攻击: 造成%d伤害, %s剩余血量: %d/%d\n",
			i+1, actualDamage, player2.GetName(),
			player2.GetHealth(), player2.GetMaxHealth())

		// 战士获得怒气
		if i == 1 {
			fmt.Printf("%s获得怒气: %d/%d\n",
				player1.GetName(), player1.GetPower(POWER_RAGE), player1.GetMaxPower(POWER_RAGE))
		}
	}

	fmt.Println("\n=== 演示2: 法术消耗能量同步 ===")

	// 法师施放法术
	fmt.Println("法师施放寒冰箭...")
	player2.CastSpell(player1, 116) // 寒冰箭

	// 模拟施法过程
	time.Sleep(200 * time.Millisecond)

	// 消耗法力 - 这会触发即时能量同步
	manaCost := uint32(200)
	player2.ModifyPower(POWER_MANA, -int32(manaCost))

	fmt.Printf("法术消耗: %d法力, %s剩余法力: %d/%d\n",
		manaCost, player2.GetName(),
		player2.GetPower(POWER_MANA), player2.GetMaxPower(POWER_MANA))

	fmt.Println("\n=== 演示3: 治疗效果同步 ===")

	// 法师治疗自己
	fmt.Println("法师施放快速治疗...")
	healAmount := uint32(500)
	player2.Heal(player2, healAmount)

	fmt.Printf("治疗效果: +%d生命值, %s当前血量: %d/%d\n",
		healAmount, player2.GetName(),
		player2.GetHealth(), player2.GetMaxHealth())

	fmt.Println("\n=== 演示4: 定期状态同步 ===")

	fmt.Println("模拟5秒定期同步...")

	// 模拟世界更新循环
	totalTime := uint32(0)
	updateInterval := uint32(100) // 100ms更新间隔

	for totalTime < 5000 { // 5秒
		world.Update(updateInterval)
		totalTime += updateInterval
		time.Sleep(time.Duration(updateInterval) * time.Millisecond)

		// 每秒显示一次状态
		if totalTime%1000 == 0 {
			fmt.Printf("[%ds] 状态检查完成\n", totalTime/1000)
		}
	}

	fmt.Println("\n=== 演示5: 并发血量变化 ===")

	fmt.Println("模拟多个同时的血量变化...")

	// 同时进行多个操作
	go func() {
		// 持续伤害效果
		for i := 0; i < 5; i++ {
			time.Sleep(200 * time.Millisecond)
			player2.ModifyHealth(-50) // DOT伤害
			fmt.Printf("[DOT] %s受到持续伤害: %d/%d\n",
				player2.GetName(), player2.GetHealth(), player2.GetMaxHealth())
		}
	}()

	go func() {
		// 持续治疗效果
		for i := 0; i < 3; i++ {
			time.Sleep(300 * time.Millisecond)
			player2.ModifyHealth(80) // HOT治疗
			fmt.Printf("[HOT] %s受到持续治疗: %d/%d\n",
				player2.GetName(), player2.GetHealth(), player2.GetMaxHealth())
		}
	}()

	// 等待并发操作完成
	time.Sleep(1500 * time.Millisecond)

	fmt.Println("\n=== 同步机制总结 ===")
	fmt.Println("✓ 即时同步: 血量/能量变化时立即广播")
	fmt.Println("✓ 定期同步: 每5秒广播完整状态")
	fmt.Println("✓ 并发安全: 支持多个同时的状态变化")
	fmt.Println("✓ 网络优化: 只在实际变化时发送更新")

	fmt.Printf("\n最终状态:\n")
	fmt.Printf("- %s: %d/%d HP, %d/%d 怒气\n",
		player1.GetName(), player1.GetHealth(), player1.GetMaxHealth(),
		player1.GetPower(POWER_RAGE), player1.GetMaxPower(POWER_RAGE))
	fmt.Printf("- %s: %d/%d HP, %d/%d 法力\n",
		player2.GetName(), player2.GetHealth(), player2.GetMaxHealth(),
		player2.GetPower(POWER_MANA), player2.GetMaxPower(POWER_MANA))

	fmt.Println("\n=== 演示完成 ===")
}

// 扩展WorldSession以支持数据包发送(仅用于演示)
func (ws *WorldSession) SendPacketDemo(packet *WorldPacket) {
	// 模拟发送数据包到客户端
	opcodeName := getOpcodeNameForDemo(packet.opcode)
	fmt.Printf("[网络] -> 客户端%d: %s\n", ws.id, opcodeName)
}

// 获取操作码名称用于演示
func getOpcodeNameForDemo(opcode uint16) string {
	switch opcode {
	case SMSG_UPDATE_OBJECT:
		return "对象状态更新"
	case SMSG_ATTACKERSTATEUPDATE:
		return "攻击状态更新"
	case SMSG_SPELL_START:
		return "法术开始"
	case SMSG_SPELLGO:
		return "法术生效"
	case SMSG_POWER_UPDATE:
		return "能量更新"
	case SMSG_HEALTH_UPDATE:
		return "血量更新"
	default:
		return fmt.Sprintf("未知消息(0x%X)", opcode)
	}
}

// 主函数中调用演示
func init() {
	// 可以在main函数中调用 RunHealthSyncDemo()
}
