package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 副本类型
const (
	DUNGEON_DEADMINES = 1
)

// 副本难度
const (
	DIFFICULTY_NORMAL = 0
	DIFFICULTY_HEROIC = 1
)

// 副本特定常量

// 副本结构
type Dungeon struct {
	id         uint32
	name       string
	difficulty uint8
	minLevel   uint8
	maxLevel   uint8
	maxPlayers uint8
	encounters []*Encounter
	trash      []*TrashGroup
	players    []*Player
	world      *World
}

// 遭遇战结构
type Encounter struct {
	id       uint32
	name     string
	boss     *Creature
	adds     []*Creature
	phase    uint8
	maxPhase uint8
	isActive bool
}

// 小怪组结构
type TrashGroup struct {
	id        uint32
	creatures []*Creature
	isCleared bool
}

// 创建死亡矿井副本
func NewDeadminesDungeon(world *World) *Dungeon {
	dungeon := &Dungeon{
		id:         DUNGEON_DEADMINES,
		name:       "死亡矿井",
		difficulty: DIFFICULTY_NORMAL,
		minLevel:   15,
		maxLevel:   25,
		maxPlayers: 5,
		world:      world,
	}

	// 创建小怪组
	dungeon.createTrashGroups()

	// 创建BOSS遭遇战
	dungeon.createEncounters()

	return dungeon
}

// 创建小怪组
func (d *Dungeon) createTrashGroups() {
	// 第一组小怪：矿工
	group1 := &TrashGroup{
		id: 1,
		creatures: []*Creature{
			d.createDefiasMiner("迪菲亚矿工", 18),
			d.createDefiasMiner("迪菲亚矿工", 18),
			d.createDefiasOverseer("迪菲亚监工", 20),
		},
	}

	// 第二组小怪：盗贼
	group2 := &TrashGroup{
		id: 2,
		creatures: []*Creature{
			d.createDefiasThug("迪菲亚暴徒", 19),
			d.createDefiasThug("迪菲亚暴徒", 19),
			d.createDefiasConjurer("迪菲亚咒术师", 20),
		},
	}

	// 第三组小怪：精英守卫
	group3 := &TrashGroup{
		id: 3,
		creatures: []*Creature{
			d.createDefiasElite("迪菲亚精英", 22),
			d.createDefiasElite("迪菲亚精英", 22),
		},
	}

	d.trash = []*TrashGroup{group1, group2, group3}
}

// 创建BOSS遭遇战
func (d *Dungeon) createEncounters() {
	// 范克里夫
	vancleef := d.createVanCleef()
	encounter1 := &Encounter{
		id:       1,
		name:     "埃德温·范克里夫",
		boss:     vancleef,
		adds:     []*Creature{},
		phase:    1,
		maxPhase: 3,
		isActive: false,
	}

	d.encounters = []*Encounter{encounter1}
}

// 创建迪菲亚矿工
func (d *Dungeon) createDefiasMiner(name string, level uint8) *Creature {
	miner := NewCreature(name, level, CREATURE_TYPE_HUMANOID)
	miner.SetMaxHealth(1200)
	miner.SetHealth(1200)
	miner.SetMaxPower(POWER_MANA, 800)
	miner.SetPower(POWER_MANA, 800)

	// 设置基础属性（简化处理）

	// 设置专门的矿工AI
	miner.SetAI(NewMinerAI(miner))

	return miner
}

// 创建迪菲亚监工
func (d *Dungeon) createDefiasOverseer(name string, level uint8) *Creature {
	overseer := NewCreature(name, level, CREATURE_TYPE_HUMANOID)
	overseer.SetMaxHealth(1800)
	overseer.SetHealth(1800)
	overseer.SetMaxPower(POWER_MANA, 1200)
	overseer.SetPower(POWER_MANA, 1200)

	// 设置基础属性（简化处理）

	overseer.SetAI(NewOverseerAI(overseer))

	return overseer
}

// 创建迪菲亚暴徒
func (d *Dungeon) createDefiasThug(name string, level uint8) *Creature {
	thug := NewCreature(name, level, CREATURE_TYPE_HUMANOID)
	thug.SetMaxHealth(1500)
	thug.SetHealth(1500)
	thug.SetMaxPower(POWER_ENERGY, 100)
	thug.SetPower(POWER_ENERGY, 100)

	// 设置基础属性（简化处理）

	thug.SetAI(NewThugAI(thug))

	return thug
}

// 创建迪菲亚咒术师
func (d *Dungeon) createDefiasConjurer(name string, level uint8) *Creature {
	conjurer := NewCreature(name, level, CREATURE_TYPE_HUMANOID)
	conjurer.SetMaxHealth(1000)
	conjurer.SetHealth(1000)
	conjurer.SetMaxPower(POWER_MANA, 2000)
	conjurer.SetPower(POWER_MANA, 2000)

	// 设置基础属性（简化处理）

	conjurer.SetAI(NewConjurerAI(conjurer))

	return conjurer
}

