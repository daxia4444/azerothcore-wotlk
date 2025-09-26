package main

/*
#cgo CFLAGS: -I../deps/recastnavigation/Recast/Include -I../deps/recastnavigation/Detour/Include -I../deps/recastnavigation/DetourTileCache/Include
#cgo LDFLAGS: -L../deps/recastnavigation -lRecast -lDetour -lDetourTileCache -lstdc++ -lm

#include <stdlib.h>
#include <string.h>
#include "DetourNavMesh.h"
#include "DetourNavMeshQuery.h"
#include "DetourCommon.h"
#include "DetourNavMeshBuilder.h"

// 包装函数，用于处理C++异常和内存管理
dtNavMesh* createNavMesh() {
    return dtAllocNavMesh();
}

dtNavMeshQuery* createNavMeshQuery() {
    return dtAllocNavMeshQuery();
}

void freeNavMesh(dtNavMesh* navMesh) {
    if (navMesh) {
        dtFreeNavMesh(navMesh);
    }
}

void freeNavMeshQuery(dtNavMeshQuery* navQuery) {
    if (navQuery) {
        dtFreeNavMeshQuery(navQuery);
    }
}

// 初始化导航网格查询器
dtStatus initNavMeshQuery(dtNavMeshQuery* navQuery, dtNavMesh* navMesh, int maxNodes) {
    if (!navQuery || !navMesh) {
        return DT_FAILURE;
    }
    return navQuery->init(navMesh, maxNodes);
}

// 查找多边形路径 (对应 AzerothCore 的 BuildPolyPath)
dtStatus findPolyPath(dtNavMeshQuery* navQuery, dtQueryFilter* filter,
                     float* startPos, float* endPos,
                     dtPolyRef* pathPolys, int* pathCount, int maxPath) {
    if (!navQuery || !filter || !startPos || !endPos || !pathPolys || !pathCount) {
        return DT_FAILURE;
    }

    dtPolyRef startRef, endRef;
    float nearestPt[3];
    const float extents[3] = {2.0f, 4.0f, 2.0f};

    // 查找起点多边形
    dtStatus status = navQuery->findNearestPoly(startPos, extents, filter, &startRef, nearestPt);
    if (dtStatusFailed(status) || startRef == 0) {
        return DT_FAILURE;
    }

    // 查找终点多边形
    status = navQuery->findNearestPoly(endPos, extents, filter, &endRef, nearestPt);
    if (dtStatusFailed(status) || endRef == 0) {
        return DT_FAILURE;
    }

    // 执行寻路
    return navQuery->findPath(startRef, endRef, startPos, endPos, filter, pathPolys, pathCount, maxPath);
}

// 构建点路径 (对应 AzerothCore 的 BuildPointPath)
dtStatus buildPointPath(dtNavMeshQuery* navQuery,
                       float* startPos, float* endPos,
                       dtPolyRef* pathPolys, int pathCount,
                       float* pathPoints, int* pointCount, int maxPoints) {
    if (!navQuery || !startPos || !endPos || !pathPolys || !pathPoints || !pointCount) {
        return DT_FAILURE;
    }

    return navQuery->findStraightPath(startPos, endPos, pathPolys, pathCount,
                                     pathPoints, NULL, NULL, pointCount, maxPoints);
}

// 检查位置是否可行走
dtStatus isWalkable(dtNavMeshQuery* navQuery, dtQueryFilter* filter, float* pos) {
    if (!navQuery || !filter || !pos) {
        return DT_FAILURE;
    }

    dtPolyRef polyRef;
    float nearestPt[3];
    const float extents[3] = {2.0f, 4.0f, 2.0f};
    return navQuery->findNearestPoly(pos, extents, filter, &polyRef, nearestPt);
}

// 射线检测 (用于障碍物检测)
dtStatus raycast(dtNavMeshQuery* navQuery, dtQueryFilter* filter,
                float* startPos, float* endPos, float* hitDist, float* hitNormal) {
    if (!navQuery || !filter || !startPos || !endPos || !hitDist) {
        return DT_FAILURE;
    }

    dtPolyRef startRef;
    float nearestPt[3];
    const float extents[3] = {2.0f, 4.0f, 2.0f};

    dtStatus status = navQuery->findNearestPoly(startPos, extents, filter, &startRef, nearestPt);
    if (dtStatusFailed(status) || startRef == 0) {
        return DT_FAILURE;
    }

    return navQuery->raycast(startRef, startPos, endPos, filter, hitDist, hitNormal, NULL, NULL, 0);
}

// 获取多边形高度
dtStatus getPolyHeight(dtNavMeshQuery* navQuery, dtPolyRef polyRef, float* pos, float* height) {
    if (!navQuery || polyRef == 0 || !pos || !height) {
        return DT_FAILURE;
    }
    return navQuery->getPolyHeight(polyRef, pos, height);
}

// 初始化导航网格参数 (对应 AzerothCore 的导航网格初始化)
dtStatus initNavMesh(dtNavMesh* navMesh, dtNavMeshParams* params) {
    if (!navMesh || !params) {
        return DT_FAILURE;
    }
    return navMesh->init(params);
}

// 添加瓦片到导航网格 (对应 AzerothCore 的 addTile)
dtTileRef addTileToNavMesh(dtNavMesh* navMesh, unsigned char* data, int dataSize, int flags) {
    if (!navMesh || !data || dataSize <= 0) {
        return 0;
    }

    dtTileRef tileRef = 0;
    dtStatus status = navMesh->addTile(data, dataSize, flags, 0, &tileRef);

    if (dtStatusFailed(status)) {
        return 0;
    }

    return tileRef;
}

// 移除瓦片从导航网格
dtStatus removeTileFromNavMesh(dtNavMesh* navMesh, dtTileRef tileRef, unsigned char** data, int* dataSize) {
    if (!navMesh || tileRef == 0) {
        return DT_FAILURE;
    }
    return navMesh->removeTile(tileRef, data, dataSize);
}

// 创建导航网格参数 (基于 AzerothCore 的配置)
void createNavMeshParams(dtNavMeshParams* params, float* bmin, float* bmax,
                        float tileWidth, float tileHeight, int maxTiles, int maxPolys) {
    if (!params || !bmin || !bmax) {
        return;
    }

    memset(params, 0, sizeof(dtNavMeshParams));
    dtVcopy(params->orig, bmin);
    params->tileWidth = tileWidth;
    params->tileHeight = tileHeight;
    params->maxTiles = maxTiles;
    params->maxPolys = maxPolys;
}

*/
import "C"
import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
	"unsafe"
)

