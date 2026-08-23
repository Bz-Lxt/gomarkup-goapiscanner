package scan

import (
	"context"
	"fmt"
	"time"

	"github.com/alkaid/goapiscanner/internal/clock"
	"github.com/alkaid/goapiscanner/internal/config"
	"github.com/alkaid/goapiscanner/internal/engine"
	"github.com/alkaid/goapiscanner/internal/fingerprint"
	"github.com/alkaid/goapiscanner/internal/logger"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/payload"
	"github.com/alkaid/goapiscanner/internal/store"
	"github.com/alkaid/goapiscanner/internal/swagger"
	"github.com/alkaid/goapiscanner/internal/ws"
	"github.com/google/uuid"
)

type Orchestrator struct {
	Cfg       config.Config
	Store     *store.Store
	Hub       *ws.Hub
	Guard     *Guard
	Canceller *Canceller
}

func NewOrchestrator(cfg config.Config, st *store.Store, hub *ws.Hub) *Orchestrator {
	return &Orchestrator{
		Cfg:       cfg,
		Store:     st,
		Hub:       hub,
		Guard:     NewGuard(),
		Canceller: NewCanceller(),
	}
}

func (o *Orchestrator) Start(task model.Task, spec []byte) {
	if !o.Guard.Acquire(task.ID) {
		o.emitLog(task.ID, "warn", "任务已在运行，忽略重入")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	o.Canceller.Bind(task.ID, cancel)
	go o.run(ctx, task, spec)
}

func (o *Orchestrator) Cancel(id string) bool {
	return o.Canceller.Cancel(id)
}

func (o *Orchestrator) run(ctx context.Context, task model.Task, spec []byte) {
	// A panic anywhere in the scan pipeline (e.g. a malformed OpenAPI doc that
	// trips a nil deref in the parser) would otherwise crash the whole process
	// and leave the task stuck in "running" forever. Recover, mark the task as
	// failed, and keep the engine alive for other scans.
	defer func() {
		if r := recover(); r != nil {
			logger.L().Error("scan panic recovered", "task", task.ID, "panic", fmt.Sprint(r))
			o.fail(task, "扫描引擎内部异常，请联系管理员")
		}
	}()
	defer o.Guard.Release(task.ID)
	defer o.Canceller.Drop(task.ID)

	task.Status = model.TaskRunning
	task.UpdatedAt = clock.NowString()
	_ = o.Store.UpdateTask(task)
	o.emitLog(task.ID, "info", "扫描启动 "+task.BaseURL)

	target, err := ResolveTarget(task.BaseURL, o.Cfg)
	if err != nil {
		o.fail(task, err.Error())
		return
	}
	o.emitLog(task.ID, "info", "目标改写为 "+target)

	if len(spec) == 0 {
		client := engine.NewClient(8 * time.Second)
		raw, src, ferr := FetchSwagger(ctx, client, target)
		if ferr != nil {
			o.fail(task, "无法获取 Swagger: "+ferr.Error())
			return
		}
		spec = raw
		task.SwaggerName = src
		o.emitLog(task.ID, "info", "已拉取接口文档 "+src)
	}

	parsed, err := swagger.Parse(spec)
	if err != nil {
		o.fail(task, "Swagger 解析失败: "+err.Error())
		return
	}
	o.emitLog(task.ID, "info", fmt.Sprintf("解析到 %d 个操作 (OpenAPI %s)", len(parsed.Endpoints), parsed.Version))

	jobs := payload.Mutate(parsed.Endpoints, payload.MutateOptions{BaseURL: target, MaxJobs: o.Cfg.MaxJobs})
	if len(jobs) == 0 {
		o.fail(task, "未生成任何变异请求")
		return
	}
	task.Total = len(jobs)
	task.UpdatedAt = clock.NowString()
	_ = o.Store.UpdateTask(task)
	o.emitProgress(task.ID, 0, task.Total, 0)
	o.emitLog(task.ID, "info", fmt.Sprintf("变异请求 %d 条，并发 %d", len(jobs), task.Concurrency))

	timeout := time.Duration(task.TimeoutMS) * time.Millisecond
	if timeout < 3*time.Second {
		timeout = 3 * time.Second
	}
	// timing payloads need headroom above SLEEP(3)
	if timeout < 6*time.Second {
		timeout = 6 * time.Second
	}
	pool := engine.NewPool(task.Concurrency, timeout)
	matcher := fingerprint.New()
	prog := &Progress{Total: len(jobs)}

	err = pool.Run(ctx, jobs, func(res engine.Result) {
		sent, total, hits := prog.AddSent()
		if sent%25 == 0 || sent == total {
			o.emitProgress(task.ID, sent, total, hits)
		}
		if res.Err != nil {
			if ctx.Err() != nil {
				return
			}
			return
		}
		probe := fingerprint.FromJob(res.Job, res.StatusCode, res.Header, res.Body, res.Latency)
		hit, ok := matcher.Match(probe)
		if !ok {
			return
		}
		f := fingerprint.FindingOf(task.ID, probe, hit, uuid.NewString(), clock.NowString())
		if err := o.Store.InsertFinding(f); err != nil {
			logger.L().Error("insert finding", "err", err.Error())
			return
		}
		_, _, hits = prog.AddHit(f.Severity.Rank())
		o.Hub.Broadcast(task.ID, model.StreamEvent{
			Type: model.EventFinding, TS: clock.NowString(), TaskID: task.ID, Finding: &f, Hits: hits, Sent: sent, Total: total,
		})
	})

	snap := prog.Snapshot()
	task.Sent = snap.Sent
	task.Hits = snap.Hits
	task.Critical = snap.Crit
	task.High = snap.High
	task.Medium = snap.Med
	task.Low = snap.Low
	task.Info = snap.Info
	task.UpdatedAt = clock.NowString()
	if ctx.Err() != nil {
		task.Status = model.TaskCancelled
		task.Error = "用户取消"
		_ = o.Store.UpdateTask(task)
		o.Hub.Broadcast(task.ID, model.StreamEvent{Type: model.EventDone, TS: clock.NowString(), TaskID: task.ID, Status: string(task.Status)})
		o.emitLog(task.ID, "warn", "扫描已取消")
		return
	}
	if err != nil {
		o.fail(task, err.Error())
		return
	}
	task.Status = model.TaskSucceeded
	_ = o.Store.UpdateTask(task)
	o.emitProgress(task.ID, task.Sent, task.Total, task.Hits)
	o.Hub.Broadcast(task.ID, model.StreamEvent{Type: model.EventDone, TS: clock.NowString(), TaskID: task.ID, Status: string(task.Status)})
	o.emitLog(task.ID, "info", fmt.Sprintf("扫描完成，命中 %d / %d", task.Hits, task.Total))
}

func (o *Orchestrator) fail(task model.Task, msg string) {
	task.Status = model.TaskFailed
	task.Error = msg
	task.UpdatedAt = clock.NowString()
	_ = o.Store.UpdateTask(task)
	o.emitLog(task.ID, "error", msg)
	o.Hub.Broadcast(task.ID, model.StreamEvent{Type: model.EventDone, TS: clock.NowString(), TaskID: task.ID, Status: string(task.Status), Message: msg})
}

func (o *Orchestrator) emitLog(taskID, level, msg string) {
	o.Hub.Broadcast(taskID, model.StreamEvent{
		Type: model.EventLog, TS: clock.NowString(), TaskID: taskID, Level: level, Message: msg,
	})
	logger.L().Info("scan", "task", taskID, "level", level, "msg", msg)
}

func (o *Orchestrator) emitProgress(taskID string, sent, total, hits int) {
	o.Hub.Broadcast(taskID, model.StreamEvent{
		Type: model.EventProgress, TS: clock.NowString(), TaskID: taskID, Sent: sent, Total: total, Hits: hits,
	})
}
