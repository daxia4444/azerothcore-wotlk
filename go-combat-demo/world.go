package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// 全局GUID生成器
var globalGUID uint64 = 1000

func generateGUID() uint64 {
	return atomic.AddUint64(&globalGUID, 1)
}

// UpdateData - 基于AzerothCore的UpdateData类，用于批量收集更新
type UpdateData struct {
	blocks map[uint32][]byte // 为每个玩家收集的更新块
	mutex  sync.RWMutex
}

func NewUpdateData() *UpdateData {
	return &UpdateData{
		blocks: make(map[uint32][]byte),
	}
}

// AddUpdateBlock 为特定玩家添加更新块
func (ud *UpdateData) AddUpdateBlock(sessionId uint32, block []byte) {
	ud.mutex.Lock()
	defer ud.mutex.Unlock()

	if existing, exists := ud.blocks[sessionId]; exists {
		// 合并更新块
		ud.blocks[sessionId] = append(existing, block...)
	} else {
		ud.blocks[sessionId] = block
	}
}

// BuildPacket 为特定玩家构建数据包
func (ud *UpdateData) BuildPacket(sessionId uint32) *WorldPacket {
	ud.mutex.RLock()
	defer ud.mutex.RUnlock()

	if block, exists := ud.blocks[sessionId]; exists && len(block) > 0 {
		packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
		packet.WriteUint32(uint32(len(block))) // 数据块大小
		packet.data = append(packet.data, block...)
		packet.wpos += len(block)
		return packet
	}
	return nil
}

// Clear 清空更新数据
func (ud *UpdateData) Clear() {
	ud.mutex.Lock()
	defer ud.mutex.Unlock()
	ud.blocks = make(map[uint32][]byte)
}

// HasUpdates 检查是否有更新
func (ud *UpdateData) HasUpdates() bool {
	ud.mutex.RLock()
	defer ud.mutex.RUnlock()
	return len(ud.blocks) > 0
}

// GetSessionCount 获取有更新的会话数量
func (ud *UpdateData) GetSessionCount() int {
	ud.mutex.RLock()
	defer ud.mutex.RUnlock()
	return len(ud.blocks)
}

// BatchUpdate 批量更新结构
type BatchUpdate struct {
	unitGUID   uint64
	updateType string // "health", "power", "spell", "attack"
	data       []byte
	targets    []uint32 // 目标会话ID列表
	timestamp  time.Time
}

// NewBatchUpdate 创建批量更新
func NewBatchUpdate(unitGUID uint64, updateType string, data []byte, targets []uint32) *BatchUpdate {
	return &BatchUpdate{
		unitGUID:   unitGUID,
		updateType: updateType,
		data:       data,
		targets:    targets,
		timestamp:  time.Now(),
	}
}

// BatchSyncManager 批量同步管理器 - 基于AzerothCore的批量同步机制
type BatchSyncManager struct {
	updateQueue    chan *BatchUpdate
	immediateQueue chan *BatchUpdate // 立即同步队列
	stopChan       chan bool
	batchInterval  time.Duration
	maxBatchSize   int
	maxQueueSize   int
	world          *World
	mutex          sync.RWMutex
	isRunning      bool
	statistics     *BatchSyncStats
}

// BatchSyncStats 批量同步统计
type BatchSyncStats struct {
	batchUpdatesSent     uint64
	immediateUpdatesSent uint64
	totalPacketsSent     uint64
	batchesProcessed     uint64
	averageLatency       time.Duration
	mutex                sync.RWMutex
}

// NewBatchSyncManager 创建批量同步管理器
func NewBatchSyncManager(world *World) *BatchSyncManager {
	return &BatchSyncManager{
		updateQueue:    make(chan *BatchUpdate, 1000),
		immediateQueue: make(chan *BatchUpdate, 200),
		stopChan:       make(chan bool),
		batchInterval:  100 * time.Millisecond, // 100ms批量间隔
		maxBatchSize:   150,                    // 最大批量大小
		maxQueueSize:   1000,
		world:          world,
		isRunning:      false,
		statistics:     &BatchSyncStats{},
	}
}

// Start 启动批量同步管理器
func (bsm *BatchSyncManager) Start() {
	bsm.mutex.Lock()
	defer bsm.mutex.Unlock()

	if bsm.isRunning {
		return
	}

	bsm.isRunning = true
	go bsm.batchProcessor()
	go bsm.immediateProcessor()
	fmt.Println("[BatchSync] 批量同步管理器已启动")
}

// Stop 停止批量同步管理器
func (bsm *BatchSyncManager) Stop() {
	bsm.mutex.Lock()
	defer bsm.mutex.Unlock()

	if !bsm.isRunning {
		return
	}

	bsm.isRunning = false
	close(bsm.stopChan)
	fmt.Println("[BatchSync] 批量同步管理器已停止")
}

