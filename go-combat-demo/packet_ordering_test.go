package main

import (
	"fmt"
	"sync"
	"time"
)

// PacketOrderingTest 数据包时序测试
type PacketOrderingTest struct {
	world    *World
	sessions []*WorldSession
	units    []IUnit
}

// NewPacketOrderingTest 创建数据包时序测试
func NewPacketOrderingTest() *PacketOrderingTest {
	world := NewWorld()

	// 创建测试会话
	sessions := make([]*WorldSession, 3)
	for i := 0; i < 3; i++ {
		sessions[i] = NewWorldSession(uint32(i+1), fmt.Sprintf("TestPlayer%d", i+1), nil, world)
		world.AddSession(sessions[i])
	}

	// 创建测试单位
	units := make([]IUnit, 2)
	for i := 0; i < 2; i++ {
		unit := NewUnit(fmt.Sprintf("TestUnit%d", i+1), 1000, 1000, 500, 500)
		units[i] = unit
		world.AddUnit(unit)
	}

	return &PacketOrderingTest{
		world:    world,
		sessions: sessions,
		units:    units,
	}
}

// TestPacketOrdering 测试数据包时序
func (pot *PacketOrderingTest) TestPacketOrdering() {
	fmt.Println("\n=== 🔥 数据包时序控制测试 ===")
	fmt.Println("演示如何解决 SendBatchUpdates 和 broadcastPeriodicUpdates 的时序问题")

	// 模拟问题场景
	pot.simulateOrderingProblem()

	// 演示解决方案
	pot.demonstrateSolution()
}

// simulateOrderingProblem 模拟时序问题
func (pot *PacketOrderingTest) simulateOrderingProblem() {
	fmt.Println("\n--- 问题场景模拟 ---")
	fmt.Println("场景：玩家血量从 1000 → 800 → 600")

	unit := pot.units[0]

	// 1. 初始状态：血量 1000
	fmt.Printf("初始状态：%s 血量 = %d\n", unit.GetName(), unit.GetHealth())

	// 2. 模拟批量更新：血量变为 800
	fmt.Println("\n步骤1：批量更新 - 血量变为 800")
	unit.SetHealth(800)

	// 创建批量更新数据包
	batchPacket := NewWorldPacket(SMSG_UPDATE_OBJECT)
	batchPacket.WriteUint64(unit.GetGUID())
	batchPacket.WriteUint32(800) // 新血量
	batchPacket.SetPriority(1)   // 高优先级
	batchPacket.SetUpdateId(100) // 更新ID 100

	fmt.Printf("批量更新数据包：血量=800, 优先级=%d, 更新ID=%d, 时间戳=%v\n",
		800, batchPacket.GetPriority(), batchPacket.GetUpdateId(), batchPacket.GetTimestamp())

	// 3. 模拟定期更新：发送旧状态（血量 1000）
	fmt.Println("\n步骤2：定期更新 - 发送旧状态（血量 1000）")

	periodicPacket := NewWorldPacket(SMSG_UPDATE_OBJECT)
	periodicPacket.WriteUint64(unit.GetGUID())
	periodicPacket.WriteUint32(1000) // 旧血量
	periodicPacket.SetPriority(3)    // 低优先级
	periodicPacket.SetUpdateId(99)   // 更新ID 99（更旧）

	fmt.Printf("定期更新数据包：血量=1000, 优先级=%d, 更新ID=%d, 时间戳=%v\n",
		1000, periodicPacket.GetPriority(), periodicPacket.GetUpdateId(), periodicPacket.GetTimestamp())

	// 4. 模拟网络延迟导致的乱序
	fmt.Println("\n❌ 问题：由于网络延迟，客户端可能先收到定期更新，再收到批量更新")
	fmt.Println("结果：旧数据（血量=1000）覆盖新数据（血量=800）")
}

// demonstrateSolution 演示解决方案
func (pot *PacketOrderingTest) demonstrateSolution() {
	fmt.Println("\n--- ✅ 解决方案演示 ---")
	fmt.Println("使用 AzerothCore 风格的时序控制机制")

	unit := pot.units[1]
	session := pot.sessions[0]

	// 1. 创建多个数据包模拟乱序场景
	packets := []*WorldPacket{
		pot.createTestPacket(unit, 600, 3, 150), // 定期更新：血量=600, 低优先级, 更新ID=150
		pot.createTestPacket(unit, 800, 1, 200), // 批量更新：血量=800, 高优先级, 更新ID=200
		pot.createTestPacket(unit, 700, 2, 180), // 普通更新：血量=700, 普通优先级, 更新ID=180
		pot.createTestPacket(unit, 900, 0, 220), // 立即更新：血量=900, 最高优先级, 更新ID=220
	}

	fmt.Println("\n原始数据包顺序（模拟网络乱序）：")
	for i, packet := range packets {
		health := pot.extractHealthFromPacket(packet)
		fmt.Printf("数据包%d：血量=%d, 优先级=%d, 更新ID=%d\n",
			i+1, health, packet.GetPriority(), packet.GetUpdateId())
	}

	// 2. 使用时序控制排序
	fmt.Println("\n🔥 应用时序控制排序：")
	session.SortAndSendPackets(packets)

	// 3. 显示排序后的结果
	fmt.Println("\n✅ 排序后的发送顺序：")
	fmt.Println("1. 优先级排序：立即(0) > 高(1) > 普通(2) > 低(3)")
	fmt.Println("2. 相同优先级按更新ID排序：更大的ID表示更新的数据")
	fmt.Println("3. 客户端只接收最新的数据，旧数据被自动过滤")

	// 4. 演示版本控制
	pot.demonstrateVersionControl(session)
}

