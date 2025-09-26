package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 法术相关常量 - 基于AzerothCore的SpellDefines.h
const (
	// 法术状态 - 基于AzerothCore的SpellState
	SPELL_STATE_NULL      = 0 // 空状态
	SPELL_STATE_PREPARING = 1 // 准备中（施法时间）
	SPELL_STATE_CASTING   = 2 // 施法中
	SPELL_STATE_FINISHED  = 3 // 完成
	SPELL_STATE_IDLE      = 4 // 空闲
	SPELL_STATE_DELAYED   = 5 // 延迟

	// 法术类型 - 基于AzerothCore的CurrentSpellTypes
	CURRENT_MELEE_SPELL      = 0 // 近战法术
	CURRENT_GENERIC_SPELL    = 1 // 通用法术
	CURRENT_CHANNELED_SPELL  = 2 // 引导法术
	CURRENT_AUTOREPEAT_SPELL = 3 // 自动重复法术

	// 法术效果类型 - 基于AzerothCore的SpellEffects
	SPELL_EFFECT_NONE          = 0   // 无效果
	SPELL_EFFECT_INSTAKILL     = 1   // 即死
	SPELL_EFFECT_SCHOOL_DAMAGE = 2   // 学派伤害
	SPELL_EFFECT_DUMMY         = 3   // 虚拟效果
	SPELL_EFFECT_HEAL          = 10  // 治疗
	SPELL_EFFECT_ENERGIZE      = 43  // 回复能量
	SPELL_EFFECT_WEAPON_DAMAGE = 121 // 武器伤害

	// 法术目标类型 - 基于AzerothCore的Targets
	TARGET_UNIT_TARGET_ENEMY = 6  // 敌方单体目标
	TARGET_UNIT_CASTER       = 1  // 施法者自己
	TARGET_UNIT_TARGET_ALLY  = 7  // 友方单体目标
	TARGET_DEST_TARGET_ENEMY = 16 // 敌方目标位置

	// 法术属性 - 基于AzerothCore的SpellAttr
	SPELL_ATTR0_ON_NEXT_SWING_1               = 0x00000004 // 下次攻击触发
	SPELL_ATTR0_DONT_AFFECT_SHEATH_STATE      = 0x00000008 // 不影响武器状态
	SPELL_ATTR0_LEVEL_DAMAGE_CALCULATION      = 0x00000020 // 等级伤害计算
	SPELL_ATTR0_STOP_ATTACK_TARGET            = 0x00000040 // 停止攻击目标
	SPELL_ATTR0_IMPOSSIBLE_DODGE_PARRY_BLOCK  = 0x00000080 // 无法闪避招架格挡
	SPELL_ATTR0_CAST_TRACK_TARGET             = 0x00000100 // 施法时跟踪目标
	SPELL_ATTR0_CASTABLE_WHILE_DEAD           = 0x00000200 // 死亡时可施放
	SPELL_ATTR0_CASTABLE_WHILE_MOUNTED        = 0x00000400 // 骑乘时可施放
	SPELL_ATTR0_DISABLED_WHILE_ACTIVE         = 0x00000800 // 激活时禁用
	SPELL_ATTR0_NEGATIVE                      = 0x00001000 // 负面法术
	SPELL_ATTR0_CASTABLE_WHILE_SITTING        = 0x00002000 // 坐着时可施放
	SPELL_ATTR0_CANT_USED_IN_COMBAT           = 0x00004000 // 战斗中无法使用
	SPELL_ATTR0_UNAFFECTED_BY_INVULNERABILITY = 0x00008000 // 不受无敌影响
	SPELL_ATTR0_HEARTBEAT_RESIST_CHECK        = 0x00010000 // 心跳抗性检查
	SPELL_ATTR0_CANT_CANCEL                   = 0x00020000 // 无法取消
)