// QueueBatchUpdate 队列批量更新
func (bsm *BatchSyncManager) QueueBatchUpdate(update *BatchUpdate) {
	select {
	case bsm.updateQueue <- update:
		// 成功加入队列
	default:
		fmt.Printf("[BatchSync] 批量更新队列已满，丢弃更新: %s\n", update.updateType)
	}
}

// QueueImmediateUpdate 队列立即更新
func (bsm *BatchSyncManager) QueueImmediateUpdate(update *BatchUpdate) {
	select {
	case bsm.immediateQueue <- update:
		// 成功加入立即队列
	default:
		fmt.Printf("[BatchSync] 立即更新队列已满，丢弃更新: %s\n", update.updateType)
	}
}

// batchProcessor 批量处理器
func (bsm *BatchSyncManager) batchProcessor() {
	ticker := time.NewTicker(bsm.batchInterval)
	defer ticker.Stop()

	batchBuffer := make([]*BatchUpdate, 0, bsm.maxBatchSize)

	for {
		select {
		case <-bsm.stopChan:
			return
		case update := <-bsm.updateQueue:
			batchBuffer = append(batchBuffer, update)
			if len(batchBuffer) >= bsm.maxBatchSize {
				bsm.processBatch(batchBuffer)
				batchBuffer = batchBuffer[:0]
			}
		case <-ticker.C:
			if len(batchBuffer) > 0 {
				bsm.processBatch(batchBuffer)
				batchBuffer = batchBuffer[:0]
			}
		}
	}
}

// immediateProcessor 立即处理器
func (bsm *BatchSyncManager) immediateProcessor() {
	for {
		select {
		case <-bsm.stopChan:
			return
		case update := <-bsm.immediateQueue:
			bsm.processImmediateUpdate(update)
		}
	}
}

// processBatch 处理批量更新
func (bsm *BatchSyncManager) processBatch(batch []*BatchUpdate) {
	if len(batch) == 0 {
		return
	}

	startTime := time.Now()
	packetsSent := 0

	// 按会话ID分组更新
	sessionUpdates := make(map[uint32][]*BatchUpdate)
	for _, update := range batch {
		for _, sessionId := range update.targets {
			sessionUpdates[sessionId] = append(sessionUpdates[sessionId], update)
		}
	}

	// 为每个会话发送合并的更新包
	for sessionId, updates := range sessionUpdates {
		if session := bsm.world.GetSession(sessionId); session != nil && session.IsConnected() {
			packet := bsm.buildBatchPacket(updates)
			if packet != nil {
				session.SendPacket(packet)
				packetsSent++
			}
		}
	}

	// 更新统计
	bsm.statistics.mutex.Lock()
	bsm.statistics.batchesProcessed++
	bsm.statistics.batchUpdatesSent += uint64(len(batch))
	bsm.statistics.totalPacketsSent += uint64(packetsSent)
	bsm.statistics.averageLatency = time.Since(startTime)
	bsm.statistics.mutex.Unlock()

	fmt.Printf("[BatchSync] 处理批量更新: %d个更新 -> %d个数据包 (耗时: %v)\n",
		len(batch), packetsSent, time.Since(startTime))
}

// processImmediateUpdate 处理立即更新
func (bsm *BatchSyncManager) processImmediateUpdate(update *BatchUpdate) {
	packetsSent := 0

	for _, sessionId := range update.targets {
		if session := bsm.world.GetSession(sessionId); session != nil && session.IsConnected() {
			packet := bsm.buildSinglePacket(update)
			if packet != nil {
				session.SendPacket(packet)
				packetsSent++
			}
		}
	}

	// 更新统计
	bsm.statistics.mutex.Lock()
	bsm.statistics.immediateUpdatesSent++
	bsm.statistics.totalPacketsSent += uint64(packetsSent)
	bsm.statistics.mutex.Unlock()

	fmt.Printf("[BatchSync] 立即更新: %s -> %d个数据包\n", update.updateType, packetsSent)
}

// buildBatchPacket 构建批量数据包
func (bsm *BatchSyncManager) buildBatchPacket(updates []*BatchUpdate) *WorldPacket {
	if len(updates) == 0 {
		return nil
	}

	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint32(uint32(len(updates))) // 更新数量

	for _, update := range updates {
		packet.WriteUint64(update.unitGUID)
		packet.WriteString(update.updateType)
		packet.WriteUint32(uint32(len(update.data)))
		packet.data = append(packet.data, update.data...)
		packet.wpos += len(update.data)
	}

	return packet
}

// buildSinglePacket 构建单个数据包
func (bsm *BatchSyncManager) buildSinglePacket(update *BatchUpdate) *WorldPacket {
	var opcode uint16
	switch update.updateType {
	case "health":
		opcode = SMSG_HEALTH_UPDATE
	case "power":
		opcode = SMSG_POWER_UPDATE
	case "spell":
		opcode = SMSG_SPELLGO
	case "attack":
		opcode = SMSG_ATTACKERSTATEUPDATE
	default:
		opcode = SMSG_UPDATE_OBJECT
	}

	packet := NewWorldPacket(opcode)
	packet.WriteUint64(update.unitGUID)
	packet.data = append(packet.data, update.data...)
	packet.wpos += len(update.data)

	return packet
}

