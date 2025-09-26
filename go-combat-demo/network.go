package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// 操作码定义 - 基于AzerothCore的Opcodes.h
const (
	// 客户端到服务器的操作码 (CMSG)
	CMSG_ATTACKSWING        = 0x141 // 攻击挥舞
	CMSG_ATTACKSTOP         = 0x142 // 停止攻击
	CMSG_SET_SELECTION      = 0x13D // 设置选择目标
	CMSG_CAST_SPELL         = 0x12E // 施放法术
	CMSG_CANCEL_CAST        = 0x12F // 取消施法
	CMSG_CANCEL_CHANNELLING = 0x130 // 取消引导
	CMSG_MOVE_START_FORWARD = 0x0B1 // 开始前进
	CMSG_MOVE_STOP          = 0x0B7 // 停止移动
	CMSG_KEEP_ALIVE         = 0x406 // 保持连接

	// 服务器到客户端的操作码 (SMSG)
	SMSG_ATTACKSTART         = 0x143 // 攻击开始
	SMSG_ATTACKSTOP          = 0x144 // 攻击停止
	SMSG_ATTACKERSTATEUPDATE = 0x14A // 攻击者状态更新
	SMSG_SPELL_START         = 0x131 // 法术开始
	SMSG_SPELLGO             = 0x132 // 法术施放
	SMSG_SPELL_FAILURE       = 0x133 // 法术失败
	SMSG_SPELL_COOLDOWN      = 0x134 // 法术冷却
	SMSG_AURA_UPDATE         = 0x495 // 光环更新
	SMSG_UPDATE_OBJECT       = 0x0A9 // 对象更新
	SMSG_POWER_UPDATE        = 0x480 // 能量更新 - 基于AzerothCore
	SMSG_HEALTH_UPDATE       = 0x481 // 血量更新 - 自定义消息
	SMSG_SPELL_HEAL_LOG      = 0x150 // 治疗日志
	SMSG_SPELL_ENERGIZE_LOG  = 0x151 // 能量恢复日志
)

// 数据包处理类型 - 基于AzerothCore的PacketProcessing
const (
	PROCESS_INPLACE      = 0 // 立即处理
	PROCESS_THREADUNSAFE = 1 // 线程不安全，在主线程处理
	PROCESS_THREADSAFE   = 2 // 线程安全，可在任意线程处理
)

// 会话状态 - 基于AzerothCore的SessionStatus
const (
	STATUS_NEVER     = 0 // 永不处理
	STATUS_UNHANDLED = 1 // 未处理
	STATUS_AUTHED    = 2 // 已认证
	STATUS_LOGGEDIN  = 3 // 已登录
)

// WorldPacket - 基于AzerothCore的WorldPacket
type WorldPacket struct {
	opcode uint16 // 操作码
	data   []byte // 数据
	rpos   int    // 读取位置
	wpos   int    // 写入位置
}

// NewWorldPacket 创建新的数据包
func NewWorldPacket(opcode uint16) *WorldPacket {
	return &WorldPacket{
		opcode: opcode,
		data:   make([]byte, 0, 1024),
		rpos:   0,
		wpos:   0,
	}
}

// GetOpcode 获取操作码
func (wp *WorldPacket) GetOpcode() uint16 {
	return wp.opcode
}

// WriteUint32 写入32位整数
func (wp *WorldPacket) WriteUint32(val uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, val)
	wp.data = append(wp.data, buf...)
	wp.wpos += 4
}

// WriteUint64 写入64位整数
func (wp *WorldPacket) WriteUint64(val uint64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, val)
	wp.data = append(wp.data, buf...)
	wp.wpos += 8
}

// WriteString 写入字符串
func (wp *WorldPacket) WriteString(str string) {
	wp.data = append(wp.data, []byte(str)...)
	wp.data = append(wp.data, 0) // null terminator
	wp.wpos += len(str) + 1
}

// ReadUint32 读取32位整数
func (wp *WorldPacket) ReadUint32() uint32 {
	if wp.rpos+4 > len(wp.data) {
		return 0
	}
	val := binary.LittleEndian.Uint32(wp.data[wp.rpos:])
	wp.rpos += 4
	return val
}