// 法术ID定义 - 基于经典魔兽世界法术
const (
	// 法师法术
	SPELL_FROSTBOLT       = 116  // 寒冰箭 - 施法法术
	SPELL_FIREBALL        = 133  // 火球术 - 施法法术
	SPELL_BLIZZARD        = 10   // 暴风雪 - 引导法术
	SPELL_FROST_NOVA      = 122  // 冰霜新星 - 即时法术
	SPELL_ARCANE_MISSILES = 5143 // 奥术飞弹 - 引导法术
	SPELL_POLYMORPH       = 118  // 变羊术 - 施法法术
	SPELL_COUNTERSPELL    = 2139 // 法术反制 - 即时法术
	SPELL_BLINK           = 1953 // 闪现 - 即时法术

	// 牧师法术
	SPELL_HEAL              = 2050 // 治疗术 - 施法法术
	SPELL_FLASH_HEAL        = 2061 // 快速治疗 - 施法法术
	SPELL_RENEW             = 139  // 恢复 - 即时法术(HOT)
	SPELL_POWER_WORD_SHIELD = 17   // 真言术：盾 - 即时法术
	SPELL_HOLY_LIGHT        = 635  // 圣光术 - 施法法术
	SPELL_SMITE             = 585  // 惩击 - 施法法术

	// 术士法术
	SPELL_SHADOW_BOLT = 686  // 暗影箭 - 施法法术
	SPELL_IMMOLATE    = 348  // 献祭 - 即时法术(DOT)
	SPELL_CORRUPTION  = 172  // 腐蚀术 - 即时法术(DOT)
	SPELL_DRAIN_LIFE  = 689  // 吸取生命 - 引导法术
	SPELL_FEAR        = 5782 // 恐惧术 - 施法法术

	// 战士技能
	SPELL_HEROIC_STRIKE = 78    // 英勇打击 - 即时技能
	SPELL_CHARGE        = 100   // 冲锋 - 即时技能
	SPELL_TAUNT         = 355   // 嘲讽 - 即时技能
	SPELL_SHIELD_SLAM   = 23922 // 盾牌猛击 - 即时技能

	// 猎人技能
	SPELL_AIMED_SHOT  = 19434 // 瞄准射击 - 施法技能
	SPELL_MULTI_SHOT  = 2643  // 多重射击 - 即时技能
	SPELL_HUNTER_MARK = 1130  // 猎人印记 - 即时技能
)

// SpellInfo 法术信息 - 基于AzerothCore的SpellInfo
type SpellInfo struct {
	ID             uint32        // 法术ID
	Name           string        // 法术名称
	Description    string        // 法术描述
	CastTime       time.Duration // 施法时间
	Cooldown       time.Duration // 冷却时间
	ManaCost       uint32        // 法力消耗
	Range          float32       // 施法距离
	SchoolMask     int           // 法术学派
	Effects        []SpellEffect // 法术效果
	Attributes     uint32        // 法术属性
	TargetType     int           // 目标类型
	IsChanneled    bool          // 是否为引导法术
	ChannelTime    time.Duration // 引导时间
	BaseDamage     uint32        // 基础伤害
	DamageVariance float32       // 伤害浮动
	Level          uint8         // 法术等级
}

// SpellEffect 法术效果
type SpellEffect struct {
	EffectType         int     // 效果类型
	BasePoints         int32   // 基础点数
	DicePerLevel       float32 // 每级骰子数
	RealPointsPerLevel float32 // 每级真实点数
	Mechanic           int     // 机制类型
	ImplicitTargetA    int     // 隐式目标A
	ImplicitTargetB    int     // 隐式目标B
	RadiusIndex        int     // 范围索引
	ApplyAuraName      int     // 应用光环名称
	Amplitude          int32   // 振幅（DOT/HOT间隔）
	MultipleValue      float32 // 倍数值
}

// Spell 法术实例 - 基于AzerothCore的Spell类
type Spell struct {
	info        *SpellInfo    // 法术信息
	caster      IUnit         // 施法者
	targets     []IUnit       // 目标列表
	state       int           // 法术状态
	castTime    time.Duration // 剩余施法时间
	channelTime time.Duration // 剩余引导时间
	startTime   time.Time     // 开始时间
	damage      uint32        // 计算出的伤害
	healing     uint32        // 计算出的治疗
	interrupted bool          // 是否被打断
	world       *World        // 世界引用
}