// GetStatistics 获取统计信息
func (bsm *BatchSyncManager) GetStatistics() BatchSyncStats {
	bsm.statistics.mutex.RLock()
	defer bsm.statistics.mutex.RUnlock()
	return *bsm.statistics
}

// PrintStatistics 打印统计信息
func (bsm *BatchSyncManager) PrintStatistics() {
	stats := bsm.GetStatistics()
	fmt.Printf("\n=== 批量同步统计 ===\n")
	fmt.Printf("批量更新发送: %d\n", stats.batchUpdatesSent)
	fmt.Printf("立即更新发送: %d\n", stats.immediateUpdatesSent)
	fmt.Printf("总数据包发送: %d\n", stats.totalPacketsSent)
	fmt.Printf("批次处理数: %d\n", stats.batchesProcessed)
	fmt.Printf("平均延迟: %v\n", stats.averageLatency)
}

// 世界管理器 - 基于AzerothCore的World类，包含批量更新机制
type World struct {
	units               map[uint64]IUnit         // 所有单位的映射
	sessions            map[uint32]*WorldSession // 所有会话的映射
	pendingUpdates      map[uint64]*UpdateData   // 待处理的批量更新
	updateQueue         []func()                 // 更新队列，用于批量处理
	mutex               sync.RWMutex             // 读写锁
	nextGUID            uint64                   // 下一个GUID
	lastUpdateTime      time.Time                // 上次更新时间
	updateInterval      time.Duration            // 更新间隔
	maxPacketsPerUpdate int                      // 每次更新最大数据包数
	batchSyncManager    *BatchSyncManager        // 批量同步管理器
}

func NewWorld() *World {
	world := &World{
		units:               make(map[uint64]IUnit),
		sessions:            make(map[uint32]*WorldSession),
		pendingUpdates:      make(map[uint64]*UpdateData),
		updateQueue:         make([]func(), 0),
		nextGUID:            1,
		lastUpdateTime:      time.Now(),
		updateInterval:      200 * time.Millisecond, // 200ms更新间隔
		maxPacketsPerUpdate: 150,                    // AzerothCore的限制
	}

	// 初始化批量同步管理器
	world.batchSyncManager = NewBatchSyncManager(world)
	world.batchSyncManager.Start()

	return world
}

// AddUnit 添加单位到世界
func (w *World) AddUnit(unit IUnit) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 如果单位没有GUID，分配一个新的
	if unit.GetGUID() == 0 {
		unit.SetGUID(w.nextGUID)
		w.nextGUID++
	}

	w.units[unit.GetGUID()] = unit
	fmt.Printf("单位 %s 加入世界 (GUID: %d)\n", unit.GetName(), unit.GetGUID())
}

func (w *World) RemoveUnit(guid uint64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if unit, exists := w.units[guid]; exists {
		delete(w.units, guid)
		fmt.Printf("单位 %s 离开世界\n", unit.GetName())
	}
}

func (w *World) GetUnit(guid uint64) IUnit {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.units[guid]
}

// AddSession 添加会话到世界
func (w *World) AddSession(session *WorldSession) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.sessions[session.id] = session
}

// RemoveSession 从世界移除会话
func (w *World) RemoveSession(sessionId uint32) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.sessions, sessionId)
}

// GetSession 获取会话
func (w *World) GetSession(sessionId uint32) *WorldSession {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.sessions[sessionId]
}

// GetSessionCount 获取会话数量
func (w *World) GetSessionCount() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return len(w.sessions)
}

// QueueUpdate 将更新操作加入队列，用于批量处理
func (w *World) QueueUpdate(updateFunc func()) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.updateQueue = append(w.updateQueue, updateFunc)
}

// GetPlayersInRange 获取指定范围内的玩家（选择性更新）
func (w *World) GetPlayersInRange(centerX, centerY, centerZ float32, rangeDist float32) []*WorldSession {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var players []*WorldSession
	for _, session := range w.sessions {
		if session.player != nil {
			player := session.player
			dx := player.GetX() - centerX
			dy := player.GetY() - centerY
			dz := player.GetZ() - centerZ
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

			if distance <= rangeDist {
				players = append(players, session)
			}
		}
	}
	return players
}

// AddBatchUpdate 添加批量更新（AzerothCore风格）
func (w *World) AddBatchUpdate(unit IUnit, sessionId uint32, updateBlock []byte) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	unitGUID := unit.GetGUID()
	if _, exists := w.pendingUpdates[unitGUID]; !exists {
		w.pendingUpdates[unitGUID] = NewUpdateData()
	}

	w.pendingUpdates[unitGUID].AddUpdateBlock(sessionId, updateBlock)
}