// ReadUint64 读取64位整数
func (wp *WorldPacket) ReadUint64() uint64 {
	if wp.rpos+8 > len(wp.data) {
		return 0
	}
	val := binary.LittleEndian.Uint64(wp.data[wp.rpos:])
	wp.rpos += 8
	return val
}

// GetData 获取数据
func (wp *WorldPacket) GetData() []byte {
	return wp.data
}

// Size 获取数据大小
func (wp *WorldPacket) Size() int {
	return len(wp.data)
}

// OpcodeHandler 操作码处理器接口
type OpcodeHandler interface {
	Handle(session *WorldSession, packet *WorldPacket)
	GetName() string
	GetStatus() int
	GetProcessing() int
}

// ClientOpcodeHandler 客户端操作码处理器
type ClientOpcodeHandler struct {
	name       string
	status     int
	processing int
	handler    func(*WorldSession, *WorldPacket)
}

func (h *ClientOpcodeHandler) Handle(session *WorldSession, packet *WorldPacket) {
	h.handler(session, packet)
}

func (h *ClientOpcodeHandler) GetName() string {
	return h.name
}

func (h *ClientOpcodeHandler) GetStatus() int {
	return h.status
}

func (h *ClientOpcodeHandler) GetProcessing() int {
	return h.processing
}

// OpcodeTable 操作码表 - 基于AzerothCore的OpcodeTable
type OpcodeTable struct {
	handlers map[uint16]OpcodeHandler
	mutex    sync.RWMutex
}

// NewOpcodeTable 创建操作码表
func NewOpcodeTable() *OpcodeTable {
	table := &OpcodeTable{
		handlers: make(map[uint16]OpcodeHandler),
	}
	table.Initialize()
	return table
}

// Initialize 初始化操作码表
func (ot *OpcodeTable) Initialize() {
	// 注册攻击相关操作码
	ot.RegisterHandler(CMSG_ATTACKSWING, &ClientOpcodeHandler{
		name:       "CMSG_ATTACKSWING",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleAttackSwingOpcode,
	})

	ot.RegisterHandler(CMSG_ATTACKSTOP, &ClientOpcodeHandler{
		name:       "CMSG_ATTACKSTOP",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleAttackStopOpcode,
	})

	ot.RegisterHandler(CMSG_SET_SELECTION, &ClientOpcodeHandler{
		name:       "CMSG_SET_SELECTION",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleSetSelectionOpcode,
	})

	ot.RegisterHandler(CMSG_CAST_SPELL, &ClientOpcodeHandler{
		name:       "CMSG_CAST_SPELL",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleCastSpellOpcode,
	})

	ot.RegisterHandler(CMSG_CANCEL_CAST, &ClientOpcodeHandler{
		name:       "CMSG_CANCEL_CAST",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleCancelCastOpcode,
	})

	ot.RegisterHandler(CMSG_CANCEL_CHANNELLING, &ClientOpcodeHandler{
		name:       "CMSG_CANCEL_CHANNELLING",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleCancelChannellingOpcode,
	})

	ot.RegisterHandler(CMSG_KEEP_ALIVE, &ClientOpcodeHandler{
		name:       "CMSG_KEEP_ALIVE",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_INPLACE,
		handler:    (*WorldSession).HandleKeepAliveOpcode,
	})
}

// RegisterHandler 注册处理器
func (ot *OpcodeTable) RegisterHandler(opcode uint16, handler OpcodeHandler) {
	ot.mutex.Lock()
	defer ot.mutex.Unlock()
	ot.handlers[opcode] = handler
}

// GetHandler 获取处理器
func (ot *OpcodeTable) GetHandler(opcode uint16) OpcodeHandler {
	ot.mutex.RLock()
	defer ot.mutex.RUnlock()
	return ot.handlers[opcode]
}

// WorldSocket - 基于AzerothCore的WorldSocket
type WorldSocket struct {
	conn         net.Conn
	session      *WorldSession
	sendQueue    chan *WorldPacket
	closed       bool
	mutex        sync.Mutex
	lastPingTime time.Time
}