// SpellManager 法术管理器 - 管理所有法术信息
type SpellManager struct {
	spells map[uint32]*SpellInfo // 法术信息表
}

// 全局法术管理器
var GlobalSpellManager *SpellManager

// 初始化法术管理器
func InitSpellManager() {
	GlobalSpellManager = &SpellManager{
		spells: make(map[uint32]*SpellInfo),
	}
	GlobalSpellManager.LoadSpells()
}

// LoadSpells 加载法术信息 - 基于AzerothCore的spell_template表
func (sm *SpellManager) LoadSpells() {
	// 法师法术
	sm.AddSpell(&SpellInfo{
		ID:             SPELL_FROSTBOLT,
		Name:           "寒冰箭",
		Description:    "向目标发射一枚寒冰箭，造成冰霜伤害并降低移动速度",
		CastTime:       2500 * time.Millisecond, // 2.5秒施法时间
		Cooldown:       0,
		ManaCost:       125,
		Range:          30.0,
		SchoolMask:     SPELL_SCHOOL_FROST,
		TargetType:     TARGET_UNIT_TARGET_ENEMY,
		BaseDamage:     350,
		DamageVariance: 0.15, // 15%伤害浮动
		Level:          20,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_SCHOOL_DAMAGE,
				BasePoints:      350,
				DicePerLevel:    2.8,
				ImplicitTargetA: TARGET_UNIT_TARGET_ENEMY,
			},
		},
	})

	sm.AddSpell(&SpellInfo{
		ID:             SPELL_FIREBALL,
		Name:           "火球术",
		Description:    "向目标发射一枚火球，造成火焰伤害",
		CastTime:       3000 * time.Millisecond, // 3秒施法时间
		Cooldown:       0,
		ManaCost:       155,
		Range:          35.0,
		SchoolMask:     SPELL_SCHOOL_FIRE,
		TargetType:     TARGET_UNIT_TARGET_ENEMY,
		BaseDamage:     450,
		DamageVariance: 0.2, // 20%伤害浮动
		Level:          25,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_SCHOOL_DAMAGE,
				BasePoints:      450,
				DicePerLevel:    3.2,
				ImplicitTargetA: TARGET_UNIT_TARGET_ENEMY,
			},
		},
	})

	sm.AddSpell(&SpellInfo{
		ID:             SPELL_FROST_NOVA,
		Name:           "冰霜新星",
		Description:    "冻结周围的敌人，造成冰霜伤害并定身",
		CastTime:       0, // 即时法术
		Cooldown:       25 * time.Second,
		ManaCost:       85,
		Range:          0, // 以自身为中心
		SchoolMask:     SPELL_SCHOOL_FROST,
		TargetType:     TARGET_UNIT_CASTER,
		BaseDamage:     180,
		DamageVariance: 0.1,
		Level:          10,
		Attributes:     SPELL_ATTR0_IMPOSSIBLE_DODGE_PARRY_BLOCK,
		Effects: []SpellEffect{
			{
				EffectType:   SPELL_EFFECT_SCHOOL_DAMAGE,
				BasePoints:   180,
				DicePerLevel: 1.5,
				RadiusIndex:  8, // 8码范围
			},
		},
	})

	sm.AddSpell(&SpellInfo{
		ID:             SPELL_BLIZZARD,
		Name:           "暴风雪",
		Description:    "在目标区域召唤暴风雪，持续造成冰霜伤害",
		CastTime:       0, // 即时施法
		Cooldown:       8 * time.Second,
		ManaCost:       320,
		Range:          35.0,
		SchoolMask:     SPELL_SCHOOL_FROST,
		TargetType:     TARGET_DEST_TARGET_ENEMY,
		IsChanneled:    true,
		ChannelTime:    8 * time.Second, // 8秒引导
		BaseDamage:     120,
		DamageVariance: 0.25,
		Level:          20,
		Effects: []SpellEffect{
			{
				EffectType:   SPELL_EFFECT_SCHOOL_DAMAGE,
				BasePoints:   120,
				DicePerLevel: 2.0,
				Amplitude:    1000, // 每秒触发
				RadiusIndex:  8,    // 8码范围
			},
		},
	})

	// 牧师法术
	sm.AddSpell(&SpellInfo{
		ID:             SPELL_HEAL,
		Name:           "治疗术",
		Description:    "治疗友方目标",
		CastTime:       3000 * time.Millisecond, // 3秒施法
		Cooldown:       0,
		ManaCost:       155,
		Range:          40.0,
		SchoolMask:     SPELL_SCHOOL_HOLY,
		TargetType:     TARGET_UNIT_TARGET_ALLY,
		BaseDamage:     600, // 这里用作治疗量
		DamageVariance: 0.15,
		Level:          15,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_HEAL,
				BasePoints:      600,
				DicePerLevel:    4.0,
				ImplicitTargetA: TARGET_UNIT_TARGET_ALLY,
			},
		},
	})

	sm.AddSpell(&SpellInfo{
		ID:             SPELL_FLASH_HEAL,
		Name:           "快速治疗",
		Description:    "快速治疗友方目标",
		CastTime:       1500 * time.Millisecond, // 1.5秒施法
		Cooldown:       0,
		ManaCost:       215,
		Range:          40.0,
		SchoolMask:     SPELL_SCHOOL_HOLY,
		TargetType:     TARGET_UNIT_TARGET_ALLY,
		BaseDamage:     400, // 治疗量
		DamageVariance: 0.2,
		Level:          20,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_HEAL,
				BasePoints:      400,
				DicePerLevel:    3.0,
				ImplicitTargetA: TARGET_UNIT_TARGET_ALLY,
			},
		},
	})

	sm.AddSpell(&SpellInfo{
		ID:             SPELL_POWER_WORD_SHIELD,
		Name:           "真言术：盾",
		Description:    "为目标提供伤害吸收护盾",
		CastTime:       0, // 即时法术
		Cooldown:       4 * time.Second,
		ManaCost:       125,
		Range:          30.0,
		SchoolMask:     SPELL_SCHOOL_HOLY,
		TargetType:     TARGET_UNIT_TARGET_ALLY,
		BaseDamage:     500, // 护盾吸收量
		DamageVariance: 0.1,
		Level:          12,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_DUMMY, // 护盾效果
				BasePoints:      500,
				DicePerLevel:    3.5,
				ImplicitTargetA: TARGET_UNIT_TARGET_ALLY,
			},
		},
	})

	// 术士法术
	sm.AddSpell(&SpellInfo{
		ID:             SPELL_SHADOW_BOLT,
		Name:           "暗影箭",
		Description:    "向目标发射暗影能量，造成暗影伤害",
		CastTime:       2500 * time.Millisecond, // 2.5秒施法
		Cooldown:       0,
		ManaCost:       140,
		Range:          30.0,
		SchoolMask:     SPELL_SCHOOL_SHADOW,
		TargetType:     TARGET_UNIT_TARGET_ENEMY,
		BaseDamage:     380,
		DamageVariance: 0.18,
		Level:          18,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_SCHOOL_DAMAGE,
				BasePoints:      380,
				DicePerLevel:    3.0,
				ImplicitTargetA: TARGET_UNIT_TARGET_ENEMY,
			},
		},
	})

	sm.AddSpell(&SpellInfo{
		ID:             SPELL_IMMOLATE,
		Name:           "献祭",
		Description:    "点燃目标，立即造成火焰伤害并持续燃烧",
		CastTime:       2000 * time.Millisecond, // 2秒施法
		Cooldown:       0,
		ManaCost:       110,
		Range:          30.0,
		SchoolMask:     SPELL_SCHOOL_FIRE,
		TargetType:     TARGET_UNIT_TARGET_ENEMY,
		BaseDamage:     180,
		DamageVariance: 0.15,
		Level:          8,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_SCHOOL_DAMAGE,
				BasePoints:      180,
				DicePerLevel:    1.8,
				ImplicitTargetA: TARGET_UNIT_TARGET_ENEMY,
			},
		},
	})

	// 战士技能
	sm.AddSpell(&SpellInfo{
		ID:             SPELL_HEROIC_STRIKE,
		Name:           "英勇打击",
		Description:    "下次近战攻击造成额外伤害",
		CastTime:       0, // 即时技能
		Cooldown:       0,
		ManaCost:       0, // 消耗怒气
		Range:          5.0,
		SchoolMask:     SPELL_SCHOOL_NORMAL,
		TargetType:     TARGET_UNIT_TARGET_ENEMY,
		BaseDamage:     150,
		DamageVariance: 0.1,
		Level:          1,
		Attributes:     SPELL_ATTR0_ON_NEXT_SWING_1,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_WEAPON_DAMAGE,
				BasePoints:      150,
				DicePerLevel:    2.0,
				ImplicitTargetA: TARGET_UNIT_TARGET_ENEMY,
			},
		},
	})

	sm.AddSpell(&SpellInfo{
		ID:          SPELL_TAUNT,
		Name:        "嘲讽",
		Description: "强制敌人攻击你",
		CastTime:    0, // 即时技能
		Cooldown:    10 * time.Second,
		ManaCost:    0,
		Range:       5.0,
		SchoolMask:  SPELL_SCHOOL_NORMAL,
		TargetType:  TARGET_UNIT_TARGET_ENEMY,
		BaseDamage:  0,
		Level:       10,
		Attributes:  SPELL_ATTR0_IMPOSSIBLE_DODGE_PARRY_BLOCK,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_DUMMY, // 嘲讽效果
				BasePoints:      0,
				ImplicitTargetA: TARGET_UNIT_TARGET_ENEMY,
			},
		},
	})

	// 猎人技能
	sm.AddSpell(&SpellInfo{
		ID:             SPELL_AIMED_SHOT,
		Name:           "瞄准射击",
		Description:    "精确瞄准射击，造成大量伤害",
		CastTime:       3000 * time.Millisecond, // 3秒施法
		Cooldown:       6 * time.Second,
		ManaCost:       0, // 消耗集中值
		Range:          35.0,
		SchoolMask:     SPELL_SCHOOL_NORMAL,
		TargetType:     TARGET_UNIT_TARGET_ENEMY,
		BaseDamage:     550,
		DamageVariance: 0.12,
		Level:          20,
		Effects: []SpellEffect{
			{
				EffectType:      SPELL_EFFECT_WEAPON_DAMAGE,
				BasePoints:      550,
				DicePerLevel:    4.5,
				ImplicitTargetA: TARGET_UNIT_TARGET_ENEMY,
			},
		},
	})

	sm.AddSpell(&SpellInfo{
		ID:             SPELL_MULTI_SHOT,
		Name:           "多重射击",
		Description:    "同时射击多个目标",
		CastTime:       0, // 即时技能
		Cooldown:       10 * time.Second,
		ManaCost:       0,
		Range:          35.0,
		SchoolMask:     SPELL_SCHOOL_NORMAL,
		TargetType:     TARGET_UNIT_TARGET_ENEMY,
		BaseDamage:     280,
		DamageVariance: 0.2,
		Level:          18,
		Effects: []SpellEffect{
			{
				EffectType:   SPELL_EFFECT_WEAPON_DAMAGE,
				BasePoints:   280,
				DicePerLevel: 2.8,
				RadiusIndex:  8, // 影响范围内多个目标
			},
		},
	})

	fmt.Printf("法术管理器初始化完成，加载了 %d 个法术\n", len(sm.spells))
}