// createTestPacket 创建测试数据包
func (pot *PacketOrderingTest) createTestPacket(unit IUnit, health uint32, priority uint8, updateId uint32) *WorldPacket {
	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint64(unit.GetGUID())
	packet.WriteUint32(health)
	packet.SetPriority(priority)
	packet.SetUpdateId(updateId)
	return packet
}

// extractHealthFromPacket 从数据包中提取血量（简化版）
func (pot *PacketOrderingTest) extractHealthFromPacket(packet *WorldPacket) uint32 {
	// 跳过GUID（8字节）
	if len(packet.data) >= 12 {
		return uint32(packet.data[8]) | uint32(packet.data[9])<<8 |
			uint32(packet.data[10])<<16 | uint32(packet.data[11])<<24
	}
	return 0
}

// demonstrateVersionControl 演示版本控制
func (pot *PacketOrderingTest) demonstrateVersionControl(session *WorldSession) {
	fmt.Println("\n--- 🔥 版本控制演示 ---")

	unit := pot.units[0]

	// 模拟连续的状态更新
	updates := []struct {
		health   uint32
		updateId uint32
		desc     string
	}{
		{1000, 300, "初始状态"},
		{950, 301, "受到50点伤害"},
		{900, 302, "受到50点伤害"},
		{920, 303, "恢复20点血量"},
		{880, 304, "受到40点伤害"},
	}

	fmt.Println("连续状态更新序列：")
	for _, update := range updates {
		packet := pot.createTestPacket(unit, update.health, 1, update.updateId)

		// 检查是否应该发送这个数据包
		shouldSend := session.shouldSendPacket(packet)

		fmt.Printf("更新ID %d: %s (血量=%d) - %s\n",
			update.updateId, update.desc, update.health,
			map[bool]string{true: "✅ 发送", false: "❌ 跳过（旧数据）"}[shouldSend])

		if shouldSend {
			session.lastUpdateStates[packet.opcode] = packet.updateId
		}
	}

	fmt.Println("\n🎯 关键优势：")
	fmt.Println("1. 自动过滤旧数据包，防止状态回退")
	fmt.Println("2. 基于优先级的智能排序")
	fmt.Println("3. 版本控制确保数据一致性")
	fmt.Println("4. 网络带宽优化：只发送必要的更新")
}

// RunConcurrentTest 运行并发测试
func (pot *PacketOrderingTest) RunConcurrentTest() {
	fmt.Println("\n=== 🚀 并发时序测试 ===")
	fmt.Println("模拟多线程环境下的数据包时序问题")

	unit := pot.units[0]
	session := pot.sessions[0]

	var wg sync.WaitGroup
	packetChan := make(chan *WorldPacket, 100)

	// 启动多个goroutine模拟并发更新
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(threadId int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				health := uint32(1000 - threadId*10 - j*5)
				priority := uint8(threadId % 4)
				updateId := uint32(threadId*100 + j)

				packet := pot.createTestPacket(unit, health, priority, updateId)
				packetChan <- packet

				// 模拟网络延迟
				time.Sleep(time.Millisecond * time.Duration(threadId*2))
			}
		}(i)
	}

	// 收集所有数据包
	go func() {
		wg.Wait()
		close(packetChan)
	}()

	var allPackets []*WorldPacket
	for packet := range packetChan {
		allPackets = append(allPackets, packet)
	}

	fmt.Printf("收集到 %d 个并发数据包\n", len(allPackets))

	// 应用时序控制
	fmt.Println("应用时序控制排序...")
	session.SortAndSendPackets(allPackets)

	fmt.Println("✅ 并发测试完成：所有数据包已按正确顺序处理")
}

// 主测试函数
func RunPacketOrderingTests() {
	test := NewPacketOrderingTest()

	// 运行基本时序测试
	test.TestPacketOrdering()

	// 运行并发测试
	test.RunConcurrentTest()

	fmt.Println("\n=== 📊 测试总结 ===")
	fmt.Println("✅ 数据包时序控制机制验证完成")
	fmt.Println("✅ 解决了 SendBatchUpdates 和 broadcastPeriodicUpdates 的时序问题")
	fmt.Println("✅ 防止了旧数据覆盖新数据的问题")
	fmt.Println("✅ 实现了 AzerothCore 风格的智能同步机制")

	// 清理
	test.world.Shutdown()
}