// NewWorldSocket 创建世界套接字
func NewWorldSocket(conn net.Conn) *WorldSocket {
	socket := &WorldSocket{
		conn:         conn,
		sendQueue:    make(chan *WorldPacket, 100),
		closed:       false,
		lastPingTime: time.Now(),
	}

	go socket.readLoop()
	go socket.writeLoop()

	return socket
}

// SetSession 设置会话
func (ws *WorldSocket) SetSession(session *WorldSession) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.session = session
}

// SendPacket 发送数据包
func (ws *WorldSocket) SendPacket(packet *WorldPacket) {
	if ws.closed {
		return
	}

	select {
	case ws.sendQueue <- packet:
	default:
		fmt.Printf("发送队列已满，丢弃数据包: %d\n", packet.GetOpcode())
	}
}

// QueuePacket 队列数据包
// QueuePacket 将数据包加入WorldSession的接收队列 - 基于AzerothCore的逻辑
func (ws *WorldSocket) QueuePacket(packet *WorldPacket) {
	if ws.closed || ws.session == nil {
		return
	}

	// 将数据包加入WorldSession的队列，而不是直接处理
	ws.session.QueuePacket(packet)
}

// Close 关闭套接字
func (ws *WorldSocket) Close() {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if !ws.closed {
		ws.closed = true
		ws.conn.Close()
		close(ws.sendQueue)
	}
}

// readLoop 读取循环
func (ws *WorldSocket) readLoop() {
	defer ws.Close()

	for !ws.closed {
		// 简化的数据包读取逻辑
		header := make([]byte, 6) // 2字节大小 + 4字节操作码
		_, err := ws.conn.Read(header)
		if err != nil {
			fmt.Printf("读取数据包头失败: %v\n", err)
			return
		}

		size := binary.LittleEndian.Uint16(header[0:2])
		opcode := binary.LittleEndian.Uint16(header[2:4])

		data := make([]byte, size-4) // 减去操作码大小
		if size > 4 {
			_, err = ws.conn.Read(data)
			if err != nil {
				fmt.Printf("读取数据包数据失败: %v\n", err)
				return
			}
		}

		packet := &WorldPacket{
			opcode: opcode,
			data:   data,
			rpos:   0,
			wpos:   len(data),
		}

		ws.QueuePacket(packet)
	}
}

// writeLoop 写入循环
func (ws *WorldSocket) writeLoop() {
	defer ws.Close()

	for packet := range ws.sendQueue {
		if ws.closed {
			return
		}

		// 构建数据包头
		size := uint16(len(packet.data) + 4) // 数据大小 + 操作码大小
		header := make([]byte, 6)
		binary.LittleEndian.PutUint16(header[0:2], size)
		binary.LittleEndian.PutUint16(header[2:4], packet.opcode)

		// 发送头部
		_, err := ws.conn.Write(header)
		if err != nil {
			fmt.Printf("发送数据包头失败: %v\n", err)
			return
		}

		// 发送数据
		if len(packet.data) > 0 {
			_, err = ws.conn.Write(packet.data)
			if err != nil {
				fmt.Printf("发送数据包数据失败: %v\n", err)
				return
			}
		}
	}
}

// IsOpen 检查套接字是否开放 - 基于AzerothCore的WorldSocket::IsOpen
func (ws *WorldSocket) IsOpen() bool {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return !ws.closed
}

// WorldSession - 基于AzerothCore的WorldSession
type WorldSession struct {
	id          uint32
	accountName string
	player      *Player
	socket      *WorldSocket
	opcodeTable *OpcodeTable
	lastUpdate  time.Time
	timeoutTime time.Time
	mutex       sync.RWMutex
	world       *World
	_recvQueue  chan *WorldPacket // 接收数据包队列，基于AzerothCore的_recvQueue
}