// SendBatchUpdates 发送批量更新（AzerothCore风格优化版 + 时序控制）
// 参考 AzerothCore 的 Map::SendObjectUpdates() 实现
func (w *World) SendBatchUpdates() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if len(w.pendingUpdates) == 0 {
		return
	}

	// 🔥 关键优化：按会话分组更新，而不是按单位分组
	// 时间复杂度从 O(U×S×P) 优化为 O(S×U)
	sessionUpdates := make(map[uint32]*UpdateData) // 每个会话的合并更新
	packetsSent := 0
	currentUpdateId := atomic.AddUint32(&globalUpdateId, 1) // 生成批量更新ID

	// 第一步：收集并合并每个会话的所有更新 - O(U×S)
	for unitGUID, updateData := range w.pendingUpdates {
		// 检查单位是否还存在
		if _, exists := w.units[unitGUID]; !exists {
			delete(w.pendingUpdates, unitGUID)
			continue
		}

		// 为每个会话合并更新数据
		for sessionId, blockData := range updateData.blocks {
			if _, exists := sessionUpdates[sessionId]; !exists {
				sessionUpdates[sessionId] = NewUpdateData()
			}
			// 合并到会话的更新数据中
			sessionUpdates[sessionId].AddUpdateBlock(sessionId, blockData)
		}
	}

	// 第二步：为每个会话发送一个合并的数据包 - O(S)
	for sessionId, mergedUpdateData := range sessionUpdates {
		if session, exists := w.sessions[sessionId]; exists && session.IsConnected() {
			// 构建合并的数据包
			packet := mergedUpdateData.BuildPacket(sessionId)
			if packet != nil {
				// 🔥 关键：设置数据包时序信息
				packet.SetUpdateId(currentUpdateId)
				packet.SetPriority(1) // 批量更新使用高优先级

				// 添加压缩支持（AzerothCore 风格）
				if packet.wpos > 100 { // 大于100字节时压缩
					compressedPacket := w.compressPacket(packet)
					if compressedPacket != nil {
						packet = compressedPacket
						packet.SetUpdateId(currentUpdateId) // 压缩后重新设置ID
					}
				}

				// 🔥 关键：使用有序发送
				session.SendPacketOrdered(packet)
				packetsSent++

				// 网络流量控制
				if packetsSent >= w.maxPacketsPerUpdate {
					break
				}
			}
		}
	}

	// 清理已处理的更新
	w.pendingUpdates = make(map[uint64]*UpdateData)

	fmt.Printf("[World] 批量更新优化: %d个会话, %d个数据包 (更新ID: %d)\n",
		len(sessionUpdates), packetsSent, currentUpdateId)

	// 处理更新队列
	queueProcessed := 0
	for len(w.updateQueue) > 0 && queueProcessed < w.maxPacketsPerUpdate {
		updateFunc := w.updateQueue[0]
		w.updateQueue = w.updateQueue[1:]
		updateFunc()
		queueProcessed++
	}
}

// compressPacket 压缩数据包（AzerothCore风格）
func (w *World) compressPacket(packet *WorldPacket) *WorldPacket {
	// 简化的压缩实现
	if packet.wpos < 100 {
		return packet // 小数据包不压缩
	}

	compressedPacket := NewWorldPacket(SMSG_COMPRESSED_UPDATE_OBJECT)
	compressedPacket.WriteUint32(uint32(packet.wpos)) // 原始大小

	// 这里应该使用 zlib 压缩，简化处理
	compressedPacket.data = append(compressedPacket.data, packet.data...)
	compressedPacket.wpos += len(packet.data)

	return compressedPacket
}

// Update 世界更新循环（AzerothCore风格）
func (w *World) Update(diff uint32) {
	currentTime := time.Now()
	elapsed := currentTime.Sub(w.lastUpdateTime)

	// 达到更新间隔时才进行批量更新
	if elapsed >= w.updateInterval {
		// 发送传统的批量更新（使用 AddBatchUpdate 收集的数据）
		w.SendBatchUpdates()
		w.lastUpdateTime = currentTime

		// 打印批量更新统计
		if len(w.pendingUpdates) > 0 {
			fmt.Printf("[World] 发送批量更新: %d个单位有待更新\n", len(w.pendingUpdates))
		}
	}

	// 更新所有会话
	w.mutex.RLock()
	for _, session := range w.sessions {
		if !session.Update(diff) {
			// 会话已断开，标记为需要清理
			continue
		}
	}
	w.mutex.RUnlock()

	// 定期广播状态更新
	w.broadcastPeriodicUpdates(diff)

	// 清理死亡单位
	w.cleanupDeadUnits()
}

