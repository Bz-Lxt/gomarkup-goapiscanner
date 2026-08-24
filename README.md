## 1. 如何启动

在项目根目录执行 `docker compose up --build -d`。首次构建会拉取 Go / Node 基础镜像。就绪后浏览器打开 `http://localhost:28481`。

## 2. 使用说明

打开扫描控制台，默认已填入内置靶场地址。勾选授权声明后启动扫描，随后进入缺陷树看板与实时打印流。扫描结束后可打开合规报告预览并下载 PDF。

## 3. 服务列表及API说明

- 前端控制台：http://localhost:28481
- 扫描引擎 API：http://localhost:28482/api/v1/health
- 内置靶场：http://localhost:28483/swagger.json
- 接口契约见 `docs/API.md`

## 4. 测试账号

本系统为单机安全工具，无需登录。演示扫描目标为内置靶场，无需账号。

## 5. 题目内容

用 Go 实现云原生 API 缺陷扫描、载荷变异发包、指纹匹配与合规报告平台；前端 Vue 3 提供控制台、缺陷树与 PDF 级报告预览。

## 6. 项目结构

- `backend`：扫描引擎、指纹匹配、SQLite、WebSocket、PDF
- `frontend-user`：Vue 3 控制台 / 看板 / 报告
- `target-lab`：故意含漏洞的演示靶场
- `tests`：API 冒烟（Mock 靶场，¥0）

## 7. API 模拟与切换指南

真实扫描通路始终接线：引擎对目标发起 HTTP 变异请求并由指纹匹配器判定。默认 `SCAN_MODE=lab` 只允许扫描内置靶场（`LAB_PUBLIC_URL` 会改写为容器内 `LAB_INTERNAL_URL`）。若要扫描其它已授权资产，将 `SCAN_MODE` 设为 `authorized` 并在前端勾选授权声明。QA 与演示一律打靶场，不产生外部费用。
