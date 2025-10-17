package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sync"
	"sync/atomic"
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
	CMSG_DAMAGE_TAKEN       = 0x200 // 自定义：客户端报告受到伤害

	// 服务器到客户端的操作码 (SMSG)
	SMSG_ATTACKSTART              = 0x143 // 攻击开始
	SMSG_ATTACKSTOP               = 0x144 // 攻击停止
	SMSG_ATTACKERSTATEUPDATE      = 0x14A // 攻击者状态更新
	SMSG_SPELL_START              = 0x131 // 法术开始
	SMSG_SPELLGO                  = 0x132 // 法术施放
	SMSG_SPELL_FAILURE            = 0x133 // 法术失败
	SMSG_SPELL_COOLDOWN           = 0x134 // 法术冷却
	SMSG_AURA_UPDATE              = 0x495 // 光环更新
	SMSG_UPDATE_OBJECT            = 0x0A9 // 对象更新
	SMSG_POWER_UPDATE             = 0x480 // 能量更新 - 基于AzerothCore
	SMSG_HEALTH_UPDATE            = 0x481 // 血量更新 - 自定义消息
	SMSG_SPELL_HEAL_LOG           = 0x150 // 治疗日志
	SMSG_SPELL_ENERGIZE_LOG       = 0x151 // 能量恢复日志
	SMSG_COMPRESSED_UPDATE_OBJECT = 0x1F6 // 压缩的对象更新
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

// WorldPacket - 基于AzerothCore的WorldPacket，增加时序控制
type WorldPacket struct {
	opcode    uint16    // 操作码
	data      []byte    // 数据
	rpos      int       // 读取位置
	wpos      int       // 写入位置
	sequence  uint32    // 序列号 - 确保数据包顺序
	timestamp time.Time // 时间戳 - 用于时序验证
	priority  uint8     // 优先级 - 0=立即, 1=高, 2=普通, 3=低
	updateId  uint32    // 更新ID - 用于版本控制
}

// NewWorldPacket 创建新的数据包
// 全局序列号生成器
var globalSequence uint32 = 0
var globalUpdateId uint32 = 0