// BroadcastToPlayersInRange 向范围内的玩家广播（选择性更新）
func (w *World) BroadcastToPlayersInRange(centerX, centerY, centerZ float32, rangeDist float32, packet *WorldPacket) {
	players := w.GetPlayersInRange(centerX, centerY, centerZ, rangeDist)

	packetsSent := 0
	for _, player := range players {
		if player.IsConnected() {
			player.SendPacket(packet)
			packetsSent++

			// 网络流量控制
			if packetsSent >= w.maxPacketsPerUpdate {
				break
			}
		}
	}
}

// BroadcastSpellStart 批量广播法术开始
func (w *World) BroadcastSpellStart(caster IUnit, spellId uint32, targets []IUnit, castTime time.Duration) {
	// 只向范围内的玩家广播
	casterX, casterY, casterZ := caster.GetPosition()
	players := w.GetPlayersInRange(casterX, casterY, casterZ, 100.0) // 100码范围

	// 构建更新数据
	packet := NewWorldPacket(SMSG_SPELL_START)
	packet.WriteUint32(spellId)
	packet.WriteUint32(uint32(castTime.Milliseconds()))
	packet.WriteUint32(uint32(len(targets)))

	for _, target := range targets {
		if target != nil {
			packet.WriteUint64(target.GetGUID())
		} else {
			packet.WriteUint64(0)
		}
	}

	// 收集目标会话ID
	var sessionTargets []uint32
	for _, player := range players {
		sessionTargets = append(sessionTargets, player.id)
	}

	// 创建批量更新
	update := NewBatchUpdate(caster.GetGUID(), "spell", packet.data, sessionTargets)

	// 法术开始通常需要立即同步
	w.batchSyncManager.QueueImmediateUpdate(update)

	// 获取法术信息用于日志
	spellInfo := GlobalSpellManager.GetSpell(spellId)
	spellName := "未知法术"
	if spellInfo != nil {
		spellName = spellInfo.Name
	}

	fmt.Printf("[批量同步] 法术开始: %s 施放 %s (范围: %d玩家)\n", caster.GetName(), spellName, len(players))
}

// BroadcastSpellGo 批量广播法术生效
func (w *World) BroadcastSpellGo(caster IUnit, spellId uint32, targets []IUnit) {
	// 只向范围内的玩家广播
	casterX, casterY, casterZ := caster.GetPosition()
	players := w.GetPlayersInRange(casterX, casterY, casterZ, 100.0) // 100码范围

	// 构建更新数据
	packet := NewWorldPacket(SMSG_SPELLGO)
	packet.WriteUint32(spellId)
	packet.WriteUint32(uint32(len(targets)))

	for _, target := range targets {
		if target != nil {
			packet.WriteUint64(target.GetGUID())
		} else {
			packet.WriteUint64(0)
		}
	}

	// 收集目标会话ID
	var sessionTargets []uint32
	for _, player := range players {
		sessionTargets = append(sessionTargets, player.id)
	}

	// 创建批量更新
	update := NewBatchUpdate(caster.GetGUID(), "spell", packet.data, sessionTargets)

	// 法术生效通常需要立即同步
	w.batchSyncManager.QueueImmediateUpdate(update)

	spellInfo := GlobalSpellManager.GetSpell(spellId)
	spellName := "未知法术"
	if spellInfo != nil {
		spellName = spellInfo.Name
	}

	fmt.Printf("[批量同步] 法术生效: %s 的 %s 生效 (范围: %d玩家)\n", caster.GetName(), spellName, len(players))
}

// BroadcastHealthUpdate 批量广播血量更新
func (w *World) BroadcastHealthUpdate(unit IUnit, oldHealth, newHealth uint32) {
	// 只向范围内的玩家广播
	unitX, unitY, unitZ := unit.GetPosition()
	players := w.GetPlayersInRange(unitX, unitY, unitZ, 100.0) // 100码范围

	// 构建更新数据
	packet := NewWorldPacket(SMSG_HEALTH_UPDATE)
	packet.WriteUint32(oldHealth)
	packet.WriteUint32(newHealth)
	packet.WriteUint32(unit.GetMaxHealth())

	// 收集目标会话ID
	var targets []uint32
	for _, player := range players {
		targets = append(targets, player.id)
	}

	// 创建批量更新
	update := NewBatchUpdate(unit.GetGUID(), "health", packet.data, targets)

	// 根据血量变化的紧急程度决定同步方式
	healthChangePercent := float32(abs(int32(newHealth-oldHealth))) / float32(unit.GetMaxHealth()) * 100
	if healthChangePercent > 20.0 || newHealth <= unit.GetMaxHealth()/10 { // 血量变化超过20%或血量低于10%
		// 立即同步
		w.batchSyncManager.QueueImmediateUpdate(update)
	} else {
		// 批量同步
		w.batchSyncManager.QueueBatchUpdate(update)
	}

	fmt.Printf("[批量同步] 血量更新: %s %d→%d/%d (范围: %d玩家, 变化: %.1f%%)\n",
		unit.GetName(), oldHealth, newHealth, unit.GetMaxHealth(), len(players), healthChangePercent)
}

