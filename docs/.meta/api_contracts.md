# API Contracts

> Contract Gate：无按量计费第三方。对端为自研 `target-lab`。
> 实呼时间：2026-08-23 23:03 GMT+8（compose 内网）

| Provider | 用途 | 状态 |
|---|---|---|
| target-lab `GET /health` | 健康 | verified · `{"status":"ok","lab":true}` |
| target-lab `GET /swagger.json` | OpenAPI 3 文档 | verified · 含 6 条故意漏洞路径 |
| target-lab 漏洞端点 | SQLi / 盲注 / XSS / 未授权 / 遍历 / 命令注入 | verified · 指纹召回 6/6 |
| 引擎 `/api/v1/health` | 健康 | verified |
| 引擎 `POST /api/v1/scans` | 需 `authorized=true` | verified · 缺省 400；成功 201 |
| 引擎 findings / report.pdf | 结果与 PDF | verified · `Content-Type: application/pdf` |

默认 `SCAN_MODE=lab`：仅允许改写后的靶场地址。`SCAN_MODE=authorized` 才接受用户勾选授权后的其它 URL（见 README §7）。