// AddSpell 添加法术
func (sm *SpellManager) AddSpell(spell *SpellInfo) {
	sm.spells[spell.ID] = spell
}

// GetSpell 获取法术信息
func (sm *SpellManager) GetSpell(spellId uint32) *SpellInfo {
	return sm.spells[spellId]
}

// NewSpell 创建法术实例 - 基于AzerothCore的Spell构造函数
func NewSpell(caster IUnit, spellInfo *SpellInfo, world *World) *Spell {
	return &Spell{
		info:        spellInfo,
		caster:      caster,
		targets:     make([]IUnit, 0),
		state:       SPELL_STATE_NULL,
		castTime:    spellInfo.CastTime,
		channelTime: spellInfo.ChannelTime,
		startTime:   time.Now(),
		world:       world,
	}
}

// Prepare 准备法术 - 基于AzerothCore的Spell::prepare
func (s *Spell) Prepare(target IUnit) bool {
	// 检查施法条件
	if !s.checkCast(target) {
		return false
	}

	// 添加目标
	if target != nil {
		s.targets = append(s.targets, target)
	}

	// 计算伤害/治疗
	s.calculateDamage()

	// 消耗法力/能量
	s.takePower()

	// 设置施法者状态
	s.caster.AddUnitState(UNIT_STATE_CASTING)

	if s.info.CastTime > 0 {
		// 有施法时间的法术
		s.state = SPELL_STATE_PREPARING
		fmt.Printf("%s 开始施放 %s (施法时间: %.1f秒)\n",
			s.caster.GetName(), s.info.Name, s.info.CastTime.Seconds())

		// 发送施法开始包给所有客户端
		s.sendSpellStart()
	} else if s.info.IsChanneled {
		// 引导法术
		s.state = SPELL_STATE_CASTING
		fmt.Printf("%s 开始引导 %s (引导时间: %.1f秒)\n",
			s.caster.GetName(), s.info.Name, s.info.ChannelTime.Seconds())

		// 立即开始引导效果
		s.sendSpellGo()
		s.startChanneling()
	} else {
		// 即时法术
		s.state = SPELL_STATE_FINISHED
		fmt.Printf("%s 施放 %s (即时法术)\n",
			s.caster.GetName(), s.info.Name)

		// 立即执行效果
		s.cast()
	}

	return true
}