// BroadcastPowerUpdate 批量广播能量更新
func (w *World) BroadcastPowerUpdate(unit IUnit, powerType uint8, oldPower, newPower uint32) {
	// 只向范围内的玩家广播
	unitX, unitY, unitZ := unit.GetPosition()
	players := w.GetPlayersInRange(unitX, unitY, unitZ, 100.0) // 100码范围

	// 构建更新数据
	packet := NewWorldPacket(SMSG_POWER_UPDATE)
	packet.WriteUint8(powerType)
	packet.WriteUint32(oldPower)
	packet.WriteUint32(newPower)
	packet.WriteUint32(unit.GetMaxPower(powerType))

	// 收集目标会话ID
	var targets []uint32
	for _, player := range players {
		targets = append(targets, player.id)
	}

	// 创建批量更新
	update := NewBatchUpdate(unit.GetGUID(), "power", packet.data, targets)

	// 能量更新通常使用批量同步
	w.batchSyncManager.QueueBatchUpdate(update)

	powerName := "未知能量"
	switch powerType {
	case POWER_MANA:
		powerName = "法力"
	case POWER_RAGE:
		powerName = "怒气"
	case POWER_ENERGY:
		powerName = "能量"
	}

	fmt.Printf("[批量同步] 能量更新: %s %s %d→%d/%d (范围: %d玩家)\n",
		unit.GetName(), powerName, oldPower, newPower, unit.GetMaxPower(powerType), len(players))
}

// BroadcastAttackerStateUpdate 广播攻击状态更新 - 完整复刻AzerothCore的SMSG_ATTACKERSTATEUPDATE
// 参考 AzerothCore 的 Unit::SendAttackStateUpdate() 实现
func (w *World) BroadcastAttackerStateUpdate(attacker, victim IUnit, damage uint32, hitResult int, schoolMask int) {
	// 只向范围内的玩家广播
	attackerX, attackerY, attackerZ := attacker.GetPosition()
	players := w.GetPlayersInRange(attackerX, attackerY, attackerZ, 100.0) // 100码范围

	if len(players) == 0 {
		return // 没有玩家在范围内
	}

	// 🔥 关键：构建完整的SMSG_ATTACKERSTATEUPDATE数据包
	// 参考 AzerothCore 的 Unit.cpp:6580-6678 实现
	packet := NewWorldPacket(SMSG_ATTACKERSTATEUPDATE)

	// 攻击信息
	packet.WriteUint32(uint32(hitResult))  // HitInfo
	packet.WriteUint64(attacker.GetGUID()) // 攻击者GUID (PackGUID格式)
	packet.WriteUint64(victim.GetGUID())   // 受害者GUID (PackGUID格式)

	// 伤害信息
	packet.WriteUint32(damage) // 总伤害
	overkill := int32(damage) - int32(victim.GetHealth())
	if overkill < 0 {
		overkill = 0
	}
	packet.WriteUint32(uint32(overkill)) // 过量伤害

	// 子伤害数量（通常为1，除非有多种伤害类型）
	packet.WriteUint8(1) // 子伤害计数

	// 子伤害详情
	packet.WriteUint32(damage)             // 伤害值
	packet.WriteUint32(uint32(schoolMask)) // 伤害学派掩码
	packet.WriteUint32(0)                  // 吸收伤害
	packet.WriteUint32(0)                  // 抵抗伤害

	// 受害者状态
	victimState := uint8(0) // VICTIMSTATE_NORMAL
	if victim.GetHealth() <= damage {
		victimState = 1 // VICTIMSTATE_DIES
	}
	packet.WriteUint8(victimState)

	// 额外信息
	packet.WriteUint32(0) // 未知攻击者状态
	packet.WriteUint32(0) // 近战法术ID

	// 根据命中类型添加额外数据
	if hitResult&MELEE_HIT_BLOCK != 0 {
		packet.WriteUint32(0) // 格挡伤害
	}

	if hitResult&0x00000040 != 0 { // HITINFO_RAGE_GAIN
		packet.WriteUint32(0) // 怒气获得
	}

	// 🔥 关键：设置最高优先级和更新ID
	damageUpdateId := atomic.AddUint32(&globalUpdateId, 1)
	packet.SetPriority(0) // 最高优先级 - 伤害信息必须立即同步
	packet.SetUpdateId(damageUpdateId)

	// 🔥 关键：立即同步到所有相关玩家
	// 伤害信息必须立即同步，不能批量处理，确保玩家看到实时的伤害数字
	packetsSent := 0
	for _, player := range players {
		if player.IsConnected() {
			player.SendPacketOrdered(packet) // 使用有序发送
			packetsSent++
		}
	}

	fmt.Printf("[🔥伤害同步] %s 对 %s 造成 %d 伤害 (命中类型: 0x%X, 同步给 %d 玩家, 优先级: 立即, 更新ID: %d)\n",
		attacker.GetName(), victim.GetName(), damage, hitResult, packetsSent, damageUpdateId)

}

