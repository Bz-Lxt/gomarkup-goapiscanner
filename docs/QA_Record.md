# QA Record

## Round 1 · 2026-08-23 23:03 (GMT+8)

**Cost**: ¥0（全程 Mock 靶场 / 离线，未调用任何计费 API）

**Environment**: `docker compose up --build -d` 后 `docker compose run --rm qa`

**Unit tests (host, then same packages in image build)**:
```
ok engine / fingerprint / payload / report / scan / store / swagger / api
```

**Smoke (`tests/api_smoke.py`)**:
```
....                                                                     [100%]
4 passed in 3.69s
```

- `[PASS] Docker Build`
- `[PASS] Health Check`（backend / lab / frontend 壳）
- `[PASS] 未授权扫描拒绝 400`
- `[PASS] 靶场召回 6/6`（sql_injection / time_blind_sqli / xss / unauthorized / path_traversal / command_injection）
- `[PASS] 报告 JSON + PDF %PDF`

**Browser**：控制台 → 已完成任务看板（258/258，严重 4 / 高危 2）→ 展开缺陷树证据 → 报告预览与下载 PDF 链接。

**结论**：PASS，进入审核。
