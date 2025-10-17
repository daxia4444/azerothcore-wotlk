# AzerothCore 游戏服务器压测方案调研报告

## 📋 目录
1. [项目背景](#项目背景)
2. [压测需求分析](#压测需求分析)
3. [现有方案分析](#现有方案分析)
4. [开源压测框架调研](#开源压测框架调研)
5. [自研压测系统方案](#自研压测系统方案)
6. [方案对比分析](#方案对比分析)
7. [最终选型建议](#最终选型建议)
8. [实施路线图](#实施路线图)

---

## 1. 项目背景

### 1.1 项目概况
- **项目名称**: AzerothCore (魔兽世界3.3.5a私服核心)
- **技术栈**: C++ (服务器端) + MySQL (数据库) + Boost.Asio (网络库)
- **架构**: 
  - AuthServer: 认证服务器 (端口3724)
  - WorldServer: 游戏世界服务器 (端口8085)
  - 数据库: LoginDB, WorldDB, CharacterDB
- **协议**: 自定义二进制协议 (基于Opcode的包结构)

### 1.2 现有压测代码
项目中已有 `go-combat-demo/` 目录，包含：
- 基础的客户端模拟器 (ClientSimulator)
- 网络协议实现 (WorldPacket, WorldSocket)
- 40人团队战斗模拟
- 批量同步机制演示

**现有代码的局限性**:
- ❌ 缺乏系统化的压测指标收集
- ❌ 没有性能瓶颈分析工具
- ❌ 无法模拟大规模并发场景 (1000+ 玩家)
- ❌ 缺少详细的压测报告生成
- ❌ 没有实时监控和可视化

---

## 2. 压测需求分析

### 2.1 核心压测目标
1. **承载量测试**: 确定单服务器最大在线玩家数
2. **性能瓶颈识别**: CPU、内存、网络、数据库瓶颈
3. **稳定性测试**: 长时间运行的稳定性
4. **响应时间**: 各类操作的延迟分析
5. **资源消耗**: 系统资源使用情况

### 2.2 关键性能指标 (KPI)

#### 2.2.1 服务器端指标
| 指标类别 | 具体指标 | 目标值 |
|---------|---------|--------|
| **并发能力** | 最大在线玩家数 | ≥ 1000 |
| **CPU使用率** | 平均/峰值CPU占用 | < 80% |
| **内存使用** | 内存占用/内存泄漏 | < 4GB |
| **网络吞吐** | 包处理速率 (pps) | ≥ 10000 pps |
| **数据库性能** | 查询响应时间 | < 50ms |
| **帧率** | 服务器更新频率 | 50 FPS (20ms/tick) |

#### 2.2.2 客户端体验指标
| 指标 | 描述 | 目标值 |
|-----|------|--------|
| **登录延迟** | 从连接到进入游戏 | < 3s |
| **移动延迟** | 移动指令响应时间 | < 100ms |
| **战斗延迟** | 攻击/技能响应时间 | < 150ms |
| **丢包率** | 网络包丢失率 | < 0.1% |
| **断线率** | 异常断线比例 | < 0.5% |

#### 2.2.3 场景化测试
- **场景1**: 100人同时登录
- **场景2**: 500人在线，50%战斗状态
- **场景3**: 1000人在线，20%战斗，30%移动，50%待机
- **场景4**: 40人团队副本 (高频战斗)
- **场景5**: 长时间稳定性测试 (24小时)

---

## 3. 现有方案分析

### 3.1 当前 go-combat-demo 分析

#### 3.1.1 优势
✅ **协议实现完整**: 已实现核心Opcode处理  
✅ **网络层可用**: WorldSocket、WorldSession 基本可用  
✅ **有基础统计**: ClientStats 收集基本指标  
✅ **真实网络交互**: 通过TCP模拟真实客户端  

#### 3.1.2 不足
❌ **规模受限**: 仅支持40人模拟  
❌ **指标不全**: 缺少服务器端指标采集  
❌ **无可视化**: 没有实时监控界面  
❌ **报告简陋**: 仅打印基础统计信息  
❌ **场景单一**: 只有战斗场景  
❌ **无压力梯度**: 不支持逐步加压测试  

### 3.2 改进建议
如果基于现有代码扩展，需要：
1. 增加 Prometheus + Grafana 监控
2. 实现压测编排器 (支持1000+ 并发)
3. 添加服务器端性能采集 (pprof)
4. 生成详细的HTML/PDF报告
5. 支持分布式压测 (多机器)

---

## 4. 开源压测框架调研

### 4.1 通用压测框架

#### 4.1.1 Locust (Python)
**官网**: https://locust.io/

**特点**:
- ✅ Python编写，易于扩展
- ✅ 支持分布式压测
- ✅ Web UI 实时监控
- ✅ 可编程场景
- ❌ 主要面向HTTP/WebSocket
- ❌ 需要自己实现游戏协议

**适配成本**: ⭐⭐⭐⭐ (高)
```python
# 需要自己实现游戏协议客户端
from locust import User, task, between

class WoWUser(User):
    wait_time = between(1, 3)
    
    def on_start(self):
        # 实现登录逻辑
        self.connect_to_server()
    
    @task
    def cast_spell(self):
        # 实现技能施放
        pass
```

#### 4.1.2 JMeter (Java)
**官网**: https://jmeter.apache.org/

**特点**:
- ✅ 功能强大，插件丰富
- ✅ GUI 配置界面
- ✅ 详细的报告生成
- ❌ Java生态，与C++服务器不匹配
- ❌ 主要面向HTTP协议
- ❌ 性能开销较大

**适配成本**: ⭐⭐⭐⭐⭐ (极高)

#### 4.1.3 Gatling (Scala)
**官网**: https://gatling.io/

**特点**:
- ✅ 高性能，基于Akka
- ✅ DSL 脚本编写
- ✅ 精美的HTML报告
- ❌ Scala学习曲线陡峭
- ❌ 需要实现自定义协议

**适配成本**: ⭐⭐⭐⭐ (高)

### 4.2 游戏专用压测框架

#### 4.2.1 Artillery (Node.js)
**官网**: https://www.artillery.io/

**特点**:
- ✅ 支持WebSocket/Socket.io
- ✅ YAML配置，简单易用
- ✅ 云原生，支持AWS/Azure
- ⚠️ 需要编写自定义引擎
- ❌ 对二进制协议支持有限

**适配成本**: ⭐⭐⭐ (中等)

#### 4.2.2 k6 (Go)
**官网**: https://k6.io/

**特点**:
- ✅ Go语言，性能优秀
- ✅ JavaScript脚本编写
- ✅ 支持自定义协议扩展
- ✅ Grafana官方支持
- ✅ 云原生架构

**适配成本**: ⭐⭐ (较低)

**示例代码**:
```javascript
import { check } from 'k6';
import ws from 'k6/ws';

export default function () {
  const url = 'ws://localhost:8085';
  const params = { tags: { my_tag: 'hello' } };

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', () => {
      // 发送登录包
      socket.sendBinary(loginPacket);
    });

    socket.on('message', (data) => {
      // 处理服务器响应
    });
  });
}
```

#### 4.2.3 Tsung (Erlang)
**官网**: http://tsung.erlang-projects.org/

**特点**:
- ✅ 专为大规模并发设计
- ✅ 支持多种协议
- ✅ 分布式架构
- ❌ Erlang生态小众
- ❌ 配置复杂

**适配成本**: ⭐⭐⭐⭐ (高)

### 4.3 游戏行业实践

#### 4.3.1 Unity Performance Testing
**适用场景**: Unity客户端压测  
**不适用**: 服务器端压测

#### 4.3.2 Unreal Engine Gauntlet
**适用场景**: UE4/UE5 自动化测试  
**不适用**: 非UE服务器

#### 4.3.3 自研方案 (腾讯/网易)
大厂通常采用自研压测平台：
- **腾讯WeTest**: 内部平台，不开源
- **网易Airtest**: 主要用于UI自动化
- **米哈游**: 自研分布式压测系统

---

## 5. 自研压测系统方案

### 5.1 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    压测控制中心 (Master)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  场景编排器   │  │  指标收集器   │  │  报告生成器   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ 压测节点 #1   │    │ 压测节点 #2   │    │ 压测节点 #N   │
│ (500 clients) │    │ (500 clients) │    │ (500 clients) │
└──────────────┘    └──────────────┘    └──────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            ▼
                ┌───────────────────────┐
                │   AzerothCore Server   │
                │  ┌─────────┐           │
                │  │ AuthSvr │           │
                │  └─────────┘           │
                │  ┌─────────┐           │
                │  │WorldSvr │           │
                │  └─────────┘           │
                └───────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │   MySQL Database      │
                └───────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │  Prometheus + Grafana │
                │  (实时监控)            │
                └───────────────────────┘
```

### 5.2 核心组件设计

#### 5.2.1 压测控制中心 (Master)
**技术栈**: Go + gRPC + Web UI

**功能**:
- 场景配置管理 (YAML/JSON)
- 压测任务调度
- 实时指标聚合
- 报告生成 (HTML/PDF/JSON)

**代码结构**:
```go
type StressTestMaster struct {
    config       *TestConfig
    workers      []*WorkerNode
    metrics      *MetricsCollector
    reporter     *ReportGenerator
    orchestrator *ScenarioOrchestrator
}

type TestConfig struct {
    Scenario      string        // 测试场景
    TotalClients  int           // 总客户端数
    RampUpTime    time.Duration // 加压时间
    Duration      time.Duration // 持续时间
    ThinkTime     time.Duration // 思考时间
}
```

#### 5.2.2 压测节点 (Worker)
**技术栈**: Go (复用现有 go-combat-demo)

**功能**:
- 模拟大量客户端 (每节点500-1000)
- 执行测试场景
- 上报性能指标
- 支持热更新配置

**优化点**:
```go
// 使用对象池减少GC压力
var packetPool = sync.Pool{
    New: func() interface{} {
        return &WorldPacket{}
    },
}

// 使用协程池控制并发
type ClientPool struct {
    workers   chan *ClientSimulator
    maxWorkers int
}
```

#### 5.2.3 指标收集器
**技术栈**: Prometheus + Custom Exporter

**服务器端指标** (需要在C++服务器中埋点):
```cpp
// 在 WorldServer 中添加 Prometheus 导出器
class MetricsExporter {
public:
    void RecordPacketReceived(OpcodeClient opcode);
    void RecordPacketSent(OpcodeServer opcode);
    void RecordSessionCount(uint32 count);
    void RecordUpdateTime(uint32 diffMs);
    void RecordDatabaseQuery(const std::string& query, uint32 timeMs);
};
```

**客户端指标** (Go):
```go
var (
    clientConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "wow_client_connections_total",
        Help: "Total number of client connections",
    })
    
    packetsSent = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "wow_packets_sent_total",
        Help: "Total packets sent by clients",
    }, []string{"opcode"})
    
    responseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "wow_response_time_seconds",
        Help: "Response time for operations",
        Buckets: prometheus.DefBuckets,
    }, []string{"operation"})
)
```

#### 5.2.4 报告生成器
**输出格式**:
1. **实时监控**: Grafana Dashboard
2. **HTML报告**: 详细的图表和分析
3. **JSON数据**: 供CI/CD集成
4. **PDF报告**: 管理层汇报

**报告内容**:
```markdown
# 压测报告

## 测试概览
- 测试时间: 2025-10-17 16:00:00
- 测试场景: 1000人在线混合场景
- 测试时长: 30分钟

## 性能指标
### 服务器性能
- 最大在线: 1000人
- 平均CPU: 65%
- 峰值CPU: 82%
- 内存占用: 3.2GB
- 网络吞吐: 8500 pps

### 客户端体验
- 平均延迟: 85ms
- P95延迟: 150ms
- P99延迟: 280ms
- 丢包率: 0.05%
- 断线率: 0.2%

## 瓶颈分析
1. **数据库查询**: 角色登录时查询耗时较长 (120ms)
2. **网络带宽**: 40人团战时带宽达到瓶颈
3. **CPU热点**: Spell::Update 占用15% CPU

## 优化建议
1. 增加数据库连接池
2. 优化批量同步算法
3. 使用缓存减少数据库查询
```

### 5.3 场景编排

#### 5.3.1 场景配置 (YAML)
```yaml
scenarios:
  - name: "登录压测"
    duration: 5m
    clients:
      total: 1000
      ramp_up: 2m
    actions:
      - type: login
        weight: 100
        
  - name: "混合场景"
    duration: 30m
    clients:
      total: 1000
      ramp_up: 5m
    actions:
      - type: idle
        weight: 50
      - type: move
        weight: 30
      - type: combat
        weight: 15
      - type: cast_spell
        weight: 5
    think_time: 1s
    
  - name: "团队副本"
    duration: 15m
    clients:
      total: 40
      ramp_up: 30s
    actions:
      - type: combat
        weight: 70
      - type: cast_spell
        weight: 20
      - type: move
        weight: 10
    think_time: 100ms
```

#### 5.3.2 压力梯度测试
```go
type LoadProfile struct {
    Stages []Stage
}

type Stage struct {
    Duration time.Duration
    Target   int  // 目标并发数
}

// 示例: 阶梯式加压
profile := LoadProfile{
    Stages: []Stage{
        {Duration: 2*time.Minute, Target: 100},   // 0-2分钟: 100人
        {Duration: 2*time.Minute, Target: 300},   // 2-4分钟: 300人
        {Duration: 2*time.Minute, Target: 500},   // 4-6分钟: 500人
        {Duration: 2*time.Minute, Target: 1000},  // 6-8分钟: 1000人
        {Duration: 10*time.Minute, Target: 1000}, // 8-18分钟: 保持1000人
        {Duration: 2*time.Minute, Target: 0},     // 18-20分钟: 逐步下降
    },
}
```

### 5.4 分布式架构

#### 5.4.1 多节点协调
```go
// Master 节点
type Master struct {
    workers map[string]*WorkerClient
}

func (m *Master) DistributeLoad(totalClients int) {
    clientsPerWorker := totalClients / len(m.workers)
    
    for _, worker := range m.workers {
        worker.StartClients(clientsPerWorker)
    }
}

// Worker 节点
type Worker struct {
    masterAddr string
    clients    []*ClientSimulator
}

func (w *Worker) StartClients(count int) error {
    for i := 0; i < count; i++ {
        client := NewClientSimulator(...)
        go client.Run()
        w.clients = append(w.clients, client)
    }
    return nil
}
```

#### 5.4.2 通信协议 (gRPC)
```protobuf
syntax = "proto3";

service StressTestService {
  rpc StartTest(TestRequest) returns (TestResponse);
  rpc StopTest(StopRequest) returns (StopResponse);
  rpc GetMetrics(MetricsRequest) returns (stream Metrics);
}

message TestRequest {
  string scenario = 1;
  int32 client_count = 2;
  int32 duration_seconds = 3;
}

message Metrics {
  int64 timestamp = 1;
  int32 active_clients = 2;
  double cpu_usage = 3;
  int64 memory_bytes = 4;
  int64 packets_sent = 5;
  int64 packets_received = 6;
}
```

### 5.5 服务器端埋点

#### 5.5.1 C++ 性能采集
```cpp
// src/server/game/World/World.cpp
void World::Update(uint32 diff)
{
    auto startTime = std::chrono::high_resolution_clock::now();
    
    // 原有更新逻辑
    UpdateSessions(diff);
    
    auto endTime = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::microseconds>(
        endTime - startTime).count();
    
    // 记录性能指标
    sMetrics->RecordUpdateTime("sessions", duration);
}

// 新增 MetricsCollector 类
class MetricsCollector
{
public:
    void RecordUpdateTime(const std::string& component, uint64 microseconds);
    void RecordPacketCount(OpcodeClient opcode);
    void RecordDatabaseQuery(const std::string& table, uint64 microseconds);
    void ExportToPrometheus(uint16 port);
    
private:
    std::unordered_map<std::string, std::vector<uint64>> m_timings;
    std::unordered_map<uint16, uint64> m_packetCounts;
};
```

#### 5.5.2 数据库性能监控
```cpp
// src/server/database/Database/DatabaseWorkerPool.h
template<class T>
class DatabaseWorkerPool
{
public:
    QueryResult Query(const char* sql)
    {
        auto start = std::chrono::high_resolution_clock::now();
        
        QueryResult result = _queue.Query(sql);
        
        auto end = std::chrono::high_resolution_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(
            end - start).count();
        
        // 记录慢查询
        if (duration > 100) {
            LOG_WARN("sql.performance", "Slow query ({}ms): {}", duration, sql);
        }
        
        sMetrics->RecordDatabaseQuery(sql, duration);
        return result;
    }
};
```

---

## 6. 方案对比分析

### 6.1 综合对比表

| 维度 | 开源框架 (k6) | 自研系统 | 扩展现有代码 |
|-----|--------------|---------|-------------|
| **开发成本** | ⭐⭐ (2周) | ⭐⭐⭐⭐ (2个月) | ⭐⭐⭐ (1个月) |
| **协议适配** | ⭐⭐⭐ (需要扩展) | ⭐⭐⭐⭐⭐ (完全定制) | ⭐⭐⭐⭐⭐ (已实现) |
| **性能** | ⭐⭐⭐⭐ (优秀) | ⭐⭐⭐⭐⭐ (极致优化) | ⭐⭐⭐ (需优化) |
| **可扩展性** | ⭐⭐⭐ (有限) | ⭐⭐⭐⭐⭐ (完全控制) | ⭐⭐⭐⭐ (较好) |
| **监控能力** | ⭐⭐⭐⭐ (Grafana) | ⭐⭐⭐⭐⭐ (定制化) | ⭐⭐ (需添加) |
| **报告质量** | ⭐⭐⭐ (标准报告) | ⭐⭐⭐⭐⭐ (深度分析) | ⭐⭐ (基础统计) |
| **学习曲线** | ⭐⭐ (简单) | ⭐⭐⭐⭐ (复杂) | ⭐⭐⭐ (中等) |
| **维护成本** | ⭐⭐ (社区支持) | ⭐⭐⭐⭐⭐ (自己维护) | ⭐⭐⭐ (中等) |
| **分布式支持** | ⭐⭐⭐⭐ (原生支持) | ⭐⭐⭐⭐⭐ (定制) | ⭐ (需重构) |
| **游戏特性** | ⭐⭐ (通用) | ⭐⭐⭐⭐⭐ (专用) | ⭐⭐⭐⭐ (已有基础) |

### 6.2 成本分析

#### 6.2.1 开发成本
| 方案 | 人力 | 时间 | 总成本 (人月) |
|-----|------|------|--------------|
| k6 扩展 | 1人 | 2周 | 0.5 |
| 扩展现有代码 | 1-2人 | 1个月 | 1.5 |
| 完全自研 | 2-3人 | 2个月 | 5 |

#### 6.2.2 运维成本
| 方案 | 服务器 | 维护 | 年成本 |
|-----|--------|------|--------|
| k6 | 2台 (Master+Worker) | 低 | ¥5000 |
| 扩展现有 | 3台 | 中 | ¥8000 |
| 完全自研 | 5台 (分布式) | 高 | ¥15000 |

### 6.3 风险评估

#### 6.3.1 技术风险
| 方案 | 风险点 | 风险等级 | 缓解措施 |
|-----|--------|---------|---------|
| k6 | 协议适配复杂 | 🟡 中 | 先做POC验证 |
| 扩展现有 | 性能瓶颈 | 🟡 中 | 逐步优化 |
| 完全自研 | 开发周期长 | 🔴 高 | 分阶段交付 |

#### 6.3.2 业务风险
| 风险 | 影响 | 概率 | 应对 |
|-----|------|------|------|
| 压测不准确 | 高 | 低 | 与真实环境对比验证 |
| 服务器崩溃 | 高 | 中 | 在测试环境进行 |
| 数据污染 | 中 | 低 | 使用独立测试数据库 |

---

## 7. 最终选型建议

### 7.1 推荐方案: **混合方案 (扩展现有代码 + Prometheus/Grafana)**

#### 7.1.1 选型理由

**✅ 优势**:
1. **快速落地**: 基于现有 `go-combat-demo`，1个月可交付
2. **协议完备**: 已实现核心Opcode，无需重复开发
3. **成本可控**: 开发成本1.5人月，运维成本适中
4. **可扩展**: 后期可逐步演进为完全自研系统
5. **技术栈统一**: Go语言，团队熟悉
6. **监控成熟**: Prometheus + Grafana 业界标准

**⚠️ 劣势**:
1. 需要重构现有代码以支持大规模并发
2. 分布式能力需要额外开发
3. 报告生成需要自己实现

**🎯 适用场景**:
- 中小型团队 (1-3人)
- 需要快速验证服务器性能
- 预算有限 (< 2人月)
- 后续有持续优化计划

#### 7.1.2 实施方案

**阶段一: 基础增强 (2周)**
1. 重构 `ClientSimulator` 支持1000+并发
2. 集成 Prometheus 指标导出
3. 搭建 Grafana 监控面板
4. 实现基础场景编排

**阶段二: 服务器埋点 (1周)**
1. 在 WorldServer 中添加性能采集
2. 导出 Prometheus 指标
3. 数据库慢查询监控

**阶段三: 报告生成 (1周)**
1. 实现 HTML 报告生成
2. 添加性能分析图表
3. 瓶颈识别算法

**交付物**:
- ✅ 支持1000人并发的压测工具
- ✅ 实时监控Dashboard
- ✅ 详细的HTML压测报告
- ✅ 使用文档和最佳实践

### 7.2 备选方案

#### 7.2.1 方案B: k6 + 自定义扩展 (适合快速验证)
**适用场景**: 
- 需要在1周内快速验证
- 对深度定制要求不高
- 团队熟悉JavaScript

**实施步骤**:
```javascript
// 1. 编写 k6 扩展 (Go)
package wow

import (
    "go.k6.io/k6/js/modules"
)

func init() {
    modules.Register("k6/x/wow", new(WoW))
}

type WoW struct{}

func (*WoW) Connect(addr string) (*Client, error) {
    // 复用 go-combat-demo 的客户端代码
    return NewClient(addr)
}

// 2. 编写测试脚本 (JavaScript)
import wow from 'k6/x/wow';

export default function() {
    const client = wow.connect('localhost:8085');
    client.login('player1', 'password');
    client.castSpell(1234);
}
```

#### 7.2.2 方案C: 完全自研 (适合长期投入)
**适用场景**:
- 大型团队 (5+人)
- 有充足预算 (3+人月)
- 需要极致性能和定制化
- 计划支持多个游戏项目

**核心特性**:
- 分布式架构 (支持10000+并发)
- AI驱动的场景生成
- 自动化瓶颈分析
- 与CI/CD深度集成
- 支持多游戏协议

---

## 8. 实施路线图

### 8.1 第一阶段: MVP (2周)

#### Week 1: 核心功能开发
**目标**: 支持500人并发压测

**任务清单**:
- [ ] 重构 `ClientSimulator` 使用协程池
- [ ] 实现压测控制器 (Master)
- [ ] 添加 Prometheus 指标导出
- [ ] 实现3个基础场景 (登录/移动/战斗)

**代码示例**:
```go
// stress_test_master.go
type StressTestMaster struct {
    config      *TestConfig
    clientPool  *ClientPool
    metrics     *prometheus.Registry
}

func (m *StressTestMaster) Run() error {
    // 1. 启动 Prometheus HTTP 服务器
    go m.startMetricsServer()
    
    // 2. 按照配置创建客户端
    for i := 0; i < m.config.TotalClients; i++ {
        client := m.clientPool.Get()
        go client.Run(m.config.Scenario)
        
        // 控制加压速度
        time.Sleep(m.config.RampUpTime / time.Duration(m.config.TotalClients))
    }
    
    // 3. 等待测试完成
    time.Sleep(m.config.Duration)
    
    // 4. 生成报告
    return m.generateReport()
}
```

#### Week 2: 监控和报告
**目标**: 可视化监控 + 基础报告

**任务清单**:
- [ ] 配置 Grafana Dashboard
- [ ] 实现 HTML 报告生成
- [ ] 添加性能指标图表
- [ ] 编写使用文档

**Grafana Dashboard 配置**:
```json
{
  "dashboard": {
    "title": "AzerothCore 压测监控",
    "panels": [
      {
        "title": "在线玩家数",
        "targets": [{
          "expr": "wow_client_connections_total"
        }]
      },
      {
        "title": "服务器CPU使用率",
        "targets": [{
          "expr": "rate(process_cpu_seconds_total[1m]) * 100"
        }]
      },
      {
        "title": "响应时间分布",
        "targets": [{
          "expr": "histogram_quantile(0.95, wow_response_time_seconds_bucket)"
        }]
      }
    ]
  }
}
```

### 8.2 第二阶段: 增强 (2周)

#### Week 3: 服务器端埋点
**目标**: 深度性能分析

**任务清单**:
- [ ] 在 WorldServer 中添加 MetricsCollector
- [ ] 实现数据库查询监控
- [ ] 添加 CPU Profiling
- [ ] 实现内存泄漏检测

**C++ 埋点代码**:
```cpp
// src/server/game/Metrics/MetricsCollector.h
class AC_GAME_API MetricsCollector
{
public:
    static MetricsCollector* instance();
    
    void RecordPacketReceived(OpcodeClient opcode);
    void RecordUpdateTime(const std::string& component, uint64 microseconds);
    void RecordDatabaseQuery(const std::string& query, uint64 milliseconds);
    
    void StartPrometheusExporter(uint16 port);
    
private:
    std::atomic<uint64> m_totalPacketsReceived{0};
    std::unordered_map<std::string, std::vector<uint64>> m_updateTimes;
    std::mutex m_mutex;
};

// 使用示例
void WorldSession::HandleAttackSwingOpcode(WorldPacket& recvPacket)
{
    sMetrics->RecordPacketReceived(CMSG_ATTACKSWING);
    
    auto start = std::chrono::high_resolution_clock::now();
    
    // 原有逻辑
    // ...
    
    auto end = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::microseconds>(
        end - start).count();
    sMetrics->RecordUpdateTime("AttackSwing", duration);
}
```

#### Week 4: 高级特性
**目标**: 分布式 + 自动化

**任务清单**:
- [ ] 实现分布式压测 (gRPC)
- [ ] 添加自动化瓶颈识别
- [ ] 实现压力梯度测试
- [ ] 集成 CI/CD

**分布式架构**:
```go
// distributed_master.go
type DistributedMaster struct {
    workers []*WorkerClient
}

func (m *DistributedMaster) DistributeLoad(totalClients int) {
    clientsPerWorker := totalClients / len(m.workers)
    
    var wg sync.WaitGroup
    for _, worker := range m.workers {
        wg.Add(1)
        go func(w *WorkerClient) {
            defer wg.Done()
            w.StartClients(clientsPerWorker)
        }(worker)
    }
    wg.Wait()
}

// worker_node.go
type WorkerNode struct {
    masterAddr string
    clients    []*ClientSimulator
}

func (w *WorkerNode) StartClients(count int) error {
    for i := 0; i < count; i++ {
        client := NewClientSimulator(...)
        go client.Run()
        w.clients = append(w.clients, client)
    }
    return nil
}
```

### 8.3 第三阶段: 优化 (持续)

**长期优化方向**:
1. **性能优化**: 支持10000+并发
2. **AI场景生成**: 基于真实玩家行为
3. **自动化回归**: 每日自动压测
4. **多游戏支持**: 抽象协议层
5. **云原生**: 支持K8s部署

---

## 9. 关键技术点

### 9.1 高并发优化

#### 9.1.1 协程池
```go
type ClientPool struct {
    clients chan *ClientSimulator
    factory func() *ClientSimulator
}

func NewClientPool(size int, factory func() *ClientSimulator) *ClientPool {
    pool := &ClientPool{
        clients: make(chan *ClientSimulator, size),
        factory: factory,
    }
    
    for i := 0; i < size; i++ {
        pool.clients <- factory()
    }
    
    return pool
}

func (p *ClientPool) Get() *ClientSimulator {
    return <-p.clients
}

func (p *ClientPool) Put(client *ClientSimulator) {
    select {
    case p.clients <- client:
    default:
        // 池已满，丢弃
    }
}
```

#### 9.1.2 对象池
```go
var packetPool = sync.Pool{
    New: func() interface{} {
        return &WorldPacket{
            data: make([]byte, 0, 1024),
        }
    },
}

func GetPacket() *WorldPacket {
    return packetPool.Get().(*WorldPacket)
}

func PutPacket(packet *WorldPacket) {
    packet.Reset()
    packetPool.Put(packet)
}
```

#### 9.1.3 批量处理
```go
type BatchProcessor struct {
    queue   chan *WorldPacket
    batchSize int
    flushInterval time.Duration
}

func (bp *BatchProcessor) Start() {
    ticker := time.NewTicker(bp.flushInterval)
    defer ticker.Stop()
    
    batch := make([]*WorldPacket, 0, bp.batchSize)
    
    for {
        select {
        case packet := <-bp.queue:
            batch = append(batch, packet)
            if len(batch) >= bp.batchSize {
                bp.processBatch(batch)
                batch = batch[:0]
            }
            
        case <-ticker.C:
            if len(batch) > 0 {
                bp.processBatch(batch)
                batch = batch[:0]
            }
        }
    }
}
```

### 9.2 指标采集最佳实践

#### 9.2.1 低开销采集
```go
// 使用原子操作避免锁竞争
type LowOverheadMetrics struct {
    packetsSent     atomic.Uint64
    packetsReceived atomic.Uint64
    totalLatency    atomic.Uint64
    sampleCount     atomic.Uint64
}

func (m *LowOverheadMetrics) RecordLatency(latency time.Duration) {
    m.totalLatency.Add(uint64(latency.Microseconds()))
    m.sampleCount.Add(1)
}

func (m *LowOverheadMetrics) GetAverageLatency() time.Duration {
    total := m.totalLatency.Load()
    count := m.sampleCount.Load()
    if count == 0 {
        return 0
    }
    return time.Duration(total/count) * time.Microsecond
}
```

#### 9.2.2 采样策略
```go
// 高频操作使用采样，避免性能影响
type SamplingMetrics struct {
    sampleRate float64 // 0.01 = 1%采样率
}

func (m *SamplingMetrics) RecordIfSampled(fn func()) {
    if rand.Float64() < m.sampleRate {
        fn()
    }
}

// 使用示例
metrics.RecordIfSampled(func() {
    responseTime.WithLabelValues("attack").Observe(latency.Seconds())
})
```

### 9.3 报告生成

#### 9.3.1 HTML模板
```go
const reportTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>压测报告 - {{.TestName}}</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <h1>{{.TestName}}</h1>
    <h2>测试概览</h2>
    <table>
        <tr><td>测试时间</td><td>{{.StartTime}}</td></tr>
        <tr><td>测试时长</td><td>{{.Duration}}</td></tr>
        <tr><td>总客户端</td><td>{{.TotalClients}}</td></tr>
    </table>
    
    <h2>性能指标</h2>
    <canvas id="latencyChart"></canvas>
    
    <script>
        const ctx = document.getElementById('latencyChart');
        new Chart(ctx, {
            type: 'line',
            data: {
                labels: {{.TimeLabels}},
                datasets: [{
                    label: '响应时间 (ms)',
                    data: {{.LatencyData}},
                }]
            }
        });
    </script>
</body>
</html>
`

type ReportData struct {
    TestName     string
    StartTime    time.Time
    Duration     time.Duration
    TotalClients int
    TimeLabels   []string
    LatencyData  []float64
}

func GenerateReport(data *ReportData) error {
    tmpl, err := template.New("report").Parse(reportTemplate)
    if err != nil {
        return err
    }
    
    f, err := os.Create("report.html")
    if err != nil {
        return err
    }
    defer f.Close()
    
    return tmpl.Execute(f, data)
}
```

---

## 10. 总结

### 10.1 核心结论

**推荐方案**: 扩展现有 `go-combat-demo` + Prometheus/Grafana

**理由**:
1. ✅ **性价比最高**: 1.5人月成本，4周交付
2. ✅ **风险可控**: 基于已有代码，技术栈熟悉
3. ✅ **功能完备**: 满足当前所有压测需求
4. ✅ **可持续演进**: 后期可逐步升级为完全自研

### 10.2 预期收益

**短期收益** (1个月内):
- 🎯 确定服务器承载量 (目标: 1000人)
- 🎯 识别性能瓶颈 (CPU/内存/网络/数据库)
- 🎯 优化服务器配置
- 🎯 建立性能基线

**长期收益** (3-6个月):
- 🎯 持续性能监控
- 🎯 自动化回归测试
- 🎯 容量规划依据
- 🎯 优化迭代闭环

### 10.3 下一步行动

**立即行动** (本周):
1. [ ] 评审本报告，确认技术方案
2. [ ] 分配开发资源 (1-2人)
3. [ ] 搭建测试环境 (独立服务器)
4. [ ] 创建项目仓库和任务看板

**第一周**:
1. [ ] 重构 `ClientSimulator` 支持1000并发
2. [ ] 集成 Prometheus 指标导出
3. [ ] 实现基础场景编排器
4. [ ] 编写单元测试

**第二周**:
1. [ ] 搭建 Grafana 监控面板
2. [ ] 实现 HTML 报告生成
3. [ ] 执行首次压测 (100人)
4. [ ] 编写使用文档

**第三周**:
1. [ ] 在 WorldServer 中添加性能埋点
2. [ ] 实现数据库监控
3. [ ] 执行中等规模压测 (500人)
4. [ ] 分析瓶颈并优化

**第四周**:
1. [ ] 实现分布式压测支持
2. [ ] 执行大规模压测 (1000人)
3. [ ] 生成完整压测报告
4. [ ] 项目总结和知识沉淀

---

## 附录

### A. 参考资料

**开源项目**:
- [k6](https://github.com/grafana/k6) - 现代化负载测试工具
- [Locust](https://github.com/locustio/locust) - Python压测框架
- [Gatling](https://github.com/gatling/gatling) - Scala压测框架

**游戏服务器压测**:
- [MMO Server Architecture](https://www.gabrielgambetta.com/client-server-game-architecture.html)
- [Game Server Performance Testing](https://aws.amazon.com/blogs/gametech/game-server-performance-testing/)

**监控和可观测性**:
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)

### B. 工具清单

**必需工具**:
- Go 1.21+ (压测客户端)
- Prometheus (指标收集)
- Grafana (可视化)
- Docker (容器化部署)

**可选工具**:
- pprof (Go性能分析)
- Valgrind (C++内存分析)
- Wireshark (网络抓包)
- MySQL Workbench (数据库分析)

### C. 团队技能要求

**必需技能**:
- Go语言开发 (中级)
- 网络编程 (TCP/Socket)
- 性能分析基础
- Linux运维基础

**加分技能**:
- C++开发 (用于服务器埋点)
- Prometheus/Grafana使用
- 分布式系统经验
- 游戏服务器经验

---

**报告编写**: AI Assistant  
**报告日期**: 2025-10-17  
**版本**: v1.0  
**状态**: 待评审
