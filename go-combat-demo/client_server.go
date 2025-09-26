package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// GameClient 游戏客户端 - 模拟客户端行为
type GameClient struct {
	id      uint32
	name    string
	conn    net.Conn
	session *WorldSession
	player  *Player
	target  IUnit
	socket  *WorldSocket
	world   *World
	mutex   sync.RWMutex
	running bool
}

// NewGameClient 创建游戏客户端
func NewGameClient(id uint32, name string, world *World) *GameClient {
	return &GameClient{
		id:      id,
		name:    name,
		world:   world,
		running: false,
	}
}

// Connect 连接到服务器
func (gc *GameClient) Connect(serverAddr string) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}

	gc.conn = conn
	gc.socket = NewWorldSocket(conn)
	gc.session = NewWorldSession(gc.id, gc.name, gc.socket, gc.world)
	gc.running = true

	fmt.Printf("客户端 %s 已连接到服务器\n", gc.name)
	return nil
}

// Login 登录玩家
func (gc *GameClient) Login(player *Player) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	gc.player = player
	gc.session.SetPlayer(player)

	fmt.Printf("客户端 %s 登录玩家: %s\n", gc.name, player.GetName())
}

// SetTarget 设置目标
func (gc *GameClient) SetTarget(target IUnit) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	gc.target = target

	// 发送设置选择目标数据包
	packet := NewWorldPacket(CMSG_SET_SELECTION)
	if target != nil {
		packet.WriteUint64(target.GetGUID())
		fmt.Printf("客户端 %s 选择目标: %s\n", gc.name, target.GetName())
	} else {
		packet.WriteUint64(0)
		fmt.Printf("客户端 %s 取消选择目标\n", gc.name)
	}

	gc.session.SendPacket(packet)
}

// Attack 攻击目标
func (gc *GameClient) Attack(target IUnit) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	if target == nil {
		fmt.Printf("客户端 %s 没有攻击目标\n", gc.name)
		return
	}

	gc.target = target

	// 发送攻击数据包
	packet := NewWorldPacket(CMSG_ATTACKSWING)
	packet.WriteUint64(target.GetGUID())
	gc.session.SendPacket(packet)

	fmt.Printf("客户端 %s 发起攻击: %s\n", gc.name, target.GetName())
}

// StopAttack 停止攻击
func (gc *GameClient) StopAttack() {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// 发送停止攻击数据包
	packet := NewWorldPacket(CMSG_ATTACKSTOP)
	gc.session.SendPacket(packet)

	fmt.Printf("客户端 %s 停止攻击\n", gc.name)
}

// CastSpell 施放法术
func (gc *GameClient) CastSpell(spellId uint32, target IUnit) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// 发送施放法术数据包
	packet := NewWorldPacket(CMSG_CAST_SPELL)
	packet.WriteUint32(spellId)
	if target != nil {
		packet.WriteUint64(target.GetGUID())
		fmt.Printf("客户端 %s 对 %s 施放法术 %d\n", gc.name, target.GetName(), spellId)
	} else {
		packet.WriteUint64(0)
		fmt.Printf("客户端 %s 施放法术 %d\n", gc.name, spellId)
	}

	gc.session.SendPacket(packet)
}

// SendKeepAlive 发送保持连接
func (gc *GameClient) SendKeepAlive() {
	packet := NewWorldPacket(CMSG_KEEP_ALIVE)
	gc.session.SendPacket(packet)
}

// Update 更新客户端 - 客户端只处理UI和发送指令，不处理服务器逻辑
func (gc *GameClient) Update(diff uint32) {
	if !gc.running || gc.session == nil {
		return
	}

	// 检查连接状态
	if !gc.session.IsConnected() {
		gc.Disconnect()
		return
	}

	// 客户端只负责发送心跳包，不处理数据包
	// 数据包处理由服务器端的session.Update()负责
	gc.SendKeepAlive()
}

// Disconnect 断开连接
func (gc *GameClient) Disconnect() {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	if !gc.running {
		return
	}

	gc.running = false

	if gc.session != nil {
		gc.session.Close()
	}

	if gc.conn != nil {
		gc.conn.Close()
	}

	fmt.Printf("客户端 %s 已断开连接\n", gc.name)
}

// GetPlayer 获取玩家
func (gc *GameClient) GetPlayer() *Player {
	gc.mutex.RLock()
	defer gc.mutex.RUnlock()
	return gc.player
}

// GetTarget 获取目标
func (gc *GameClient) GetTarget() IUnit {
	gc.mutex.RLock()
	defer gc.mutex.RUnlock()
	return gc.target
}

// IsRunning 是否运行中
func (gc *GameClient) IsRunning() bool {
	gc.mutex.RLock()
	defer gc.mutex.RUnlock()
	return gc.running
}