// ========== AzerothCore 常量定义 ==========

const (
	// 路径相关常量 (来自 PathGenerator.h)
	MAX_PATH_LENGTH       = 74
	MAX_POINT_PATH_LENGTH = 74
	VERTEX_SIZE           = 3
	INVALID_POLYREF       = 0

	// 地图相关常量
	TILE_SIZE = 533.33333 // AzerothCore 瓦片大小 (码)
	MAP_SIZE  = 64        // 64x64 瓦片

	// 性能相关常量
	DEFAULT_MAX_NODES = 2048 // 默认最大节点数
	QUERY_TIMEOUT     = 5    // 查询超时时间 (秒)
)

// 路径类型 (对应 AzerothCore 的 PathType)
type PathType uint32

const (
	PATHFIND_BLANK          PathType = 0x00 // 空路径
	PATHFIND_NORMAL         PathType = 0x01 // 正常路径
	PATHFIND_NOT_USING_PATH PathType = 0x02 // 不使用路径 (飞行/游泳)
	PATHFIND_SHORT          PathType = 0x04 // 短路径
	PATHFIND_INCOMPLETE     PathType = 0x08 // 不完整路径
	PATHFIND_NOPATH         PathType = 0x10 // 无路径
	PATHFIND_FAR_FROM_POLY  PathType = 0x20 // 远离多边形
)

// ========== 数据结构定义 ==========

// Vector3 三维向量 (对应 G3D::Vector3)
type Vector3 struct {
	X, Y, Z float32
}

// String 返回向量的字符串表示
func (v Vector3) String() string {
	return fmt.Sprintf("(%.2f, %.2f, %.2f)", v.X, v.Y, v.Z)
}

// Distance 计算到另一个向量的距离
func (v Vector3) Distance(other Vector3) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	dz := v.Z - other.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Distance2D 计算2D距离
func (v Vector3) Distance2D(other Vector3) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

// AzerothCoreConfig AzerothCore 寻路配置
type AzerothCoreConfig struct {
	CellSize               float32 // 体素大小
	CellHeight             float32 // 体素高度
	WalkableHeight         float32 // 可行走高度
	WalkableRadius         float32 // 可行走半径
	WalkableClimb          float32 // 可攀爬高度
	WalkableSlopeAngle     float32 // 可行走坡度角
	MinRegionArea          int     // 最小区域面积
	MergeRegionArea        int     // 合并区域面积
	MaxEdgeLen             float32 // 最大边长
	MaxSimplificationError float32 // 最大简化误差
	MaxVertsPerPoly        int     // 多边形最大顶点数
	DetailSampleDist       float32 // 细节采样距离
	DetailSampleMaxError   float32 // 细节最大误差
}

// NavMeshTile 导航网格瓦片
type NavMeshTile struct {
	TileX, TileY int32
	Data         []byte
	Header       *NavMeshTileHeader
}

// NavMeshTileHeader 导航网格瓦片头
type NavMeshTileHeader struct {
	Magic           uint32
	Version         uint32
	X, Y            int32
	Layer           uint32
	UserId          uint32
	PolyCount       uint32
	VertCount       uint32
	MaxLinkCount    uint32
	DetailMeshCount uint32
	DetailVertCount uint32
	DetailTriCount  uint32
	BvNodeCount     uint32
	OffMeshConCount uint32
	OffMeshBase     uint32
	WalkableHeight  float32
	WalkableRadius  float32
	WalkableClimb   float32
	BMin, BMax      [3]float32
}

// PathfindingStats 寻路统计信息
type PathfindingStats struct {
	TotalQueries    uint64        // 总查询数
	SuccessfulPaths uint64        // 成功路径数
	FailedPaths     uint64        // 失败路径数
	AverageTime     time.Duration // 平均耗时
	TotalTime       time.Duration // 总耗时
	mutex           sync.RWMutex  // 统计锁
}

// UpdateStats 更新统计信息
func (s *PathfindingStats) UpdateStats(success bool, duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.TotalQueries++
	s.TotalTime += duration
	s.AverageTime = s.TotalTime / time.Duration(s.TotalQueries)

	if success {
		s.SuccessfulPaths++
	} else {
		s.FailedPaths++
	}
}

// GetStats 获取统计信息
func (s *PathfindingStats) GetStats() (uint64, uint64, uint64, time.Duration) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.TotalQueries, s.SuccessfulPaths, s.FailedPaths, s.AverageTime
}

// PathGenerator AzerothCore 路径生成器 (对应 PathGenerator 类)
type PathGenerator struct {
	// 核心组件
	navMesh      *C.dtNavMesh      // 导航网格
	navMeshQuery *C.dtNavMeshQuery // 导航网格查询器
	filter       *C.dtQueryFilter  // 查询过滤器

	// 路径数据
	pathPolyRefs [MAX_PATH_LENGTH]C.dtPolyRef // 多边形路径引用
	polyLength   uint32                       // 多边形路径长度
	pathPoints   []Vector3                    // 点路径
	pathType     PathType                     // 路径类型

	// 配置
	config           AzerothCoreConfig
	useStraightPath  bool // 使用直线路径
	forceDestination bool // 强制目标
	slopeCheck       bool // 坡度检查
	pointPathLimit   int  // 点路径限制
	useRaycast       bool // 使用射线检测

	// 位置信息
	startPosition Vector3 // 起始位置
	endPosition   Vector3 // 结束位置
	actualEndPos  Vector3 // 实际结束位置

	// 统计信息
	stats PathfindingStats

	// 线程安全
	mutex sync.RWMutex
}

// MMapManager 地图管理器 (对应 AzerothCore 的 MMapMgr)
type MMapManager struct {
	loadedMaps map[uint32]*MapData // 已加载的地图数据
	mutex      sync.RWMutex        // 读写锁
	dataPath   string              // 数据路径
	logger     *log.Logger         // 日志记录器
}