// NewWorldSession 创建世界会话
func NewWorldSession(id uint32, accountName string, socket *WorldSocket, world *World) *WorldSession {
	session := &WorldSession{
		id:          id,
		accountName: accountName,
		socket:      socket,
		opcodeTable: NewOpcodeTable(),
		lastUpdate:  time.Now(),
		timeoutTime: time.Now().Add(60 * time.Second), // 60秒超时
		world:       world,
		_recvQueue:  make(chan *WorldPacket, 200), // 基于AzerothCore的接收队列
	}

	socket.SetSession(session)
	return session
}

// GetPlayer 获取玩家
func (ws *WorldSession) GetPlayer() *Player {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return ws.player
}

// SetPlayer 设置玩家
func (ws *WorldSession) SetPlayer(player *Player) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.player = player
}

// SendPacket 发送数据包
func (ws *WorldSession) SendPacket(packet *WorldPacket) {
	if ws.socket != nil {
		ws.socket.SendPacket(packet)
	}
}

// Update 更新会话 - 基于AzerothCore的WorldSession::Update
func (ws *WorldSession) Update(diff uint32) bool {
	// 检查超时
	if time.Now().After(ws.timeoutTime) {
		fmt.Printf("会话 %d 超时，断开连接\n", ws.id)
		return false
	}

	// 检查连接状态
	if !ws.IsConnected() {
		return false
	}

	// 处理接收队列中的数据包 - 基于AzerothCore的逻辑
	processedPackets := 0
	const MAX_PROCESSED_PACKETS = 150 // 基于AzerothCore的限制

	for processedPackets < MAX_PROCESSED_PACKETS {
		select {
		case packet := <-ws._recvQueue:
			if packet == nil {
				return false
			}
			ws.handlePacket(packet)
			processedPackets++
		default:
			break // 没有更多数据包
		}
	}

	ws.lastUpdate = time.Now()
	return true
}

// processPackets 处理数据包 - 已废弃，使用ProcessIncomingPackets()替代
// 保留此函数以防其他地方有调用，但内部逻辑已移除避免重复处理
func (ws *WorldSession) processPackets() {
	// 此函数已废弃，数据包处理统一由ProcessIncomingPackets()完成
	// 避免重复处理数据包的问题
}

// handlePacket 处理单个数据包
func (ws *WorldSession) handlePacket(packet *WorldPacket) {
	handler := ws.opcodeTable.GetHandler(packet.GetOpcode())
	if handler == nil {
		fmt.Printf("未知操作码: 0x%X\n", packet.GetOpcode())
		return
	}

	// 检查会话状态
	if handler.GetStatus() == STATUS_LOGGEDIN && ws.player == nil {
		fmt.Printf("玩家未登录，忽略操作码: %s\n", handler.GetName())
		return
	}

	fmt.Printf("处理操作码: %s\n", handler.GetName())
	handler.Handle(ws, packet)
}

// ResetTimeOutTime 重置超时时间
func (ws *WorldSession) ResetTimeOutTime(fromPing bool) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if fromPing {
		ws.timeoutTime = time.Now().Add(60 * time.Second)
	} else {
		ws.timeoutTime = time.Now().Add(30 * time.Second)
	}
}

// GetPlayerInfo 获取玩家信息
func (ws *WorldSession) GetPlayerInfo() string {
	if ws.player != nil {
		return fmt.Sprintf("%s (ID: %d)", ws.player.GetName(), ws.id)
	}
	return fmt.Sprintf("Account: %s (ID: %d)", ws.accountName, ws.id)
}

// Close 关闭会话
func (ws *WorldSession) Close() {
	if ws.socket != nil {
		ws.socket.Close()
	}

	// 关闭接收队列
	if ws._recvQueue != nil {
		close(ws._recvQueue)
		ws._recvQueue = nil
	}
}

// IsConnected 检查连接是否有效
func (ws *WorldSession) IsConnected() bool {
	if ws.socket == nil {
		return false
	}

	// 检查套接字是否关闭
	ws.socket.mutex.Lock()
	defer ws.socket.mutex.Unlock()
	return !ws.socket.closed
}