// GameServer 游戏服务器 - 基于AzerothCore的World类
type GameServer struct {
	listener    net.Listener
	sessions    map[uint32]*WorldSession
	clients     map[uint32]*GameClient
	world       *World
	running     bool
	mutex       sync.RWMutex
	nextId      uint32
	updateTimer *time.Ticker
}

// NewGameServer 创建游戏服务器
func NewGameServer(world *World) *GameServer {
	return &GameServer{
		sessions:    make(map[uint32]*WorldSession),
		clients:     make(map[uint32]*GameClient),
		world:       world,
		running:     false,
		nextId:      1,
		updateTimer: time.NewTicker(200 * time.Millisecond), // 200ms更新间隔
	}
}

// Start 启动服务器
func (gs *GameServer) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("启动服务器失败: %v", err)
	}

	gs.listener = listener
	gs.running = true

	fmt.Printf("游戏服务器已启动，监听地址: %s\n", addr)

	// 启动更新循环
	go gs.updateLoop()

	// 接受连接
	go gs.acceptLoop()

	return nil
}

// acceptLoop 接受连接循环
func (gs *GameServer) acceptLoop() {
	for gs.running {
		conn, err := gs.listener.Accept()
		if err != nil {
			if gs.running {
				fmt.Printf("接受连接失败: %v\n", err)
			}
			continue
		}

		go gs.handleConnection(conn)
	}
}

// handleConnection 处理连接 - 基于AzerothCore的连接管理逻辑
func (gs *GameServer) handleConnection(conn net.Conn) {
	gs.mutex.Lock()
	sessionId := gs.nextId
	gs.nextId++
	gs.mutex.Unlock()

	socket := NewWorldSocket(conn)
	session := NewWorldSession(sessionId, fmt.Sprintf("Account_%d", sessionId), socket, gs.world)

	gs.mutex.Lock()
	gs.sessions[sessionId] = session
	gs.mutex.Unlock()

	fmt.Printf("新会话连接: ID %d\n", sessionId)

	// 基于AzerothCore的设计：连接线程只负责维持连接状态
	// 数据包处理由World::UpdateSessions()在主循环中完成
	for gs.running {
		// 检查连接是否还有效
		if !session.IsConnected() {
			break
		}

		// 只维持连接，不处理数据包
		time.Sleep(100 * time.Millisecond)
	}

	// 清理会话
	gs.mutex.Lock()
	delete(gs.sessions, sessionId)
	gs.mutex.Unlock()

	session.Close()
	fmt.Printf("会话 %d 已断开\n", sessionId)
}

// updateLoop 更新循环 - 基于AzerothCore的World::Update
func (gs *GameServer) updateLoop() {
	for range gs.updateTimer.C {
		if !gs.running {
			break
		}

		gs.Update(200) // 200ms更新间隔
	}
}

// Update 更新服务器 - 基于AzerothCore的World::Update
func (gs *GameServer) Update(diff uint32) {
	// 更新世界
	if gs.world != nil {
		gs.world.Update(diff)
	}

	// 更新所有会话 - 基于AzerothCore的WorldSessionMgr::UpdateSessions
	gs.UpdateSessions(diff)
}

// UpdateSessions 更新所有会话 - 基于AzerothCore的WorldSessionMgr::UpdateSessions
func (gs *GameServer) UpdateSessions(diff uint32) {
	gs.mutex.RLock()
	sessions := make([]*WorldSession, 0, len(gs.sessions))
	for _, session := range gs.sessions {
		sessions = append(sessions, session)
	}
	gs.mutex.RUnlock()

	// 在读锁外更新会话，避免死锁
	// 这里处理所有数据包 - 符合AzerothCore的设计
	for _, session := range sessions {
		if !session.Update(diff) {
			// 会话更新失败，标记为需要移除
			// 在实际实现中，这里应该标记会话为待删除
			// 为了简化demo，我们在这里直接忽略
		}
	}
}

// Stop 停止服务器
func (gs *GameServer) Stop() {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	if !gs.running {
		return
	}

	gs.running = false

	// 停止更新计时器
	if gs.updateTimer != nil {
		gs.updateTimer.Stop()
	}

	// 关闭监听器
	if gs.listener != nil {
		gs.listener.Close()
	}

	// 关闭所有会话
	for _, session := range gs.sessions {
		session.Close()
	}

	fmt.Println("游戏服务器已停止")
}

// GetSessionCount 获取会话数量
func (gs *GameServer) GetSessionCount() int {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()
	return len(gs.sessions)
}

// BroadcastPacket 广播数据包
func (gs *GameServer) BroadcastPacket(packet *WorldPacket) {
	gs.mutex.RLock()
	sessions := make([]*WorldSession, 0, len(gs.sessions))
	for _, session := range gs.sessions {
		sessions = append(sessions, session)
	}
	gs.mutex.RUnlock()

	for _, session := range sessions {
		session.SendPacket(packet)
	}
}

// GetWorld 获取世界
func (gs *GameServer) GetWorld() *World {
	return gs.world
}