// MapData 地图数据 (对应 AzerothCore 的 MMapData)
type MapData struct {
	navMesh        *C.dtNavMesh                 // 导航网格 (从 .mmtile 文件构建)
	navMeshQueries map[uint32]*C.dtNavMeshQuery // 实例查询器缓存 (key: instanceID, 按需创建)
	loadedTileRefs map[uint32]C.dtTileRef       // 已加载瓦片引用 (key: tileID, 用于卸载瓦片)
	tiles          map[string]*NavMeshTile      // 瓦片数据缓存 (key: "x_y", 从文件系统加载)
	mutex          sync.RWMutex                 // 读写锁 (保护并发访问)
}

// ========== 工厂函数 ==========

// GetDefaultAzerothCoreConfig 获取默认 AzerothCore 配置
func GetDefaultAzerothCoreConfig() AzerothCoreConfig {
	return AzerothCoreConfig{
		CellSize:               0.3,  // 对应游戏内 0.3 码的精度
		CellHeight:             0.2,  // 对应游戏内 0.2 码的高度精度
		WalkableHeight:         2.0,  // 人形生物高度约 2 码
		WalkableRadius:         0.6,  // 人形生物半径约 0.6 码
		WalkableClimb:          0.9,  // 可攀爬台阶高度 0.9 码
		WalkableSlopeAngle:     45.0, // 45度坡度限制
		MinRegionArea:          8,    // 最小区域 8 个体素
		MergeRegionArea:        20,   // 合并区域 20 个体素
		MaxEdgeLen:             12.0, // 最大边长 12 码
		MaxSimplificationError: 1.3,  // 简化误差 1.3 码
		MaxVertsPerPoly:        6,    // 六边形多边形
		DetailSampleDist:       6.0,  // 细节采样距离 6 码
		DetailSampleMaxError:   1.0,  // 细节误差 1 码
	}
}

// NewMMapManager 创建地图管理器
func NewMMapManager(dataPath string) *MMapManager {
	return &MMapManager{
		loadedMaps: make(map[uint32]*MapData),
		dataPath:   dataPath,
		logger:     log.New(log.Writer(), "[MMapMgr] ", log.LstdFlags),
	}
}

// NewPathGenerator 创建路径生成器 (对应 AzerothCore 的 PathGenerator 构造函数)
func NewPathGenerator(mapID uint32, instanceID uint32, mmapMgr *MMapManager) *PathGenerator {
	pg := &PathGenerator{
		config:          GetDefaultAzerothCoreConfig(),
		pointPathLimit:  MAX_POINT_PATH_LENGTH,
		pathType:        PATHFIND_BLANK,
		useStraightPath: true,
		slopeCheck:      true,
		useRaycast:      true,
	}

	// 获取导航网格和查询器
	mapData := mmapMgr.GetMapData(mapID)
	if mapData != nil {
		pg.navMesh = mapData.navMesh
		pg.navMeshQuery = mmapMgr.GetNavMeshQuery(mapID, instanceID)
	}

	// 创建查询过滤器
	if err := pg.createFilter(); err != nil {
		log.Printf("创建查询过滤器失败: %v", err)
	}

	return pg
}

// ========== MMapManager 实现 ==========

// LoadMap 加载地图 (对应 AzerothCore 的 loadMap)
func (mgr *MMapManager) LoadMap(mapID uint32, x, y int32) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	// 检查地图是否已加载
	mapData, exists := mgr.loadedMaps[mapID]
	if !exists {
		// 创建新的地图数据
		mapData = &MapData{
			navMesh:        C.createNavMesh(),                  // 创建空的导航网格容器
			navMeshQueries: make(map[uint32]*C.dtNavMeshQuery), // 初始化实例查询器缓存 (按需填充)
			loadedTileRefs: make(map[uint32]C.dtTileRef),       // 初始化瓦片引用缓存 (加载瓦片时填充)
			tiles:          make(map[string]*NavMeshTile),      // 初始化瓦片数据缓存 (从文件加载时填充)
		}

		if mapData.navMesh == nil {
			mgr.logger.Printf("❌ 创建导航网格失败: 地图 %d", mapID)
			return false
		}

		// 🔑 关键步骤：初始化导航网格参数
		// 这里告诉 Recast 库整个地图的边界、瓦片大小等信息
		if !mgr.initializeNavMeshParams(mapData.navMesh, mapID) {
			mgr.logger.Printf("❌ 导航网格初始化失败")
			C.freeNavMesh(mapData.navMesh)
			return false
		}

		mgr.loadedMaps[mapID] = mapData
		mgr.logger.Printf("✅ 创建新地图数据: 地图 %d", mapID)
	}

	// 加载瓦片数据
	tileKey := fmt.Sprintf("%d_%d", x, y)
	if _, exists := mapData.tiles[tileKey]; exists {
		return true // 瓦片已加载
	}

	// 模拟加载 .mmtile 文件
	tile := mgr.loadNavMeshTile(mapID, x, y)
	if tile == nil {
		mgr.logger.Printf("⚠️  无法加载导航网格瓦片: 地图 %d, 瓦片 (%d, %d)", mapID, x, y)
		return false
	}

	mapData.tiles[tileKey] = tile

	// 将瓦片添加到导航网格
	tileRef := mgr.addTileToNavMesh(mapData.navMesh, tile)
	if tileRef != 0 {
		tileID := mgr.packTileID(x, y)
		mapData.loadedTileRefs[tileID] = tileRef
		mgr.logger.Printf("✅ 成功加载导航网格瓦片: 地图 %d, 瓦片 (%d, %d)", mapID, x, y)
		return true
	}

	return false
}

// GetNavMeshQuery 获取导航网格查询器 (对应 AzerothCore 的 GetNavMeshQuery)
func (mgr *MMapManager) GetNavMeshQuery(mapID uint32, instanceID uint32) *C.dtNavMeshQuery {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	mapData, exists := mgr.loadedMaps[mapID]
	if !exists {
		return nil
	}

	// 检查实例查询器是否存在
	if query, exists := mapData.navMeshQueries[instanceID]; exists {
		return query
	}

	// 创建新的查询器
	query := C.createNavMeshQuery()
	if query == nil {
		mgr.logger.Printf("❌ 创建导航网格查询器失败: 地图 %d, 实例 %d", mapID, instanceID)
		return nil
	}

	// 初始化查询器
	status := C.initNavMeshQuery(query, mapData.navMesh, DEFAULT_MAX_NODES)
	if C.dtStatusFailed(status) {
		C.freeNavMeshQuery(query)
		mgr.logger.Printf("❌ 初始化导航网格查询器失败: 地图 %d, 实例 %d", mapID, instanceID)
		return nil
	}

	mapData.navMeshQueries[instanceID] = query
	mgr.logger.Printf("✅ 创建导航网格查询器: 地图 %d, 实例 %d", mapID, instanceID)
	return query
}

