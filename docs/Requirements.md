# 需求文档 (Requirements) — GoAPIScanner

> 云原生自动化 API 缺陷扫描 · 渗透漏洞利用 · 合规报告平台
> 状态：**冻结 (Frozen)** ｜ 版本：v1.0 ｜ 冻结时间：2026-08-23 22:50 (GMT+8)

---

## 1. 项目定位 (WHAT)

一个面向企业上线前的**防御性安全测试平台**：解析目标接口(基础 URL 或 Swagger JSON)，
自动生成携带恶意载荷的变异请求进行扫描，实时推送发现的漏洞，并生成 PDF 级合规报告。

**合规红线**：本工具仅用于扫描**用户拥有或已获授权**的目标。系统默认自带**内置漏洞靶场**
(`target-lab`)，所有演示/测试均对内网靶场进行，不对公网发包。前端提交扫描前须勾选授权声明。

---

## 2. 角色与用户旅程

| 角色 | 核心旅程 |
|---|---|
| 安全工程师 | 输入 URL / 上传 Swagger → 勾选授权 → 启动扫描 → 实时看漏洞流 → 查看缺陷树 → 导出合规报告 |
| 开发人员 | 打开报告 → 按漏洞定位接口 → 阅读代码修复建议 |

---

## 3. 功能需求 (Functional)

### 3.1 前端 (Vue 3 + Vite + TS)
- **F1 扫描控制台**：输入基础 URL；或上传 Swagger/OpenAPI JSON 文件(拖拽 + 校验)；配置并发数、超时；授权声明勾选框。
- **F2 漏洞监控看板**：
  - 全生命周期**缺陷树**(接口 → 端点 → 漏洞项)，树级折叠。
  - 严重度**高亮标签**：`严重(Critical)` / `高危(High)` / `中危(Medium)` / `低危(Low)` / `信息(Info)`。
  - **漏洞实时打印流**：WebSocket/SSE 逐条追加扫描日志与命中漏洞。
  - 扫描进度条 + 统计(已发请求数 / 命中数 / 各严重度计数)。
- **F3 合规报告预览页**：渲染精美报告(封面/摘要/漏洞明细/修复建议)，一键下载 **PDF**。

### 3.2 后端 (Go)
- **B1 Swagger 解析器**：解析 OpenAPI 2.0 / 3.0 JSON，抽取 path、method、参数(query/path/body/header)。
- **B2 载荷变异引擎**：内置 Payload 库(SQLi / XSS / 未授权 / 路径遍历 / 命令注入等)，对每个参数点生成变异请求集。
- **B3 高性能并发发包引擎**：基于 `net/http` + goroutine + worker pool + 信号量限流；
  **防重入与防死锁**(context 取消、超时、优雅关闭、无共享可变状态竞争)。
- **B4 漏洞指纹匹配器**：手写规则引擎，依据 **状态码 + 响应头 + Body 特征串 + 时序特征(盲注延迟)** 判定漏洞并定级。
- **B5 任务管理**：扫描任务创建/查询/进度；结果落库(SQLite，零外部依赖)。
- **B6 实时推送**：WebSocket/SSE 将扫描事件推给前端。
- **B7 报告生成**：服务端生成 PDF(gofpdf 或 maroto)，含漏洞清单 + 修复建议映射。

### 3.3 内置靶场 (target-lab)
- 独立 Go 服务，故意暴露 SQLi / XSS / 未授权 / 时序盲注等可被稳定命中的端点，并附带 Swagger JSON，供全流程闭环演示。

---

## 4. 非功能需求 & 可度量验收基线 (Acceptance Baselines)

| 维度 | 基线 |
|---|---|
| 并发性能 | worker pool 可配置并发；对靶场 ≥ 1000 变异请求场景稳定完成，无 goroutine 泄漏 / 死锁 |
| 扫描准确性 | 对内置靶场已知漏洞**召回率 = 100%**(靶场漏洞全部命中)；无 panic |
| 实时性 | 漏洞事件从命中到前端展示延迟 < 1s |
| 稳定性 | context 取消可中断扫描；重复启动/停止不产生残留 goroutine |
| 报告 | PDF 成功生成且可下载，含摘要 + 明细 + 修复建议 |
| 交付 | `docker compose up --build` 一键起 3 服务，localhost 可访问 |
| 测试 | 后端覆盖 Swagger 解析 / 指纹匹配 / 并发引擎核心单元测试；提供 API Smoke 测试(Mock 靶场，成本 ¥0) |
| 时区 | 全部时间使用 GMT+8 |

---

## 5. 技术栈

- **后端**：Go 1.22+，标准库 `net/http`，`gorilla/websocket`(或标准 SSE)，`modernc.org/sqlite`(纯 Go 免 CGO)，PDF: `go-pdf/fpdf`。
- **前端**：Vue 3 + Vite + TypeScript + Pinia + Element Plus / 自研树组件 + WebSocket。
- **靶场**：Go(独立最小服务)。
- **部署**：Docker 多阶段构建 + docker-compose，`TZ=Asia/Shanghai`。

---

## 6. 交付形态 (Docker Delivery Standard)

```
docker compose up --build -d
```

| 服务 | 说明 | 端口(Dev 随机 / 交付 8081+) |
|---|---|---|
| frontend-user | Vue 扫描控制台 + 看板 + 报告 | Web |
| backend | Go 扫描引擎 + API + WS + 报告 | API |
| target-lab | 内置漏洞靶场(扫描目标) | API |

---

## 7. 范围边界 (No Scope Drift)

- ✅ 做：Swagger 解析、载荷变异、并发扫描、指纹匹配、实时推送、缺陷树、PDF 报告、内置靶场。
- ❌ 不做：分布式集群调度、真实公网目标扫描、账号权限体系(单机工具，可留最简登录占位)、真实 CVE 情报库对接(用内置规则库)。

---

## 8. 待架构阶段决策 (交由 Phase 1)

- **构建顺序**：看板/缺陷树是**由扫描结果数据模型派生**的可视化 → 倾向 **Logic-First**(先定数据模型与引擎，再建 UI)。最终由 Chief Architect 在 `docs/Roadmap.md` 记录并给出一句话理由。