// 创建迪菲亚精英
func (d *Dungeon) createDefiasElite(name string, level uint8) *Creature {
	elite := NewCreature(name, level, CREATURE_TYPE_HUMANOID)
	elite.SetMaxHealth(2500)
	elite.SetHealth(2500)
	elite.SetMaxPower(POWER_RAGE, 100)
	elite.SetPower(POWER_RAGE, 0)

	// 设置基础属性（简化处理）

	elite.SetAI(NewEliteAI(elite))

	return elite
}

// 创建范克里夫BOSS
func (d *Dungeon) createVanCleef() *Creature {
	vancleef := NewCreature("埃德温·范克里夫", 26, CREATURE_TYPE_HUMANOID)
	vancleef.SetMaxHealth(8000)
	vancleef.SetHealth(8000)
	vancleef.SetMaxPower(POWER_ENERGY, 100)
	vancleef.SetPower(POWER_ENERGY, 100)

	// BOSS级别属性（简化处理）

	vancleef.SetAI(NewVanCleefAI(vancleef))

	return vancleef
}

// 添加玩家到副本
func (d *Dungeon) AddPlayer(player *Player) bool {
	if len(d.players) >= int(d.maxPlayers) {
		fmt.Printf("副本 %s 已满员\n", d.name)
		return false
	}

	if player.GetLevel() < d.minLevel {
		fmt.Printf("玩家 %s 等级过低，无法进入副本 %s\n", player.GetName(), d.name)
		return false
	}

	d.players = append(d.players, player)
	fmt.Printf("玩家 %s 进入副本 %s\n", player.GetName(), d.name)
	return true
}

// 开始副本
func (d *Dungeon) Start() {
	fmt.Printf("\n=== 副本 %s 开始 ===\n", d.name)
	fmt.Printf("参与玩家: ")
	for i, player := range d.players {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s(%s)", player.GetName(), d.getClassName(player.GetClass()))
	}
	fmt.Println()

	// 清理小怪组
	for i, group := range d.trash {
		fmt.Printf("\n--- 第%d组小怪 ---\n", i+1)
		d.fightTrashGroup(group)
		if !d.allPlayersAlive() {
			fmt.Println("团队全灭，副本失败！")
			return
		}

		// 战斗间隙恢复
		d.restorePlayers()
		time.Sleep(1 * time.Second)
	}

	// BOSS战
	for _, encounter := range d.encounters {
		fmt.Printf("\n=== BOSS战：%s ===\n", encounter.name)
		d.fightBoss(encounter)
		if !d.allPlayersAlive() {
			fmt.Println("团队全灭，副本失败！")
			return
		}
	}

	fmt.Printf("\n🎉 恭喜！副本 %s 通关成功！\n", d.name)
}

// 战斗小怪组
func (d *Dungeon) fightTrashGroup(group *TrashGroup) {
	fmt.Printf("遭遇小怪组：")
	for i, creature := range group.creatures {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(creature.GetName())
	}
	fmt.Println()

	// 开始战斗
	for _, creature := range group.creatures {
		d.world.AddUnit(creature)
	}

	// 玩家开始攻击
	tank := d.getTank()
	if tank != nil {
		// 坦克拉怪
		for _, creature := range group.creatures {
			tank.Attack(creature)
		}
	}

	// 战斗循环
	combatTime := 0
	maxCombatTime := 30000 // 30秒超时

	for d.hasAliveEnemies(group.creatures) && d.allPlayersAlive() && combatTime < maxCombatTime {
		// 更新所有单位
		for _, player := range d.players {
			if player.IsAlive() {
				player.Update(100)
			}
		}

		for _, creature := range group.creatures {
			if creature.IsAlive() {
				creature.Update(100)
			}
		}

		combatTime += 100
		time.Sleep(200 * time.Millisecond)

		// 每5秒显示一次战斗状态
		if combatTime%5000 == 0 {
			aliveEnemies := 0
			for _, creature := range group.creatures {
				if creature.IsAlive() {
					aliveEnemies++
				}
			}
			fmt.Printf("战斗进行中... 剩余敌人: %d\n", aliveEnemies)
		}
	}

	if combatTime >= maxCombatTime {
		fmt.Println("战斗超时！")
	}

	group.isCleared = true
	fmt.Printf("小怪组清理完成！\n")
}