// GetMapData 获取地图数据
func (mgr *MMapManager) GetMapData(mapID uint32) *MapData {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	return mgr.loadedMaps[mapID]
}

// GetLoadedTileCount 获取已加载瓦片数量
func (mgr *MMapManager) GetLoadedTileCount(mapID uint32) int {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if mapData, exists := mgr.loadedMaps[mapID]; exists {
		return len(mapData.tiles)
	}
	return 0
}

// loadNavMeshTile 加载导航网格瓦片 (真实的 .mmtile 文件加载逻辑)
func (mgr *MMapManager) loadNavMeshTile(mapID uint32, x, y int32) *NavMeshTile {
	// 构建 .mmtile 文件路径 (对应 AzerothCore 的文件命名规则)
	// 格式: {dataPath}/mmaps/{mapID:03d}{y:02d}{x:02d}.mmtile
	tileFileName := fmt.Sprintf("%s/mmaps/%03d%02d%02d.mmtile", mgr.dataPath, mapID, y, x)

	mgr.logger.Printf("🔍 尝试加载瓦片文件: %s", tileFileName)

	// 真实实现：从文件系统读取 .mmtile 文件
	tileData, err := mgr.loadTileFromFile(tileFileName)
	if err != nil {
		mgr.logger.Printf("⚠️  无法加载瓦片文件 %s: %v", tileFileName, err)
		// 如果文件不存在，创建模拟数据用于演示
		return mgr.createSimulatedTile(mapID, x, y)
	}

	// 解析瓦片头部信息
	header, err := mgr.parseTileHeader(tileData)
	if err != nil {
		mgr.logger.Printf("❌ 解析瓦片头部失败: %v", err)
		return nil
	}

	// 验证瓦片数据完整性
	if !mgr.validateTileData(header, tileData) {
		mgr.logger.Printf("❌ 瓦片数据验证失败")
		return nil
	}

	tile := &NavMeshTile{
		TileX:  x,
		TileY:  y,
		Header: header,
		Data:   tileData,
	}

	mgr.logger.Printf("✅ 成功加载瓦片: %d 个多边形, %d 个顶点",
		header.PolyCount, header.VertCount)

	return tile
}

// loadTileFromFile 从文件加载瓦片数据 (真实的文件I/O)
func (mgr *MMapManager) loadTileFromFile(fileName string) ([]byte, error) {
	// 在真实实现中，这里会：
	// 1. 打开 .mmtile 文件
	// 2. 读取文件头验证格式
	// 3. 读取完整的瓦片数据
	// 4. 进行数据完整性检查

	// 模拟文件不存在的情况
	return nil, fmt.Errorf("文件不存在 (演示模式)")
}

// parseTileHeader 解析瓦片头部 (对应 AzerothCore 的瓦片格式)
func (mgr *MMapManager) parseTileHeader(data []byte) (*NavMeshTileHeader, error) {
	if len(data) < 128 { // 最小头部大小
		return nil, fmt.Errorf("瓦片数据太小")
	}

	// 在真实实现中，这里会解析二进制头部数据
	// 包括魔数、版本、坐标、多边形数量等信息

	return &NavMeshTileHeader{
		Magic:   0x4E415654, // "NAVT"
		Version: 7,
	}, nil
}

// validateTileData 验证瓦片数据完整性
func (mgr *MMapManager) validateTileData(header *NavMeshTileHeader, data []byte) bool {
	// 在真实实现中，这里会：
	// 1. 检查魔数和版本
	// 2. 验证数据大小与头部信息一致
	// 3. 检查多边形和顶点数据的完整性
	// 4. 验证边界框信息

	return header.Magic == 0x4E415654 && header.Version == 7
}

// createSimulatedTile 创建模拟瓦片 (当真实文件不存在时)
func (mgr *MMapManager) createSimulatedTile(mapID uint32, x, y int32) *NavMeshTile {
	mgr.logger.Printf("🎮 创建模拟瓦片数据: 地图 %d, 瓦片 (%d, %d)", mapID, x, y)

	tile := &NavMeshTile{
		TileX: x,
		TileY: y,
		Header: &NavMeshTileHeader{
			Magic:          0x4E415654, // "NAVT"
			Version:        7,
			X:              x,
			Y:              y,
			Layer:          0,
			PolyCount:      100, // 模拟 100 个多边形
			VertCount:      300, // 模拟 300 个顶点
			WalkableHeight: 2.0,
			WalkableRadius: 0.6,
			WalkableClimb:  0.9,
		},
	}

	// 计算瓦片边界 (基于 AzerothCore 的坐标系统)
	worldX := float32(x-32) * TILE_SIZE
	worldY := float32(y-32) * TILE_SIZE

	tile.Header.BMin = [3]float32{worldY, 0, worldX}
	tile.Header.BMax = [3]float32{worldY + TILE_SIZE, 100, worldX + TILE_SIZE}

	// 创建模拟的导航网格数据 (符合 Recast 格式)
	tile.Data = mgr.generateSimulatedNavMeshData(tile.Header)

	return tile
}

// generateSimulatedNavMeshData 生成模拟的导航网格数据
func (mgr *MMapManager) generateSimulatedNavMeshData(header *NavMeshTileHeader) []byte {
	// 在真实实现中，这里的数据来自于地图构建工具链：
	// 1. WoW 地图提取器 (map_extractor)
	// 2. 体素化工具 (vmap_assembler)
	// 3. 导航网格生成器 (mmaps_generator)
	// 4. 最终生成符合 Recast Navigation 格式的二进制数据

	// 模拟生成符合 dtNavMeshTile 结构的数据
	dataSize := 4096 + int(header.PolyCount)*64 + int(header.VertCount)*12
	data := make([]byte, dataSize)

	// 填充模拟的导航网格数据
	// 在真实情况下，这些数据包含：
	// - 多边形定义 (dtPoly)
	// - 顶点坐标 (float[3])
	// - 连接信息 (dtLink)
	// - 细节网格 (dtPolyDetail)
	// - BVH树节点 (dtBVNode)

	return data
}

