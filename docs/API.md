# API 说明

基准路径：`/api/v1`  
时间字段：`yyyy-MM-dd HH:mm:ss`（GMT+8）  
统一信封：`{ "ok": true, "data": ..., "error": "" }`

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/health` | 健康检查 |
| GET | `/meta` | 默认靶场 URL、扫描模式 |
| POST | `/scans` | 创建并启动扫描 |
| GET | `/scans` | 任务列表 |
| GET | `/scans/{id}` | 任务详情 |
| GET | `/scans/{id}/findings` | 发现 + 缺陷树 |
| POST | `/scans/{id}/cancel` | 取消运行中任务 |
| GET | `/scans/{id}/report` | 报告 JSON |
| GET | `/scans/{id}/report.pdf` | 下载 PDF |
| GET | `/ws?task_id=` | WebSocket 实时事件 |

## POST /scans

JSON 示例：

```json
{
  "base_url": "http://localhost:28483",
  "concurrency": 16,
  "timeout_ms": 5000,
  "authorized": true
}
```

亦可 `multipart/form-data`：字段同上，外加文件字段 `swagger`。

成功：`201`，`data` 为任务对象。

错误码：
- `400` 未授权声明 / JSON 非法 / URL 非法
- `404` 任务不存在
- `409` 任务未在运行（取消时）
- `500` 落库或 PDF 失败

## WS 事件

`log` / `progress` / `finding` / `done`