// BroadcastUnitUpdate 批量广播单位状态更新
func (w *World) BroadcastUnitUpdate(unit IUnit) {
	w.BroadcastUnitUpdateWithPriority(unit, 2, 0) // 默认普通优先级
}

// BroadcastUnitUpdateWithPriority 带优先级的单位状态更新
func (w *World) BroadcastUnitUpdateWithPriority(unit IUnit, priority uint8, updateId uint32) {
	// 只向范围内的玩家广播
	unitX, unitY, unitZ := unit.GetPosition()
	players := w.GetPlayersInRange(unitX, unitY, unitZ, 100.0) // 100码范围

	if len(players) == 0 {
		return // 没有玩家在范围内
	}

	// 如果没有指定更新ID，生成一个新的
	if updateId == 0 {
		updateId = atomic.AddUint32(&globalUpdateId, 1)
	}

	// 🔥 关键：为每个会话发送有序的更新数据包
	for _, player := range players {
		if player.IsConnected() {
			// 构建更新数据包
			packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
			packet.WriteUint64(unit.GetGUID())
			packet.WriteUint32(unit.GetHealth())
			packet.WriteUint32(unit.GetMaxHealth())
			packet.WriteUint32(unit.GetPower(POWER_MANA))
			packet.WriteUint32(unit.GetMaxPower(POWER_MANA))

			// 🔥 关键：设置时序信息
			packet.SetPriority(priority)
			packet.SetUpdateId(updateId)

			// 使用有序发送
			player.SendPacketOrdered(packet)
		}
	}

	priorityName := "未知"
	switch priority {
	case 0:
		priorityName = "立即"
	case 1:
		priorityName = "高"
	case 2:
		priorityName = "普通"
	case 3:
		priorityName = "低"
	}

	fmt.Printf("[时序同步] 单位状态更新: %s (范围: %d玩家, 优先级: %s, 更新ID: %d)\n",
		unit.GetName(), len(players), priorityName, updateId)
}

// 获取能量类型名称
func getPowerTypeName(powerType uint8) string {
	switch powerType {
	case POWER_MANA:
		return "法力值"
	case POWER_RAGE:
		return "怒气值"
	case POWER_FOCUS:
		return "集中值"
	case POWER_ENERGY:
		return "能量值"
	default:
		return "未知能量"
	}
}

// 定期广播状态更新 - 基于AzerothCore的定期同步机制 + 时序控制
var lastPeriodicUpdate uint32 = 0

const PERIODIC_UPDATE_INTERVAL = 5000 // 5秒

func (w *World) broadcastPeriodicUpdates(diff uint32) {
	lastPeriodicUpdate += diff
	if lastPeriodicUpdate >= PERIODIC_UPDATE_INTERVAL {
		lastPeriodicUpdate = 0

		// 🔥 关键：生成定期更新ID，确保时序正确
		periodicUpdateId := atomic.AddUint32(&globalUpdateId, 1)

		// 广播所有单位的完整状态
		w.mutex.RLock()
		for _, unit := range w.units {
			if unit.IsAlive() {
				w.BroadcastUnitUpdateWithPriority(unit, 3, periodicUpdateId) // 使用低优先级
			}
		}
		w.mutex.RUnlock()

		fmt.Printf("[网络] 定期状态同步完成 (更新ID: %d, 优先级: 低)\n", periodicUpdateId)
	}
}

// ProcessIncomingPackets 处理入站数据包 - 基于AzerothCore的数据包处理机制
func (w *World) ProcessIncomingPackets() {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for _, session := range w.sessions {
		if session.IsConnected() {
			// 每个会话的Update方法会处理其接收队列中的数据包
			// 这里不需要额外处理，因为Update方法已经包含了数据包处理逻辑
		}
	}
}

func (w *World) cleanupDeadUnits() {
	// 简化版：不立即清理死亡单位，让它们保持在世界中用于演示
}

// Shutdown 关闭世界
func (w *World) Shutdown() {
	if w.batchSyncManager != nil {
		w.batchSyncManager.Stop()
	}

	// 关闭所有会话
	w.mutex.Lock()
	for _, session := range w.sessions {
		session.Close()
	}
	w.sessions = make(map[uint32]*WorldSession)
	w.mutex.Unlock()

	fmt.Println("[World] 世界已关闭")
}

// GetBatchSyncManager 获取批量同步管理器
func (w *World) GetBatchSyncManager() *BatchSyncManager {
	return w.batchSyncManager
}

func (w *World) GetAliveUnits() []IUnit {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var aliveUnits []IUnit
	for _, unit := range w.units {
		if unit.IsAlive() {
			aliveUnits = append(aliveUnits, unit)
		}
	}
	return aliveUnits
}