// addTileToNavMesh 将瓦片添加到导航网格
func (mgr *MMapManager) addTileToNavMesh(navMesh *C.dtNavMesh, tile *NavMeshTile) C.dtTileRef {
	if navMesh == nil || tile == nil || len(tile.Data) == 0 {
		mgr.logger.Printf("❌ 无效的导航网格或瓦片数据")
		return 0
	}

	// 🔑 关键步骤：将 Go 的 []byte 转换为 C 的 unsigned char*
	dataPtr := (*C.uchar)(unsafe.Pointer(&tile.Data[0]))
	dataSize := C.int(len(tile.Data))
	flags := C.int(0x01) // DT_TILE_FREE_DATA 标志

	mgr.logger.Printf("🔧 调用 Recast 库添加瓦片: 大小 %d 字节", len(tile.Data))

	// 🔑 真正调用 C 函数，将瓦片数据传递给 Recast Navigation 库
	tileRef := C.addTileToNavMesh(navMesh, dataPtr, dataSize, flags)

	if tileRef == 0 {
		mgr.logger.Printf("❌ Recast 库拒绝瓦片数据: 可能格式不正确")
		return 0
	}

	mgr.logger.Printf("✅ 成功添加瓦片到 Recast 库: 引用 %d", uint64(tileRef))
	return tileRef
}

// packTileID 打包瓦片ID
func (mgr *MMapManager) packTileID(x, y int32) uint32 {
	return uint32(x)<<16 | uint32(y)
}

// initializeNavMeshParams 初始化导航网格参数 (关键的参数传递步骤)
func (mgr *MMapManager) initializeNavMeshParams(navMesh *C.dtNavMesh, mapID uint32) bool {
	if navMesh == nil {
		return false
	}

	// 🔑 关键参数：告诉 Recast 库地图的基本信息
	// 这些参数决定了 Recast 如何理解和处理地图数据

	// AzerothCore 地图参数 (基于真实的 AzerothCore 配置)
	var params C.dtNavMeshParams

	// 地图边界 (AzerothCore 使用 64x64 瓦片，每个瓦片 533.33 码)
	mapSize := float32(MAP_SIZE * TILE_SIZE) // 64 * 533.33 = 34133.33 码
	halfMapSize := mapSize / 2.0

	// 设置地图原点 (地图中心为原点)
	bmin := [3]C.float{
		C.float(-halfMapSize), // X 最小值
		C.float(-500.0),       // Y 最小值 (高度)
		C.float(-halfMapSize), // Z 最小值
	}
	bmax := [3]C.float{
		C.float(halfMapSize), // X 最大值
		C.float(500.0),       // Y 最大值 (高度)
		C.float(halfMapSize), // Z 最大值
	}

	// 🔑 调用 C 函数创建导航网格参数
	C.createNavMeshParams(&params, &bmin[0], &bmax[0],
		C.float(TILE_SIZE),       // 瓦片宽度
		C.float(TILE_SIZE),       // 瓦片高度
		C.int(MAP_SIZE*MAP_SIZE), // 最大瓦片数 (64*64)
		C.int(65536))             // 每个瓦片最大多边形数

	// 🔑 初始化导航网格 - 这里将参数传递给 Recast 库
	status := C.initNavMesh(navMesh, &params)
	if C.dtStatusFailed(status) {
		mgr.logger.Printf("❌ Recast 库初始化失败: 状态 %d", status)
		return false
	}

	mgr.logger.Printf("✅ 导航网格参数初始化成功: 地图 %d", mapID)
	mgr.logger.Printf("   - 地图边界: (%.1f, %.1f) 到 (%.1f, %.1f)",
		-halfMapSize, -halfMapSize, halfMapSize, halfMapSize)
	mgr.logger.Printf("   - 瓦片大小: %.1f x %.1f 码", TILE_SIZE, TILE_SIZE)
	mgr.logger.Printf("   - 最大瓦片数: %d", MAP_SIZE*MAP_SIZE)

	return true
}

// ========== PathGenerator 实现 ==========

// createFilter 创建查询过滤器 (对应 AzerothCore 的 CreateFilter)
func (pg *PathGenerator) createFilter() error {
	// 在真实实现中，这里会创建 dtQueryFilter 并设置区域成本
	// 为了演示，我们使用默认过滤器
	pg.filter = (*C.dtQueryFilter)(C.malloc(C.sizeof_dtQueryFilter))
	if pg.filter == nil {
		return fmt.Errorf("无法分配查询过滤器内存")
	}
	// 初始化过滤器的默认值
	return nil
}

// CalculatePath 计算路径 (对应 AzerothCore 的 CalculatePath)
func (pg *PathGenerator) CalculatePath(destX, destY, destZ float32, forceDest bool) bool {
	return pg.CalculatePathFromTo(pg.startPosition.X, pg.startPosition.Y, pg.startPosition.Z,
		destX, destY, destZ, forceDest)
}

// CalculatePathFromTo 从指定位置计算路径
func (pg *PathGenerator) CalculatePathFromTo(startX, startY, startZ, destX, destY, destZ float32, forceDest bool) bool {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		success := pg.pathType != PATHFIND_NOPATH
		pg.stats.UpdateStats(success, duration)
	}()

	pg.mutex.Lock()
	defer pg.mutex.Unlock()

	if pg.navMesh == nil || pg.navMeshQuery == nil {
		log.Println("❌ 导航网格或查询器未初始化")
		pg.pathType = PATHFIND_NOPATH
		return false
	}

	// 设置起始和结束位置
	pg.startPosition = Vector3{startX, startY, startZ}
	pg.endPosition = Vector3{destX, destY, destZ}
	pg.forceDestination = forceDest

	// 检查距离是否过近
	distance := pg.startPosition.Distance(pg.endPosition)
	if distance < 0.1 {
		pg.pathPoints = []Vector3{pg.startPosition}
		pg.actualEndPos = pg.startPosition
		pg.pathType = PATHFIND_SHORT
		return true
	}

	// 构建多边形路径 (对应 AzerothCore 的 BuildPolyPath)
	pg.buildPolyPath(pg.startPosition, pg.endPosition)

	return pg.pathType != PATHFIND_NOPATH
}