func NewWorldPacket(opcode uint16) *WorldPacket {
	return &WorldPacket{
		opcode:    opcode,
		data:      make([]byte, 0, 1024),
		rpos:      0,
		wpos:      0,
		sequence:  atomic.AddUint32(&globalSequence, 1),
		timestamp: time.Now(),
		priority:  2, // 默认普通优先级
		updateId:  atomic.AddUint32(&globalUpdateId, 1),
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

// WriteUint8 写入8位整数
func (wp *WorldPacket) WriteUint8(val uint8) {
	wp.data = append(wp.data, byte(val))
	wp.wpos += 1
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

// WriteFloat32 写入32位浮点数
func (wp *WorldPacket) WriteFloat32(val float32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(val))
	wp.data = append(wp.data, buf...)
	wp.wpos += 4
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

// ReadFloat32 读取32位浮点数
func (wp *WorldPacket) ReadFloat32() float32 {
	if wp.rpos+4 > len(wp.data) {
		return 0.0
	}
	bits := binary.LittleEndian.Uint32(wp.data[wp.rpos:])
	wp.rpos += 4
	return math.Float32frombits(bits)
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

// 🔥 关键：数据包时序控制方法 - 基于AzerothCore的时序机制

// SetPriority 设置数据包优先级
func (wp *WorldPacket) SetPriority(priority uint8) {
	wp.priority = priority
}

// SetUpdateId 设置更新ID（用于版本控制）
func (wp *WorldPacket) SetUpdateId(updateId uint32) {
	wp.updateId = updateId
}

// GetSequence 获取序列号
func (wp *WorldPacket) GetSequence() uint32 {
	return wp.sequence
}

// GetTimestamp 获取时间戳
func (wp *WorldPacket) GetTimestamp() time.Time {
	return wp.timestamp
}

// GetPriority 获取优先级
func (wp *WorldPacket) GetPriority() uint8 {
	return wp.priority
}

// GetUpdateId 获取更新ID
func (wp *WorldPacket) GetUpdateId() uint32 {
	return wp.updateId
}

// IsNewerThan 检查是否比另一个数据包更新
func (wp *WorldPacket) IsNewerThan(other *WorldPacket) bool {
	// 首先比较更新ID
	if wp.updateId != other.updateId {
		return wp.updateId > other.updateId
	}
	// 然后比较时间戳
	return wp.timestamp.After(other.timestamp)
}

// ShouldOverride 检查是否应该覆盖另一个数据包
func (wp *WorldPacket) ShouldOverride(other *WorldPacket) bool {
	// 相同操作码才能覆盖
	if wp.opcode != other.opcode {
		return false
	}

	// 高优先级可以覆盖低优先级
	if wp.priority < other.priority {
		return true
	}

	// 相同优先级时，比较更新ID和时间戳
	if wp.priority == other.priority {
		return wp.IsNewerThan(other)
	}

	return false
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

	ot.RegisterHandler(CMSG_DAMAGE_TAKEN, &ClientOpcodeHandler{
		name:       "CMSG_DAMAGE_TAKEN",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleDamageTakenOpcode,
	})

	// 注册移动相关操作码 - 基于AzerothCore的移动系统
	ot.RegisterHandler(CMSG_MOVE_START_FORWARD, &ClientOpcodeHandler{
		name:       "CMSG_MOVE_START_FORWARD",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleMoveStartForwardOpcode,
	})

	ot.RegisterHandler(CMSG_MOVE_STOP, &ClientOpcodeHandler{
		name:       "CMSG_MOVE_STOP",
		status:     STATUS_LOGGEDIN,
		processing: PROCESS_THREADSAFE,
		handler:    (*WorldSession).HandleMoveStopOpcode,
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

// SendPacket 发送数据包（单个广播）
// 注意：此方法将数据包加入发送队列，不会立即发送
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

// BatchSendPackets 批量发送数据包（批量广播）
// 注意：此方法用于批量发送多个数据包，减少网络开销
func (ws *WorldSocket) BatchSendPackets(packets []*WorldPacket) {
	if ws.closed {
		return
	}

	for _, packet := range packets {
		select {
		case ws.sendQueue <- packet:
		default:
			fmt.Printf("发送队列已满，丢弃数据包: %d\n", packet.GetOpcode())
		}
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

// QueueReceivedPacket 将服务器发送的数据包加入客户端接收队列
func (ws *WorldSocket) QueueReceivedPacket(packet *WorldPacket) {
	if ws.closed || ws.session == nil {
		return
	}

	// 将服务器发送的数据包加入客户端接收队列
	ws.session.QueueReceivedPacket(packet)
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
	player      IUnit
	socket      *WorldSocket

	opcodeTable *OpcodeTable
	lastUpdate  time.Time
	timeoutTime time.Time
	mutex       sync.RWMutex
	world       *World
	_recvQueue  chan *WorldPacket // 接收数据包队列，基于AzerothCore的_recvQueue
	// 客户端接收到的数据包队列（用于客户端处理）
	_receivedQueue chan *WorldPacket

	// 🔥 关键：数据包时序控制 - 基于AzerothCore的时序机制
	lastSequence     uint32                  // 最后处理的序列号
	pendingPackets   map[uint32]*WorldPacket // 待排序的数据包
	lastUpdateStates map[uint16]uint32       // 每种操作码的最后更新ID
	packetBuffer     []*WorldPacket          // 数据包缓冲区
	sortMutex        sync.Mutex              // 排序锁
}

// NewWorldSession 创建世界会话
func NewWorldSession(id uint32, accountName string, socket *WorldSocket, world *World) *WorldSession {
	session := &WorldSession{
		id:             id,
		accountName:    accountName,
		socket:         socket,
		opcodeTable:    NewOpcodeTable(),
		lastUpdate:     time.Now(),
		timeoutTime:    time.Now().Add(60 * time.Second), // 60秒超时
		world:          world,
		_recvQueue:     make(chan *WorldPacket, 200), // 基于AzerothCore的接收队列
		_receivedQueue: make(chan *WorldPacket, 100), // 客户端接收队列

		// 🔥 关键：初始化时序控制字段
		lastSequence:     0,
		pendingPackets:   make(map[uint32]*WorldPacket),
		lastUpdateStates: make(map[uint16]uint32),
		packetBuffer:     make([]*WorldPacket, 0, 50),
	}

	if socket != nil {
		socket.SetSession(session)
	}
	return session

}

// GetPlayer 获取玩家
func (ws *WorldSession) GetPlayer() IUnit {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return ws.player
}

// SetPlayer 设置玩家
func (ws *WorldSession) SetPlayer(player IUnit) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.player = player
}

// SendPacket 发送数据包
func (ws *WorldSession) SendPacket(packet *WorldPacket) {
	if ws.socket != nil {
		ws.socket.SendPacket(packet)
		// 同时将数据包加入客户端接收队列（模拟客户端接收）
		ws.socket.QueueReceivedPacket(packet)
	}
}

// Update 更新会话 - 基于AzerothCore的WorldSession::Update
// 注意：此方法实现了批量处理接收队列中的数据包，优化性能
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

// 🔥 关键：数据包时序控制方法 - 基于AzerothCore的时序机制

// SendPacketOrdered 发送有序数据包
func (ws *WorldSession) SendPacketOrdered(packet *WorldPacket) {
	ws.sortMutex.Lock()
	defer ws.sortMutex.Unlock()

	// 检查是否应该覆盖现有的数据包
	if ws.shouldOverridePacket(packet) {
		ws.removeOldPackets(packet.opcode, packet.updateId)
	}

	// 更新最后的更新状态
	ws.lastUpdateStates[packet.opcode] = packet.updateId

	// 发送数据包
	if ws.socket != nil {
		ws.socket.SendPacket(packet)
		ws.socket.QueueReceivedPacket(packet)
	}
}

// shouldOverridePacket 检查是否应该覆盖现有数据包
func (ws *WorldSession) shouldOverridePacket(newPacket *WorldPacket) bool {
	lastUpdateId, exists := ws.lastUpdateStates[newPacket.opcode]
	if !exists {
		return false
	}

	// 如果新数据包的更新ID更大，则应该覆盖
	return newPacket.updateId > lastUpdateId
}

// removeOldPackets 移除旧的数据包
func (ws *WorldSession) removeOldPackets(opcode uint16, newUpdateId uint32) {
	// 从缓冲区中移除旧的相同类型数据包
	filteredBuffer := make([]*WorldPacket, 0, len(ws.packetBuffer))
	for _, packet := range ws.packetBuffer {
		if packet.opcode != opcode || packet.updateId >= newUpdateId {
			filteredBuffer = append(filteredBuffer, packet)
		}
	}
	ws.packetBuffer = filteredBuffer
}

// SortAndSendPackets 排序并发送数据包
func (ws *WorldSession) SortAndSendPackets(packets []*WorldPacket) {
	if len(packets) == 0 {
		return
	}

	ws.sortMutex.Lock()
	defer ws.sortMutex.Unlock()

	// 按优先级和时间戳排序
	ws.sortPacketsByPriority(packets)

	// 发送排序后的数据包
	for _, packet := range packets {
		if ws.shouldSendPacket(packet) {
			if ws.socket != nil {
				ws.socket.SendPacket(packet)
				ws.socket.QueueReceivedPacket(packet)
			}
			ws.lastUpdateStates[packet.opcode] = packet.updateId
		}
	}
}

// sortPacketsByPriority 按优先级排序数据包
func (ws *WorldSession) sortPacketsByPriority(packets []*WorldPacket) {
	// 简单的冒泡排序，按优先级和时间戳排序
	n := len(packets)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if ws.shouldSwapPackets(packets[j], packets[j+1]) {
				packets[j], packets[j+1] = packets[j+1], packets[j]
			}
		}
	}
}

// shouldSwapPackets 检查是否应该交换两个数据包的顺序
func (ws *WorldSession) shouldSwapPackets(a, b *WorldPacket) bool {
	// 优先级低的数字表示高优先级
	if a.priority != b.priority {
		return a.priority > b.priority
	}

	// 相同优先级时，按时间戳排序
	return a.timestamp.After(b.timestamp)
}

// shouldSendPacket 检查是否应该发送数据包
func (ws *WorldSession) shouldSendPacket(packet *WorldPacket) bool {
	lastUpdateId, exists := ws.lastUpdateStates[packet.opcode]
	if !exists {
		return true
	}

	// 只发送更新的数据包
	return packet.updateId > lastUpdateId
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

	// 关闭客户端接收队列
	if ws._receivedQueue != nil {
		close(ws._receivedQueue)
		ws._receivedQueue = nil
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
	target := ws.world.GetUnit(targetGuid)
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

	// 添加批量更新 - 基于AzerothCore的攻击状态同步
	if unit, ok := player.(*Unit); ok {
		unit.AddBatchUpdateForFullState() // 攻击状态变化需要完整状态更新
	}
}

// HandleAttackStopOpcode 处理停止攻击操作码
func (ws *WorldSession) HandleAttackStopOpcode(packet *WorldPacket) {
	fmt.Printf("玩家 %s 停止攻击\n", ws.GetPlayerInfo())

	player := ws.GetPlayer()
	if player != nil {
		player.AttackStop()

		// 添加批量更新 - 基于AzerothCore的攻击状态同步
		if unit, ok := player.(*Unit); ok {
			unit.AddBatchUpdateForFullState() // 停止攻击状态变化需要完整状态更新
		}
	}
}

// HandleSetSelectionOpcode 处理设置选择目标操作码
func (ws *WorldSession) HandleSetSelectionOpcode(packet *WorldPacket) {
	targetGuid := packet.ReadUint64()

	fmt.Printf("玩家 %s 选择目标 GUID: %d\n", ws.GetPlayerInfo(), targetGuid)

	player := ws.GetPlayer()
	if player != nil {
		target := ws.world.GetUnit(targetGuid)
		player.SetTarget(target)

		// 添加批量更新 - 基于AzerothCore的目标选择同步
		if unit, ok := player.(*Unit); ok {
			unit.AddBatchUpdateForFullState() // 目标变化需要完整状态更新
		}
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
		target = ws.world.GetUnit(targetGuid)
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
		if unit, ok := player.(*Unit); ok {
			unit.InterruptSpell(CURRENT_GENERIC_SPELL)
		}
	}

}

// HandleCancelChannellingOpcode 处理取消引导操作码
func (ws *WorldSession) HandleCancelChannellingOpcode(packet *WorldPacket) {
	fmt.Printf("玩家 %s 取消引导\n", ws.GetPlayerInfo())

	player := ws.GetPlayer()
	if player != nil {
		if unit, ok := player.(*Unit); ok {
			unit.InterruptSpell(CURRENT_CHANNELED_SPELL)
		}
	}

}

// HandleKeepAliveOpcode 处理保持连接操作码
func (ws *WorldSession) HandleKeepAliveOpcode(packet *WorldPacket) {
	ws.ResetTimeOutTime(true)
}

// HandleDamageTakenOpcode 处理受到伤害操作码
func (ws *WorldSession) HandleDamageTakenOpcode(packet *WorldPacket) {
	targetGuid := packet.ReadUint64()
	damage := packet.ReadUint32()

	player := ws.GetPlayer()
	if player == nil || player.GetGUID() != targetGuid {
		return
	}

	fmt.Printf("玩家 %s 受到伤害: %d\n", ws.GetPlayerInfo(), damage)

	// 处理伤害
	oldHealth := player.GetHealth()
	newHealth := oldHealth
	if oldHealth > damage {
		newHealth = oldHealth - damage
	} else {
		newHealth = 1 // 保持至少1点血
	}

	player.SetHealth(newHealth)

	// 发送血量更新给客户端
	ws.SendHealthUpdate(player, newHealth, player.GetMaxHealth())

	// 广播血量更新给其他玩家
	ws.world.BroadcastHealthUpdate(player, oldHealth, newHealth)

	// 添加批量更新 - 基于AzerothCore的血量同步
	if unit, ok := player.(*Unit); ok {
		unit.AddBatchUpdateForFullState() // 血量变化需要完整状态更新
	}

	// 使用oldHealth避免编译警告
	_ = oldHealth

}

// HandleMoveStartForwardOpcode 处理开始前进操作码 - 基于AzerothCore的移动同步
func (ws *WorldSession) HandleMoveStartForwardOpcode(packet *WorldPacket) {
	// 读取移动数据
	x := packet.ReadFloat32()
	y := packet.ReadFloat32()
	z := packet.ReadFloat32()
	orientation := packet.ReadFloat32()

	player := ws.GetPlayer()
	if player == nil {
		return
	}

	fmt.Printf("玩家 %s 开始前进到位置: (%.2f, %.2f, %.2f), 朝向: %.2f\n",
		ws.GetPlayerInfo(), x, y, z, orientation)

	// 更新玩家位置
	if unit, ok := player.(*Unit); ok {
		unit.SetPosition(x, y, z)
		unit.orientation = orientation

		// 添加批量更新 - 基于AzerothCore的移动同步
		unit.AddBatchUpdateForMovement() // 移动需要位置更新

		fmt.Printf("[BatchUpdate] 移动更新: %s 位置(%.2f, %.2f, %.2f)\n",
			unit.GetName(), x, y, z)
	}
}

// HandleMoveStopOpcode 处理停止移动操作码 - 基于AzerothCore的移动同步
func (ws *WorldSession) HandleMoveStopOpcode(packet *WorldPacket) {
	// 读取停止位置数据
	x := packet.ReadFloat32()
	y := packet.ReadFloat32()
	z := packet.ReadFloat32()
	orientation := packet.ReadFloat32()

	player := ws.GetPlayer()
	if player == nil {
		return
	}

	fmt.Printf("玩家 %s 停止移动在位置: (%.2f, %.2f, %.2f), 朝向: %.2f\n",
		ws.GetPlayerInfo(), x, y, z, orientation)

	// 更新玩家位置
	if unit, ok := player.(*Unit); ok {
		unit.SetPosition(x, y, z)
		unit.orientation = orientation

		// 添加批量更新 - 基于AzerothCore的移动同步
		unit.AddBatchUpdateForMovement() // 停止移动也需要位置更新

		fmt.Printf("[BatchUpdate] 停止移动更新: %s 位置(%.2f, %.2f, %.2f)\n",
			unit.GetName(), x, y, z)
	}
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

// GetNextReceivedPacket 获取下一个接收到的数据包（用于客户端处理）
func (ws *WorldSession) GetNextReceivedPacket() *WorldPacket {
	select {
	case packet := <-ws._receivedQueue:
		return packet
	default:
		return nil
	}
}

// QueueReceivedPacket 将数据包加入客户端接收队列
func (ws *WorldSession) QueueReceivedPacket(packet *WorldPacket) {
	if ws._receivedQueue == nil {
		return
	}

	select {
	case ws._receivedQueue <- packet:
		// 成功加入队列
	default:
		fmt.Printf("会话 %d 客户端接收队列已满，丢弃数据包: 0x%X\n", ws.id, packet.GetOpcode())
	}
}

// SendHealthUpdate 发送血量更新
func (ws *WorldSession) SendHealthUpdate(unit IUnit, health, maxHealth uint32) {
	packet := NewWorldPacket(SMSG_HEALTH_UPDATE)
	packet.WriteUint64(unit.GetGUID())
	packet.WriteUint32(health)
	packet.WriteUint32(maxHealth)
	ws.SendPacket(packet)
}