// checkCast 检查施法条件 - 基于AzerothCore的Spell::CheckCast
func (s *Spell) checkCast(target IUnit) bool {
	// 检查施法者是否存活
	if !s.caster.IsAlive() {
		fmt.Printf("%s 已死亡，无法施法\n", s.caster.GetName())
		return false
	}

	// 检查法力/能量
	if !s.checkPower() {
		fmt.Printf("%s 能量不足，无法施放 %s\n", s.caster.GetName(), s.info.Name)
		return false
	}

	// 检查目标
	if target != nil && !s.isValidTarget(target) {
		fmt.Printf("无效的目标\n")
		return false
	}

	// 检查距离
	if target != nil && s.info.Range > 0 {
		distance := s.caster.GetDistanceTo(target)
		if distance > s.info.Range {
			fmt.Printf("目标距离过远 (%.1f > %.1f)\n", distance, s.info.Range)
			return false
		}
	}

	return true
}

// checkPower 检查能量消耗
func (s *Spell) checkPower() bool {
	if s.info.ManaCost == 0 {
		return true
	}

	// 根据职业检查不同的能量类型
	if player, ok := s.caster.(*Player); ok {
		switch player.class {
		case CLASS_WARRIOR:
			// 战士技能通常消耗怒气，这里简化处理
			return s.caster.GetPower(POWER_RAGE) >= s.info.ManaCost
		case CLASS_HUNTER:
			// 猎人技能消耗集中值
			return s.caster.GetPower(POWER_FOCUS) >= s.info.ManaCost
		case CLASS_ROGUE:
			// 盗贼技能消耗能量
			return s.caster.GetPower(POWER_ENERGY) >= s.info.ManaCost
		default:
			// 其他职业消耗法力
			return s.caster.GetPower(POWER_MANA) >= s.info.ManaCost
		}
	}

	return s.caster.GetPower(POWER_MANA) >= s.info.ManaCost
}