// QueuePacket 将数据包加入接收队列 - 基于AzerothCore的WorldSession::QueuePacket
func (ws *WorldSession) QueuePacket(packet *WorldPacket) {
	if ws._recvQueue == nil {
		return
	}

	select {
	case ws._recvQueue <- packet:
		// 成功加入队列
	default:
		fmt.Printf("会话 %d 接收队列已满，丢弃数据包: 0x%X\n", ws.id, packet.GetOpcode())
	}
}

// === 数据包处理器实现 ===

// HandleAttackSwingOpcode 处理攻击挥舞操作码
func (ws *WorldSession) HandleAttackSwingOpcode(packet *WorldPacket) {
	targetGuid := packet.ReadUint64()

	fmt.Printf("玩家 %s 攻击目标 GUID: %d\n", ws.GetPlayerInfo(), targetGuid)

	player := ws.GetPlayer()
	if player == nil {
		return
	}

	// 查找目标
	target := ws.world.GetUnitByGUID(targetGuid)
	if target == nil {
		// 发送攻击停止
		ws.SendAttackStop(nil)
		return
	}

	// 验证攻击目标
	if !player.IsValidAttackTarget(target) {
		ws.SendAttackStop(target)
		return
	}

	// 开始攻击
	player.Attack(target)

	// 发送攻击开始确认
	ws.SendAttackStart(player, target)
}

// HandleAttackStopOpcode 处理停止攻击操作码
func (ws *WorldSession) HandleAttackStopOpcode(packet *WorldPacket) {
	fmt.Printf("玩家 %s 停止攻击\n", ws.GetPlayerInfo())

	player := ws.GetPlayer()
	if player != nil {
		player.AttackStop()
	}
}

// HandleSetSelectionOpcode 处理设置选择目标操作码
func (ws *WorldSession) HandleSetSelectionOpcode(packet *WorldPacket) {
	targetGuid := packet.ReadUint64()

	fmt.Printf("玩家 %s 选择目标 GUID: %d\n", ws.GetPlayerInfo(), targetGuid)

	player := ws.GetPlayer()
	if player != nil {
		target := ws.world.GetUnitByGUID(targetGuid)
		player.SetTarget(target)
	}
}

// HandleCastSpellOpcode 处理施放法术操作码 - 基于AzerothCore的WorldSession::HandleCastSpellOpcode
func (ws *WorldSession) HandleCastSpellOpcode(packet *WorldPacket) {
	spellId := packet.ReadUint32()
	targetGuid := packet.ReadUint64()

	player := ws.GetPlayer()
	if player == nil {
		return
	}

	// 获取法术信息
	spellInfo := GlobalSpellManager.GetSpell(spellId)
	if spellInfo == nil {
		fmt.Printf("玩家 %s 尝试施放未知法术 %d\n", ws.GetPlayerInfo(), spellId)
		ws.SendSpellFailure(player, spellId, "未知法术")
		return
	}

	fmt.Printf("玩家 %s 施放法术 %s (ID: %d)，目标 GUID: %d\n",
		ws.GetPlayerInfo(), spellInfo.Name, spellId, targetGuid)

	// 确定目标
	var target IUnit
	if targetGuid == 0 || targetGuid == player.GetGUID() {
		target = player // 自己作为目标
	} else {
		target = ws.world.GetUnitByGUID(targetGuid)
		if target == nil {
			fmt.Printf("找不到目标 GUID: %d\n", targetGuid)
			ws.SendSpellFailure(player, spellId, "无效目标")
			return
		}
	}

	// 施放法术
	player.CastSpell(target, spellId)
}

// HandleCancelCastOpcode 处理取消施法操作码
func (ws *WorldSession) HandleCancelCastOpcode(packet *WorldPacket) {
	spellId := packet.ReadUint32()

	fmt.Printf("玩家 %s 取消施法 %d\n", ws.GetPlayerInfo(), spellId)

	player := ws.GetPlayer()
	if player != nil {
		player.InterruptSpell(CURRENT_GENERIC_SPELL)
	}
}

// HandleCancelChannellingOpcode 处理取消引导操作码
func (ws *WorldSession) HandleCancelChannellingOpcode(packet *WorldPacket) {
	fmt.Printf("玩家 %s 取消引导\n", ws.GetPlayerInfo())

	player := ws.GetPlayer()
	if player != nil {
		player.InterruptSpell(CURRENT_CHANNELED_SPELL)
	}
}