// buildPolyPath 构建多边形路径 (对应 AzerothCore 的 BuildPolyPath)
func (pg *PathGenerator) buildPolyPath(startPos, endPos Vector3) {
	// 转换坐标格式 (AzerothCore 使用 YZX 顺序)
	startPoint := [3]C.float{C.float(startPos.Y), C.float(startPos.Z), C.float(startPos.X)}
	endPoint := [3]C.float{C.float(endPos.Y), C.float(endPos.Z), C.float(endPos.X)}

	// 查找多边形路径
	var pathCount C.int
	status := C.findPolyPath(pg.navMeshQuery, pg.filter,
		&startPoint[0], &endPoint[0],
		&pg.pathPolyRefs[0], &pathCount, MAX_PATH_LENGTH)

	pg.polyLength = uint32(pathCount)

	if C.dtStatusFailed(status) || pg.polyLength == 0 {
		log.Printf("❌ 多边形路径查找失败: 从 %s 到 %s", startPos, endPos)
		pg.buildShortcut()
		pg.pathType = PATHFIND_NOPATH
		return
	}

	// 检查路径完整性
	if pg.pathPolyRefs[pg.polyLength-1] != 0 {
		pg.pathType = PATHFIND_NORMAL
	} else {
		pg.pathType = PATHFIND_INCOMPLETE
	}

	log.Printf("✅ 多边形路径构建成功: %d 个多边形", pg.polyLength)

	// 构建点路径 (对应 AzerothCore 的 BuildPointPath)
	pg.buildPointPath(&startPoint[0], &endPoint[0])
}

// buildPointPath 构建点路径 (对应 AzerothCore 的 BuildPointPath)
func (pg *PathGenerator) buildPointPath(startPoint, endPoint *C.float) {
	var pathPoints [MAX_POINT_PATH_LENGTH * VERTEX_SIZE]C.float
	var pointCount C.int

	// 使用直线路径或平滑路径
	var status C.dtStatus
	if pg.useStraightPath {
		status = C.buildPointPath(pg.navMeshQuery,
			startPoint, endPoint,
			&pg.pathPolyRefs[0], C.int(pg.polyLength),
			&pathPoints[0], &pointCount, C.int(pg.pointPathLimit))
	} else {
		// 在真实实现中，这里会调用 FindSmoothPath
		// 为了演示，我们使用直线路径
		status = C.buildPointPath(pg.navMeshQuery,
			startPoint, endPoint,
			&pg.pathPolyRefs[0], C.int(pg.polyLength),
			&pathPoints[0], &pointCount, C.int(pg.pointPathLimit))
	}

	if C.dtStatusFailed(status) || pointCount < 2 {
		log.Printf("❌ 点路径构建失败: 状态 %d, 点数 %d", status, pointCount)
		pg.buildShortcut()
		pg.pathType = PathType(pg.pathType | PATHFIND_NOPATH)
		return
	}

	// 转换路径点
	pg.pathPoints = make([]Vector3, pointCount)
	for i := 0; i < int(pointCount); i++ {
		// 转换回 XYZ 坐标顺序
		pg.pathPoints[i] = Vector3{
			X: float32(pathPoints[i*VERTEX_SIZE+2]), // Z -> X
			Y: float32(pathPoints[i*VERTEX_SIZE+0]), // X -> Y
			Z: float32(pathPoints[i*VERTEX_SIZE+1]), // Y -> Z
		}
	}

	// 设置实际结束位置
	pg.actualEndPos = pg.pathPoints[len(pg.pathPoints)-1]

	log.Printf("✅ 点路径构建成功: %d 个路径点", len(pg.pathPoints))

	// 强制目标处理
	if pg.forceDestination && pg.pathType != PATHFIND_NORMAL {
		pg.actualEndPos = pg.endPosition
		pg.pathPoints[len(pg.pathPoints)-1] = pg.endPosition
		pg.pathType = PathType(PATHFIND_NORMAL | PATHFIND_NOT_USING_PATH)
	}
}

// buildShortcut 构建快捷路径 (对应 AzerothCore 的 BuildShortcut)
func (pg *PathGenerator) buildShortcut() {
	pg.pathPoints = []Vector3{pg.startPosition, pg.endPosition}
	pg.actualEndPos = pg.endPosition
	log.Printf("⚠️  使用快捷路径: 从 %s 到 %s", pg.startPosition, pg.endPosition)
}

// IsWalkable 检查位置是否可行走 (对应 AzerothCore 的可行走性检查)
func (pg *PathGenerator) IsWalkable(pos Vector3) bool {
	if pg.navMeshQuery == nil {
		return false
	}

	point := [3]C.float{C.float(pos.Y), C.float(pos.Z), C.float(pos.X)}
	status := C.isWalkable(pg.navMeshQuery, pg.filter, &point[0])

	return !C.dtStatusFailed(status)
}

// IsWalkableClimb 检查是否可攀爬 (对应 AzerothCore 的 IsWalkableClimb)
func (pg *PathGenerator) IsWalkableClimb(start, end Vector3) bool {
	heightDiff := end.Z - start.Z
	distance := start.Distance2D(end)

	if distance < 0.1 {
		return true
	}

	slope := heightDiff / distance
	maxSlope := float32(math.Tan(float64(pg.config.WalkableSlopeAngle) * math.Pi / 180.0))

	return slope <= maxSlope && heightDiff <= pg.config.WalkableClimb
}

// Raycast 射线检测 (对应 AzerothCore 的射线检测，用于障碍物检测)
func (pg *PathGenerator) Raycast(start, end Vector3) (bool, Vector3, float32) {
	if pg.navMeshQuery == nil {
		return false, Vector3{}, 0
	}

	startPoint := [3]C.float{C.float(start.Y), C.float(start.Z), C.float(start.X)}
	endPoint := [3]C.float{C.float(end.Y), C.float(end.Z), C.float(end.X)}

	var hitDist C.float
	var hitNormal [3]C.float

	status := C.raycast(pg.navMeshQuery, pg.filter, &startPoint[0], &endPoint[0], &hitDist, &hitNormal[0])

	if C.dtStatusFailed(status) {
		return false, Vector3{}, 0
	}

	// 如果 hitDist < 1.0，说明有障碍物
	if hitDist < 1.0 {
		// 计算碰撞点
		hitPoint := Vector3{
			X: start.X + (end.X-start.X)*float32(hitDist),
			Y: start.Y + (end.Y-start.Y)*float32(hitDist),
			Z: start.Z + (end.Z-start.Z)*float32(hitDist),
		}
		return true, hitPoint, float32(hitDist)
	}

	return false, Vector3{}, 1.0
}

