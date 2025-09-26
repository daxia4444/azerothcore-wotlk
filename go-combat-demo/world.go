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

// 世界管理器 - 基于AzerothCore的World类
type World struct {
	units    map[uint64]IUnit         // 所有单位的映射
	sessions map[uint32]*WorldSession // 所有会话的映射
	mutex    sync.RWMutex             // 读写锁
	nextGUID uint64                   // 下一个GUID
}

func NewWorld() *World {
	return &World{
		units:    make(map[uint64]IUnit),
		sessions: make(map[uint32]*WorldSession),
		nextGUID: 1,
	}
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

// GetUnitByGUID 根据GUID获取单位
func (w *World) GetUnitByGUID(guid uint64) IUnit {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.units[guid]
}

// BroadcastPacket 向所有会话广播数据包
func (w *World) BroadcastPacket(packet *WorldPacket) {
	w.mutex.RLock()
	sessions := make([]*WorldSession, 0, len(w.sessions))
	for _, session := range w.sessions {
		sessions = append(sessions, session)
	}
	w.mutex.RUnlock()

	for _, session := range sessions {
		session.SendPacket(packet)
	}
}

// BroadcastSpellStart 广播法术开始 - 基于AzerothCore的SMSG_SPELL_START
func (w *World) BroadcastSpellStart(caster IUnit, spellId uint32, targets []IUnit, castTime time.Duration) {
	packet := NewWorldPacket(SMSG_SPELL_START)
	packet.WriteUint64(caster.GetGUID())
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

	w.BroadcastPacket(packet)

	// 获取法术信息用于日志
	spellInfo := GlobalSpellManager.GetSpell(spellId)
	spellName := "未知法术"
	if spellInfo != nil {
		spellName = spellInfo.Name
	}

	fmt.Printf("[网络] 广播法术开始: %s 施放 %s\n", caster.GetName(), spellName)
}

// BroadcastSpellGo 广播法术生效 - 基于AzerothCore的SMSG_SPELL_GO
func (w *World) BroadcastSpellGo(caster IUnit, spellId uint32, targets []IUnit) {
	packet := NewWorldPacket(SMSG_SPELLGO)
	packet.WriteUint64(caster.GetGUID())
	packet.WriteUint32(spellId)
	packet.WriteUint32(uint32(len(targets)))

	for _, target := range targets {
		if target != nil {
			packet.WriteUint64(target.GetGUID())
		} else {
			packet.WriteUint64(0)
		}
	}

	w.BroadcastPacket(packet)

	// 获取法术信息用于日志
	spellInfo := GlobalSpellManager.GetSpell(spellId)
	spellName := "未知法术"
	if spellInfo != nil {
		spellName = spellInfo.Name
	}

	fmt.Printf("[网络] 广播法术生效: %s 的 %s 生效\n", caster.GetName(), spellName)
}

// BroadcastAttackerStateUpdate 广播攻击状态更新 - 基于AzerothCore的SMSG_ATTACKERSTATEUPDATE
func (w *World) BroadcastAttackerStateUpdate(attacker, victim IUnit, damage uint32, hitResult int, schoolMask int) {
	packet := NewWorldPacket(SMSG_ATTACKERSTATEUPDATE)
	packet.WriteUint32(uint32(hitResult))
	packet.WriteUint64(attacker.GetGUID())
	packet.WriteUint64(victim.GetGUID())
	packet.WriteUint32(damage)
	packet.WriteUint32(uint32(schoolMask))

	w.BroadcastPacket(packet)

	fmt.Printf("[网络] 广播攻击状态: %s 对 %s 造成 %d 伤害\n",
		attacker.GetName(), victim.GetName(), damage)
}

// BroadcastUnitUpdate 广播单位状态更新 - 基于AzerothCore的SMSG_UPDATE_OBJECT
func (w *World) BroadcastUnitUpdate(unit IUnit) {
	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint64(unit.GetGUID())
	packet.WriteUint32(unit.GetHealth())
	packet.WriteUint32(unit.GetMaxHealth())

	// 添加能量信息
	packet.WriteUint32(unit.GetPower(POWER_MANA))
	packet.WriteUint32(unit.GetMaxPower(POWER_MANA))

	w.BroadcastPacket(packet)
}

// BroadcastHealthUpdate 广播血量更新 - 基于AzerothCore的即时血量同步
func (w *World) BroadcastHealthUpdate(unit IUnit, oldHealth, newHealth uint32) {
	// 创建血量更新数据包
	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint64(unit.GetGUID())
	packet.WriteUint32(1) // 更新类型：血量
	packet.WriteUint32(oldHealth)
	packet.WriteUint32(newHealth)
	packet.WriteUint32(unit.GetMaxHealth())

	w.BroadcastPacket(packet)

	fmt.Printf("[网络] 广播血量更新: %s %d->%d/%d\n",
		unit.GetName(), oldHealth, newHealth, unit.GetMaxHealth())
}

// BroadcastPowerUpdate 广播能量更新 - 基于AzerothCore的SMSG_POWER_UPDATE
func (w *World) BroadcastPowerUpdate(unit IUnit, powerType int, oldPower, newPower uint32) {
	// 创建能量更新数据包
	packet := NewWorldPacket(SMSG_UPDATE_OBJECT)
	packet.WriteUint64(unit.GetGUID())
	packet.WriteUint32(2) // 更新类型：能量
	packet.WriteUint32(uint32(powerType))
	packet.WriteUint32(oldPower)
	packet.WriteUint32(newPower)
	packet.WriteUint32(unit.GetMaxPower(powerType))

	w.BroadcastPacket(packet)

	powerName := getPowerTypeName(powerType)
	fmt.Printf("[网络] 广播%s更新: %s %d->%d/%d\n",
		powerName, unit.GetName(), oldPower, newPower, unit.GetMaxPower(powerType))
}

// 获取能量类型名称
func getPowerTypeName(powerType int) string {
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

// Update 更新世界 - 基于AzerothCore的World::Update
func (w *World) Update(diff uint32) {
	w.mutex.RLock()
	units := make([]IUnit, 0, len(w.units))
	for _, unit := range w.units {
		units = append(units, unit)
	}
	w.mutex.RUnlock()

	// 更新所有单位
	for _, unit := range units {
		if unit.IsAlive() {
			unit.Update(diff)
		}
	}

	// 定期广播完整状态更新 - 基于AzerothCore的批量更新机制
	w.broadcastPeriodicUpdates(diff)

	// 清理死亡单位（可选）
	w.cleanupDeadUnits()
}

// 定期广播状态更新 - 基于AzerothCore的定期同步机制
var lastPeriodicUpdate uint32 = 0

const PERIODIC_UPDATE_INTERVAL = 5000 // 5秒

func (w *World) broadcastPeriodicUpdates(diff uint32) {
	lastPeriodicUpdate += diff
	if lastPeriodicUpdate >= PERIODIC_UPDATE_INTERVAL {
		lastPeriodicUpdate = 0

		// 广播所有单位的完整状态
		w.mutex.RLock()
		for _, unit := range w.units {
			if unit.IsAlive() {
				w.BroadcastUnitUpdate(unit)
			}
		}
		w.mutex.RUnlock()

		fmt.Printf("[网络] 定期状态同步完成\n")
	}
}

func (w *World) cleanupDeadUnits() {
	// 简化版：不立即清理死亡单位，让它们保持在世界中用于演示
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