// HandleKeepAliveOpcode 处理保持连接操作码
func (ws *WorldSession) HandleKeepAliveOpcode(packet *WorldPacket) {
	ws.ResetTimeOutTime(true)
}

// === 服务器数据包发送方法 ===

// SendAttackStart 发送攻击开始
func (ws *WorldSession) SendAttackStart(attacker, victim IUnit) {
	packet := NewWorldPacket(SMSG_ATTACKSTART)
	packet.WriteUint64(attacker.GetGUID())
	packet.WriteUint64(victim.GetGUID())
	ws.SendPacket(packet)
}

// SendAttackStop 发送攻击停止
func (ws *WorldSession) SendAttackStop(victim IUnit) {
	packet := NewWorldPacket(SMSG_ATTACKSTOP)
	if victim != nil {
		packet.WriteUint64(victim.GetGUID())
	} else {
		packet.WriteUint64(0)
	}
	ws.SendPacket(packet)
}

// SendAttackerStateUpdate 发送攻击者状态更新
func (ws *WorldSession) SendAttackerStateUpdate(attacker, victim IUnit, damage uint32, hitResult int) {
	packet := NewWorldPacket(SMSG_ATTACKERSTATEUPDATE)
	packet.WriteUint32(uint32(hitResult))
	packet.WriteUint64(attacker.GetGUID())
	packet.WriteUint64(victim.GetGUID())
	packet.WriteUint32(damage)
	ws.SendPacket(packet)
}

// SendSpellGo 发送法术施放
func (ws *WorldSession) SendSpellGo(caster IUnit, spellId uint32, targets []IUnit) {
	packet := NewWorldPacket(SMSG_SPELLGO)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint32(spellId)
	packet.WriteUint32(uint32(len(targets)))
	for _, target := range targets {
		packet.WriteUint64(target.GetGUID())
	}
	ws.SendPacket(packet)
}

// SendSpellStart 发送法术开始
func (ws *WorldSession) SendSpellStart(caster IUnit, spellId uint32, targets []IUnit, castTime time.Duration) {
	packet := NewWorldPacket(SMSG_SPELL_START)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint32(spellId)
	packet.WriteUint32(uint32(castTime.Milliseconds()))
	packet.WriteUint32(uint32(len(targets)))
	for _, target := range targets {
		packet.WriteUint64(target.GetGUID())
	}
	ws.SendPacket(packet)
}

// SendSpellFailure 发送法术失败
func (ws *WorldSession) SendSpellFailure(caster IUnit, spellId uint32, reason string) {
	packet := NewWorldPacket(SMSG_SPELL_FAILURE)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint32(spellId)
	packet.WriteString(reason)
	ws.SendPacket(packet)
}

// SendSpellCooldown 发送法术冷却
func (ws *WorldSession) SendSpellCooldown(caster IUnit, spellId uint32, cooldown time.Duration) {
	packet := NewWorldPacket(SMSG_SPELL_COOLDOWN)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint32(spellId)
	packet.WriteUint32(uint32(cooldown.Milliseconds()))
	ws.SendPacket(packet)
}

// SendSpellHealLog 发送治疗日志
func (ws *WorldSession) SendSpellHealLog(caster, target IUnit, spellId, healing uint32) {
	packet := NewWorldPacket(SMSG_SPELL_HEAL_LOG)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint64(target.GetGUID())
	packet.WriteUint32(spellId)
	packet.WriteUint32(healing)
	ws.SendPacket(packet)
}

// SendSpellEnergizeLog 发送能量恢复日志
func (ws *WorldSession) SendSpellEnergizeLog(caster, target IUnit, spellId, amount uint32, powerType int) {
	packet := NewWorldPacket(SMSG_SPELL_ENERGIZE_LOG)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint64(target.GetGUID())
	packet.WriteUint32(spellId)
	packet.WriteUint32(amount)
	packet.WriteUint32(uint32(powerType))
	ws.SendPacket(packet)
}