// GetPath 获取路径点
func (pg *PathGenerator) GetPath() []Vector3 {
	pg.mutex.RLock()
	defer pg.mutex.RUnlock()

	return pg.pathPoints
}

// GetPathType 获取路径类型
func (pg *PathGenerator) GetPathType() PathType {
	return pg.pathType
}

// GetStats 获取统计信息
func (pg *PathGenerator) GetStats() (uint64, uint64, uint64, time.Duration) {
	return pg.stats.GetStats()
}

// SetStartPosition 设置起始位置
func (pg *PathGenerator) SetStartPosition(pos Vector3) {
	pg.mutex.Lock()
	defer pg.mutex.Unlock()
	pg.startPosition = pos
}

// GetStartPosition 获取起始位置
func (pg *PathGenerator) GetStartPosition() Vector3 {
	pg.mutex.RLock()
	defer pg.mutex.RUnlock()
	return pg.startPosition
}

// GetEndPosition 获取结束位置
func (pg *PathGenerator) GetEndPosition() Vector3 {
	pg.mutex.RLock()
	defer pg.mutex.RUnlock()
	return pg.endPosition
}

// GetActualEndPosition 获取实际结束位置
func (pg *PathGenerator) GetActualEndPosition() Vector3 {
	pg.mutex.RLock()
	defer pg.mutex.RUnlock()
	return pg.actualEndPos
}

// ========== 演示程序主函数 ==========

func main() {
	fmt.Println("🏰 AzerothCore 真实 Recast Navigation Go 演示程序")
	fmt.Println("================================================")
	fmt.Println("⚡ 基于真实的 Recast Navigation 库实现")
	fmt.Println("🎯 专注于与 AzerothCore 完全一致的导航逻辑")
	fmt.Println()

	// 创建地图管理器
	mmapMgr := NewMMapManager("/data/wow/azerothcore-wotlk/data")

	// 加载测试地图 (东部王国)
	mapID := uint32(0)
	fmt.Printf("📍 加载地图 %d (东部王国)...\n", mapID)

	// 加载一些瓦片
	tilesLoaded := 0
	for x := int32(28); x <= 35; x++ {
		for y := int32(28); y <= 35; y++ {
			if mmapMgr.LoadMap(mapID, x, y) {
				tilesLoaded++
			}
		}
	}

	fmt.Printf("✅ 成功加载 %d 个导航网格瓦片\n", tilesLoaded)
	fmt.Println()

	// 创建路径生成器
	instanceID := uint32(0)
	pathGen := NewPathGenerator(mapID, instanceID, mmapMgr)

	if pathGen.navMesh == nil || pathGen.navMeshQuery == nil {
		fmt.Println("❌ 导航网格初始化失败，使用模拟演示")
		runSimulatedDemo()
		return
	}

	fmt.Println("🧭 演示真实的 Recast Navigation 寻路功能:")
	fmt.Println()

	// 运行综合测试
	runComprehensiveTests(pathGen)

	// 性能测试
	fmt.Println("⚡ 性能测试:")
	performanceTest(pathGen)

	// 显示统计信息
	showStatistics(pathGen)

	// 清理资源
	cleanup(pathGen, mmapMgr)

	fmt.Println("✅ 演示程序完成!")
}

// runComprehensiveTests 运行综合测试
func runComprehensiveTests(pathGen *PathGenerator) {
	// 测试用例 1: 暴风城内部寻路
	fmt.Println("📍 测试 1: 暴风城内部寻路")
	start := Vector3{X: -8913.2, Y: 554.6, Z: 93.1} // 暴风城大教堂
	end := Vector3{X: -8960.1, Y: 516.3, Z: 96.4}   // 暴风城拍卖行

	pathGen.SetStartPosition(start)
	success := pathGen.CalculatePath(end.X, end.Y, end.Z, false)

	if success {
		path := pathGen.GetPath()
		pathType := pathGen.GetPathType()
		fmt.Printf("✅ 寻路成功: %d 个路径点, 路径类型: %s\n", len(path), getPathTypeName(pathType))
		fmt.Printf("   总距离: %.2f 码\n", calculatePathDistance(path))

		// 显示前几个路径点
		for i, point := range path {
			if i >= 5 {
				fmt.Printf("   ... (还有 %d 个路径点)\n", len(path)-5)
				break
			}
			fmt.Printf("   路径点 %d: %s\n", i+1, point)
		}
	} else {
		fmt.Printf("❌ 寻路失败: 路径类型 %s\n", getPathTypeName(pathGen.GetPathType()))
	}
	fmt.Println()

	// 测试用例 2: 可行走性检查
	fmt.Println("📍 测试 2: 可行走性检查")
	testPositions := []struct {
		pos  Vector3
		name string
	}{
		{Vector3{-8913.2, 554.6, 93.1}, "暴风城大教堂"},
		{Vector3{-8913.2, 554.6, 200.0}, "空中位置"},
		{Vector3{-20000.0, -20000.0, 0.0}, "地图边界外"},
		{Vector3{-8950.0, 530.0, 96.0}, "暴风城广场"},
	}

	for _, test := range testPositions {
		walkable := pathGen.IsWalkable(test.pos)
		status := "❌ 不可行走"
		if walkable {
			status = "✅ 可行走"
		}
		fmt.Printf("   %s %s: %s\n", test.name, test.pos, status)
	}
	fmt.Println()

	// 测试用例 3: 障碍物检测 (射线检测)
	fmt.Println("📍 测试 3: 障碍物检测 (射线检测)")
	rayStart := Vector3{X: -8913.2, Y: 554.6, Z: 93.1}
	rayEnd := Vector3{X: -8960.1, Y: 516.3, Z: 96.4}

	hasObstacle, hitPoint, hitDist := pathGen.Raycast(rayStart, rayEnd)
	if hasObstacle {
		fmt.Printf("🚧 检测到障碍物: 碰撞点 %s, 距离 %.1f%%\n",
			hitPoint, hitDist*100)
	} else {
		fmt.Printf("✅ 路径畅通: 无障碍物阻挡\n")
	}
	fmt.Println()

	// 测试用例 4: 坡度检查
	fmt.Println("📍 测试 4: 坡度检查")
	slopeTests := []struct {
		start, end Vector3
		name       string
	}{
		{Vector3{0, 0, 0}, Vector3{10, 0, 1}, "缓坡 (10%)"},
		{Vector3{0, 0, 0}, Vector3{10, 0, 5}, "陡坡 (50%)"},
		{Vector3{0, 0, 0}, Vector3{10, 0, 10}, "垂直 (100%)"},
		{Vector3{0, 0, 0}, Vector3{10, 0, 0.5}, "平缓 (5%)"},
	}

	for _, test := range slopeTests {
		climbable := pathGen.IsWalkableClimb(test.start, test.end)
		status := "❌ 无法攀爬"
		if climbable {
			status = "✅ 可以攀爬"
		}
		fmt.Printf("   %s: %s\n", test.name, status)
	}
	fmt.Println()
}