// takePower 消耗能量
func (s *Spell) takePower() {
	if s.info.ManaCost == 0 {
		return
	}

	if player, ok := s.caster.(*Player); ok {
		switch player.class {
		case CLASS_WARRIOR:
			s.caster.ModifyPower(POWER_RAGE, -int32(s.info.ManaCost))
		case CLASS_HUNTER:
			s.caster.ModifyPower(POWER_FOCUS, -int32(s.info.ManaCost))
		case CLASS_ROGUE:
			s.caster.ModifyPower(POWER_ENERGY, -int32(s.info.ManaCost))
		default:
			s.caster.ModifyPower(POWER_MANA, -int32(s.info.ManaCost))
		}
	} else {
		s.caster.ModifyPower(POWER_MANA, -int32(s.info.ManaCost))
	}
}

// isValidTarget 检查目标有效性
func (s *Spell) isValidTarget(target IUnit) bool {
	if target == nil {
		return false
	}

	switch s.info.TargetType {
	case TARGET_UNIT_TARGET_ENEMY:
		// 敌方目标
		return s.caster.IsValidAttackTarget(target)
	case TARGET_UNIT_TARGET_ALLY:
		// 友方目标 - 简化处理，同类型为友方
		if _, ok := s.caster.(*Player); ok {
			if _, ok := target.(*Player); ok {
				return true // 玩家之间可以互相治疗
			}
		}
		return target == s.caster // 可以对自己施法
	case TARGET_UNIT_CASTER:
		// 施法者自己
		return target == s.caster
	}

	return true
}

