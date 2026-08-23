# 路线图 (Roadmap) — GoAPIScanner

> 版本：v1.0 ｜ 冻结需求见 `docs/Requirements.md`
> **构建顺序：Logic-First（Phase 3 先于 Phase 2）**
> 理由：缺陷树、实时流、报告预览的组件结构由扫描任务 / 发现 / 事件数据模型派生，必须先冻结引擎契约再铺 UI。

---

## 阶段边界

| 阶段 | 范围 | 状态 |
|---|---|---|
| MVP | Swagger 解析 + 载荷变异 + 并发引擎 + 指纹匹配 + 内置靶场 + 任务/WS + 三页 Vue + PDF | 完成 |
| V1 | 与 MVP 同界（<10k LoC，一次交付） | 完成 |
| V2 | 超出范围：集群调度、真实 CVE 情报、账号体系 | 不做 |

---

## Phase 1 — 架构骨架

- [x] Git 初始化与 `.gitignore`
- [x] 目录：`backend` / `frontend-user` / `target-lab`
- [x] `docker-compose.yml` Dev 随机端口：28481 / 28482 / 28483
- [x] 阶段顺序决策：Logic-First

## Phase 3 — 逻辑引擎（先于 UI）

- [x] B1 OpenAPI 2/3 解析器
- [x] B2 载荷库 + 参数点变异
- [x] B3 worker pool 发包引擎（限流 / context 取消 / 防重入）
- [x] B4 手写指纹匹配器（状态码 / 头 / Body / 时序）
- [x] B5 SQLite 任务与发现落库
- [x] B6 WebSocket 事件推送
- [x] B7 PDF 合规报告
- [x] 内置靶场 `target-lab` + Swagger
- [x] 单元测试：解析 / 匹配 / 引擎

## Phase 2 — UI

- [x] `docs/DesignSpec.md`（Midnight Radar 视觉）
- [x] F1 扫描控制台
- [x] F2 缺陷树看板 + 实时打印流
- [x] F3 报告预览 + PDF 下载

## Phase 4 — QA

- [x] `tests/api_smoke.py`（Mock 靶场，¥0）
- [x] Playwright 关键路径脚本（`tests/e2e_flow.spec.ts`）
- [x] `docs/QA_Record.md`

## Phase 5 — 审核

- [x] `docs/AuditReport.md` PASS
- [x] Knowledge Harvest