// 法术广播方法 - 基于AzerothCore的法术网络同步

// 光环系统（简化版）
type AuraType int

const (
	AURA_MOD_DAMAGE_DONE = iota
	AURA_MOD_DAMAGE_TAKEN
	AURA_MOD_HEALING_DONE
	AURA_MOD_HEALING_TAKEN
	AURA_MOD_ATTACK_SPEED
	AURA_MOD_CAST_SPEED
	AURA_MOD_STAT
	AURA_PERIODIC_DAMAGE
	AURA_PERIODIC_HEAL
)

type Aura struct {
	id       uint32
	caster   IUnit
	target   IUnit
	duration uint32
	auraType AuraType
	value    int32
}

// 战斗日志系统
type CombatLog struct {
	entries []CombatLogEntry
}

type CombatLogEntry struct {
	timestamp uint32
	eventType string
	source    string
	target    string
	value     uint32
	details   string
}

func NewCombatLog() *CombatLog {
	return &CombatLog{
		entries: make([]CombatLogEntry, 0),
	}
}

func (cl *CombatLog) AddEntry(eventType, source, target string, value uint32, details string) {
	entry := CombatLogEntry{
		timestamp: getCurrentTime(),
		eventType: eventType,
		source:    source,
		target:    target,
		value:     value,
		details:   details,
	}
	cl.entries = append(cl.entries, entry)

	// 打印日志
	fmt.Printf("[CombatLog] %s: %s -> %s (%d) %s\n",
		eventType, source, target, value, details)
}

func getCurrentTime() uint32 {
	// 简化版：返回固定时间戳
	return 12345
}

// 统计系统
type CombatStats struct {
	totalDamageDealt  uint32
	totalDamageTaken  uint32
	totalHealingDone  uint32
	totalHealingTaken uint32
	attacksLanded     uint32
	attacksMissed     uint32
	criticalHits      uint32
}

func NewCombatStats() *CombatStats {
	return &CombatStats{}
}

func (cs *CombatStats) RecordDamageDealt(damage uint32) {
	cs.totalDamageDealt += damage
}

func (cs *CombatStats) RecordDamageTaken(damage uint32) {
	cs.totalDamageTaken += damage
}

func (cs *CombatStats) RecordAttackLanded() {
	cs.attacksLanded++
}

func (cs *CombatStats) RecordAttackMissed() {
	cs.attacksMissed++
}

func (cs *CombatStats) RecordCriticalHit() {
	cs.criticalHits++
}

func (cs *CombatStats) GetHitRate() float32 {
	total := cs.attacksLanded + cs.attacksMissed
	if total == 0 {
		return 0
	}
	return float32(cs.attacksLanded) / float32(total) * 100
}

func (cs *CombatStats) GetCritRate() float32 {
	if cs.attacksLanded == 0 {
		return 0
	}
	return float32(cs.criticalHits) / float32(cs.attacksLanded) * 100
}

func (cs *CombatStats) PrintStats(unitName string) {
	fmt.Printf("\n=== %s 的战斗统计 ===\n", unitName)
	fmt.Printf("造成伤害: %d\n", cs.totalDamageDealt)
	fmt.Printf("承受伤害: %d\n", cs.totalDamageTaken)
	fmt.Printf("攻击命中: %d\n", cs.attacksLanded)
	fmt.Printf("攻击未命中: %d\n", cs.attacksMissed)
	fmt.Printf("暴击次数: %d\n", cs.criticalHits)
	fmt.Printf("命中率: %.1f%%\n", cs.GetHitRate())
	fmt.Printf("暴击率: %.1f%%\n", cs.GetCritRate())
}

// 扩展Unit结构以支持统计
func (u *Unit) initStats() {
	// 这里可以初始化战斗统计
}

// 工具函数
func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func clamp(value, minVal, maxVal uint32) uint32 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// 百分比计算
func calculatePct(base uint32, pct float32) uint32 {
	return uint32(float32(base) * pct / 100.0)
}

// 距离计算
func calculateDistance2D(x1, y1, x2, y2 float32) float32 {
	dx := x1 - x2
	dy := y1 - y2
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

// 角度计算
func calculateAngle(x1, y1, x2, y2 float32) float32 {
	return float32(math.Atan2(float64(y2-y1), float64(x2-x1)))
}

// 随机数工具
func rollChance(chance float32) bool {
	return rand.Float32()*100 < chance
}

func rollDice(sides int) int {
	return rand.Intn(sides) + 1
}

// abs 计算绝对值
func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// 时间工具
func getMSTime() uint32 {
	return uint32(time.Now().UnixNano() / 1000000)
}

// 调试工具
func debugPrint(format string, args ...interface{}) {
	// 可以通过配置开关控制是否打印调试信息
	if true { // DEBUG_MODE
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}