// calculateDamage 计算伤害/治疗 - 基于AzerothCore的伤害计算
func (s *Spell) calculateDamage() {
	baseDamage := float32(s.info.BaseDamage)

	// 等级加成
	if player, ok := s.caster.(*Player); ok {
		levelBonus := float32(player.level-s.info.Level) * 2.0
		if levelBonus > 0 {
			baseDamage += levelBonus
		}
	}

	// 随机浮动
	variance := baseDamage * s.info.DamageVariance
	finalDamage := baseDamage + (rand.Float32()-0.5)*2*variance

	if finalDamage < 1 {
		finalDamage = 1
	}

	// 根据法术效果类型设置
	for _, effect := range s.info.Effects {
		switch effect.EffectType {
		case SPELL_EFFECT_SCHOOL_DAMAGE, SPELL_EFFECT_WEAPON_DAMAGE:
			s.damage = uint32(finalDamage)
		case SPELL_EFFECT_HEAL:
			s.healing = uint32(finalDamage)
		}
	}
}

// Update 更新法术状态 - 基于AzerothCore的Spell::update
func (s *Spell) Update(diff uint32) bool {
	switch s.state {
	case SPELL_STATE_PREPARING:
		// 施法时间倒计时
		s.castTime -= time.Duration(diff) * time.Millisecond
		if s.castTime <= 0 {
			// 施法完成
			s.state = SPELL_STATE_FINISHED
			s.cast()
			return false // 法术完成，可以移除
		}
		return true

	case SPELL_STATE_CASTING:
		// 引导法术更新
		if s.info.IsChanneled {
			s.channelTime -= time.Duration(diff) * time.Millisecond
			if s.channelTime <= 0 {
				// 引导完成
				s.finishChanneling()
				return false
			}
			// 每秒触发一次效果
			if int(time.Since(s.startTime).Milliseconds())%1000 < int(diff) {
				s.applyChannelEffect()
			}
		}
		return true

	case SPELL_STATE_FINISHED:
		return false // 法术已完成

	case SPELL_STATE_DELAYED:
		// 处理延迟状态
		return true
	}

	return false
}

// cast 执行法术效果 - 基于AzerothCore的Spell::cast
func (s *Spell) cast() {
	// 清除施法状态
	s.caster.ClearUnitState(UNIT_STATE_CASTING)

	// 发送法术生效包
	s.sendSpellGo()

	// 应用法术效果
	s.applyEffects()

	fmt.Printf("%s 完成施放 %s\n", s.caster.GetName(), s.info.Name)
}

// applyEffects 应用法术效果
func (s *Spell) applyEffects() {
	for _, target := range s.targets {
		if target == nil || !target.IsAlive() {
			continue
		}

		for _, effect := range s.info.Effects {
			s.applyEffect(target, &effect)
		}
	}
}