// BOSS战
func (d *Dungeon) fightBoss(encounter *Encounter) {
	boss := encounter.boss
	d.world.AddUnit(boss)

	fmt.Printf("BOSS %s 出现！生命值：%d/%d\n",
		boss.GetName(), boss.GetHealth(), boss.GetMaxHealth())

	// 坦克开怪
	tank := d.getTank()
	if tank != nil {
		tank.Attack(boss)
	}

	encounter.isActive = true

	// BOSS战循环
	combatTime := 0
	maxCombatTime := 60000 // 60秒超时

	for boss.IsAlive() && d.allPlayersAlive() && combatTime < maxCombatTime {
		// 检查阶段转换
		d.checkPhaseTransition(encounter)

		// 更新所有单位
		for _, player := range d.players {
			if player.IsAlive() {
				player.Update(100)
			}
		}

		boss.Update(100)

		// 更新小怪
		for _, add := range encounter.adds {
			if add.IsAlive() {
				add.Update(100)
			}
		}

		combatTime += 100
		time.Sleep(200 * time.Millisecond)

		// 每10秒显示一次BOSS状态
		if combatTime%10000 == 0 {
			healthPercent := float64(boss.GetHealth()) / float64(boss.GetMaxHealth()) * 100
			fmt.Printf("BOSS %s 生命值: %.1f%%\n", boss.GetName(), healthPercent)
		}
	}

	if combatTime >= maxCombatTime {
		fmt.Println("BOSS战超时！")
	}

	encounter.isActive = false
	fmt.Printf("🎉 BOSS %s 被击败！\n", boss.GetName())
}

// 检查阶段转换
func (d *Dungeon) checkPhaseTransition(encounter *Encounter) {
	boss := encounter.boss
	healthPercent := float64(boss.GetHealth()) / float64(boss.GetMaxHealth()) * 100

	// 范克里夫的阶段转换
	if encounter.name == "埃德温·范克里夫" {
		if encounter.phase == 1 && healthPercent <= 66 {
			encounter.phase = 2
			fmt.Printf("🔥 %s 进入第二阶段！召唤小弟！\n", boss.GetName())
			d.spawnVanCleefAdds(encounter)
		} else if encounter.phase == 2 && healthPercent <= 33 {
			encounter.phase = 3
			fmt.Printf("⚡ %s 进入第三阶段！狂暴状态！\n", boss.GetName())
			// 增加攻击力和攻击速度
		}
	}
}

// 召唤范克里夫的小弟
func (d *Dungeon) spawnVanCleefAdds(encounter *Encounter) {
	add1 := d.createDefiasThug("迪菲亚保镖", 24)
	add2 := d.createDefiasThug("迪菲亚保镖", 24)

	encounter.adds = append(encounter.adds, add1, add2)
	d.world.AddUnit(add1)
	d.world.AddUnit(add2)

	// 小弟攻击随机玩家
	if len(d.players) > 0 {
		target1 := d.players[rand.Intn(len(d.players))]
		target2 := d.players[rand.Intn(len(d.players))]
		add1.Attack(target1)
		add2.Attack(target2)
	}
}

// 获取坦克
func (d *Dungeon) getTank() *Player {
	for _, player := range d.players {
		if player.GetClass() == CLASS_WARRIOR && player.IsAlive() {
			return player
		}
	}
	return nil
}

// 检查是否还有存活的敌人
func (d *Dungeon) hasAliveEnemies(creatures []*Creature) bool {
	for _, creature := range creatures {
		if creature.IsAlive() {
			return true
		}
	}
	return false
}

// 检查所有玩家是否存活
func (d *Dungeon) allPlayersAlive() bool {
	for _, player := range d.players {
		if player.IsAlive() {
			return true
		}
	}
	return false
}

// 恢复玩家状态
func (d *Dungeon) restorePlayers() {
	fmt.Println("战斗间隙，队伍休整...")
	for _, player := range d.players {
		if player.IsAlive() {
			// 恢复生命值
			player.SetHealth(player.GetMaxHealth())
			// 恢复能量
			switch player.GetClass() {
			case CLASS_WARRIOR:
				player.SetPower(POWER_RAGE, 0)
			case CLASS_MAGE, CLASS_PRIEST, CLASS_WARLOCK, CLASS_HUNTER:
				player.SetPower(POWER_MANA, player.GetMaxPower(POWER_MANA))
			case CLASS_ROGUE:
				player.SetPower(POWER_ENERGY, player.GetMaxPower(POWER_ENERGY))
			}
		}
	}
}

// 获取职业名称
func (d *Dungeon) getClassName(class uint8) string {
	switch class {
	case CLASS_WARRIOR:
		return "战士"
	case CLASS_MAGE:
		return "法师"
	case CLASS_PRIEST:
		return "牧师"
	case CLASS_HUNTER:
		return "猎人"
	case CLASS_WARLOCK:
		return "术士"
	case CLASS_ROGUE:
		return "盗贼"
	default:
		return "未知"
	}
}