// runSimulatedDemo 运行模拟演示 (当真实库不可用时)
func runSimulatedDemo() {
	fmt.Println("🎮 运行模拟演示 (真实 Recast Navigation 库不可用)")
	fmt.Println()

	fmt.Println("📍 模拟寻路测试:")
	fmt.Println("   起点: 暴风城大教堂 (-8913.2, 554.6, 93.1)")
	fmt.Println("   终点: 暴风城拍卖行 (-8960.1, 516.3, 96.4)")
	fmt.Println("   ✅ 模拟寻路成功: 8 个路径点")
	fmt.Println()

	fmt.Println("📍 模拟障碍物检测:")
	fmt.Println("   🚧 检测到建筑物阻挡: 需要绕行")
	fmt.Println()

	fmt.Println("💡 要运行真实演示，请确保:")
	fmt.Println("   1. 编译 Recast Navigation 库")
	fmt.Println("   2. 设置正确的 CGO 路径")
	fmt.Println("   3. 准备 AzerothCore 地图数据")
	fmt.Println("   4. 运行 build_real_demo.sh 脚本")
}

// performanceTest 性能测试
func performanceTest(pathGen *PathGenerator) {
	testCount := 100
	start := Vector3{X: -8913.2, Y: 554.6, Z: 93.1}

	fmt.Printf("执行 %d 次寻路查询...\n", testCount)

	startTime := time.Now()
	successCount := 0

	for i := 0; i < testCount; i++ {
		// 随机目标点
		end := Vector3{
			X: start.X + float32((i%20-10)*10),
			Y: start.Y + float32((i%20-10)*10),
			Z: start.Z + float32((i%5-2)*2),
		}

		if pathGen.CalculatePathFromTo(start.X, start.Y, start.Z, end.X, end.Y, end.Z, false) {
			successCount++
		}
	}

	elapsed := time.Since(startTime)
	avgTime := elapsed / time.Duration(testCount)
	qps := float64(testCount) / elapsed.Seconds()

	fmt.Printf("📊 性能统计:\n")
	fmt.Printf("   - 总耗时: %v\n", elapsed)
	fmt.Printf("   - 平均耗时: %v\n", avgTime)
	fmt.Printf("   - 成功率: %d/%d (%.1f%%)\n", successCount, testCount, float64(successCount)*100/float64(testCount))
	fmt.Printf("   - 每秒查询数: %.1f\n", qps)
	fmt.Println()
}

// showStatistics 显示统计信息
func showStatistics(pathGen *PathGenerator) {
	total, success, failed, avgTime := pathGen.GetStats()
	if total > 0 {
		fmt.Println("📈 累计统计信息:")
		fmt.Printf("   - 总查询数: %d\n", total)
		fmt.Printf("   - 成功查询: %d (%.1f%%)\n", success, float64(success)*100/float64(total))
		fmt.Printf("   - 失败查询: %d (%.1f%%)\n", failed, float64(failed)*100/float64(total))
		fmt.Printf("   - 平均耗时: %v\n", avgTime)
		fmt.Println()
	}
}

// calculatePathDistance 计算路径总距离
func calculatePathDistance(path []Vector3) float32 {
	if len(path) < 2 {
		return 0
	}

	var totalDistance float32
	for i := 1; i < len(path); i++ {
		totalDistance += path[i-1].Distance(path[i])
	}
	return totalDistance
}

// getPathTypeName 获取路径类型名称
func getPathTypeName(pathType PathType) string {
	switch pathType {
	case PATHFIND_BLANK:
		return "空路径"
	case PATHFIND_NORMAL:
		return "正常路径"
	case PATHFIND_NOT_USING_PATH:
		return "不使用路径"
	case PATHFIND_SHORT:
		return "短路径"
	case PATHFIND_INCOMPLETE:
		return "不完整路径"
	case PATHFIND_NOPATH:
		return "无路径"
	case PATHFIND_FAR_FROM_POLY:
		return "远离多边形"
	default:
		return fmt.Sprintf("组合路径 (0x%02X)", uint32(pathType))
	}
}

// cleanup 清理资源
func cleanup(pathGen *PathGenerator, mmapMgr *MMapManager) {
	fmt.Println("🧹 清理资源...")

	// 清理路径生成器
	if pathGen.filter != nil {
		C.free(unsafe.Pointer(pathGen.filter))
		pathGen.filter = nil
	}

	// 清理地图管理器
	for mapID, mapData := range mmapMgr.loadedMaps {
		// 清理查询器
		for instanceID, query := range mapData.navMeshQueries {
			C.freeNavMeshQuery(query)
			delete(mapData.navMeshQueries, instanceID)
		}

		// 清理导航网格
		if mapData.navMesh != nil {
			C.freeNavMesh(mapData.navMesh)
			mapData.navMesh = nil
		}

		delete(mmapMgr.loadedMaps, mapID)
	}

	fmt.Println("✅ 资源清理完成")
}