// applyEffect 应用单个效果
func (s *Spell) applyEffect(target IUnit, effect *SpellEffect) {
	switch effect.EffectType {
	case SPELL_EFFECT_SCHOOL_DAMAGE, SPELL_EFFECT_WEAPON_DAMAGE:
		// 造成伤害
		actualDamage := target.DealDamage(s.caster, s.damage, SPELL_DIRECT_DAMAGE, s.info.SchoolMask)
		fmt.Printf("%s 对 %s 造成 %d 点%s伤害\n",
			s.caster.GetName(), target.GetName(), actualDamage, s.getSchoolName())

	case SPELL_EFFECT_HEAL:
		// 治疗效果
		target.Heal(s.caster, s.healing)

	case SPELL_EFFECT_DUMMY:
		// 特殊效果处理
		s.handleDummyEffect(target, effect)
	}
}

// handleDummyEffect 处理特殊效果
func (s *Spell) handleDummyEffect(target IUnit, effect *SpellEffect) {
	switch s.info.ID {
	case SPELL_TAUNT:
		// 嘲讽效果
		if target.GetAI() != nil {
			target.SetVictim(s.caster)
			fmt.Printf("%s 被 %s 嘲讽了\n", target.GetName(), s.caster.GetName())
		}

	case SPELL_POWER_WORD_SHIELD:
		// 护盾效果 - 简化实现
		fmt.Printf("%s 获得了 %d 点护盾保护\n", target.GetName(), s.healing)
	}
}

// startChanneling 开始引导
func (s *Spell) startChanneling() {
	fmt.Printf("%s 开始引导 %s\n", s.caster.GetName(), s.info.Name)
}

// applyChannelEffect 应用引导效果
func (s *Spell) applyChannelEffect() {
	// 暴风雪等引导法术的持续效果
	if s.info.ID == SPELL_BLIZZARD {
		for _, target := range s.targets {
			if target != nil && target.IsAlive() {
				damage := s.damage / 8 // 分8次造成伤害
				actualDamage := target.DealDamage(s.caster, damage, SPELL_DIRECT_DAMAGE, s.info.SchoolMask)
				fmt.Printf("暴风雪对 %s 造成 %d 点冰霜伤害\n", target.GetName(), actualDamage)
			}
		}
	}
}

// finishChanneling 完成引导
func (s *Spell) finishChanneling() {
	s.caster.ClearUnitState(UNIT_STATE_CASTING)
	fmt.Printf("%s 完成引导 %s\n", s.caster.GetName(), s.info.Name)
}

// interrupt 打断法术 - 基于AzerothCore的Spell::cancel
func (s *Spell) Interrupt() {
	if s.state == SPELL_STATE_PREPARING || s.state == SPELL_STATE_CASTING {
		s.interrupted = true
		s.state = SPELL_STATE_FINISHED
		s.caster.ClearUnitState(UNIT_STATE_CASTING)
		fmt.Printf("%s 的 %s 被打断了\n", s.caster.GetName(), s.info.Name)
	}
}

// getSchoolName 获取法术学派名称
func (s *Spell) getSchoolName() string {
	switch s.info.SchoolMask {
	case SPELL_SCHOOL_FIRE:
		return "火焰"
	case SPELL_SCHOOL_FROST:
		return "冰霜"
	case SPELL_SCHOOL_SHADOW:
		return "暗影"
	case SPELL_SCHOOL_HOLY:
		return "神圣"
	case SPELL_SCHOOL_NATURE:
		return "自然"
	case SPELL_SCHOOL_ARCANE:
		return "奥术"
	default:
		return "物理"
	}
}

// sendSpellStart 发送法术开始包 - 基于AzerothCore的SMSG_SPELL_START
func (s *Spell) sendSpellStart() {
	if s.world == nil {
		return
	}

	// 向所有相关客户端发送法术开始包
	s.world.BroadcastSpellStart(s.caster, s.info.ID, s.targets, s.info.CastTime)
}

// sendSpellGo 发送法术生效包 - 基于AzerothCore的SMSG_SPELL_GO
func (s *Spell) sendSpellGo() {
	if s.world == nil {
		return
	}

	// 向所有相关客户端发送法术生效包
	s.world.BroadcastSpellGo(s.caster, s.info.ID, s.targets)
}
