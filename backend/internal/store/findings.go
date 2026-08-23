package store

import "github.com/alkaid/goapiscanner/internal/model"

func (s *Store) InsertFinding(f model.Finding) error {
	_, err := s.db.Exec(`INSERT INTO findings(
		id,task_id,endpoint,method,class,severity,title,evidence,payload,param_name,
		status_code,latency_ms,advice,created_at
	) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		f.ID, f.TaskID, f.Endpoint, f.Method, f.Class, f.Severity, f.Title, f.Evidence, f.Payload, f.ParamName,
		f.StatusCode, f.LatencyMS, f.Advice, f.CreatedAt,
	)
	return err
}

func (s *Store) ListFindings(taskID string) ([]model.Finding, error) {
	rows, err := s.db.Query(`SELECT id,task_id,endpoint,method,class,severity,title,evidence,payload,param_name,
		status_code,latency_ms,advice,created_at
		FROM findings WHERE task_id=? ORDER BY created_at ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Finding, 0)
	for rows.Next() {
		var f model.Finding
		if err := rows.Scan(&f.ID, &f.TaskID, &f.Endpoint, &f.Method, &f.Class, &f.Severity, &f.Title, &f.Evidence, &f.Payload, &f.ParamName,
			&f.StatusCode, &f.LatencyMS, &f.Advice, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func BuildTree(findings []model.Finding) []model.DefectNode {
	type leaf struct {
		f model.Finding
	}
	groups := map[string][]model.Finding{}
	order := make([]string, 0)
	for _, f := range findings {
		key := f.Method + " " + f.Endpoint
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], f)
	}
	roots := make([]model.DefectNode, 0, len(order))
	for _, key := range order {
		fs := groups[key]
		node := model.DefectNode{
			Key:      key,
			Label:    key,
			Method:   fs[0].Method,
			Path:     fs[0].Endpoint,
			Children: make([]model.DefectNode, 0, len(fs)),
			Count:    len(fs),
		}
		best := model.SeverityInfo
		for _, f := range fs {
			cp := f
			child := model.DefectNode{
				Key:      f.ID,
				Label:    f.Title,
				Severity: f.Severity,
				Count:    1,
				Finding:  &cp,
				Children: []model.DefectNode{},
			}
			if f.Severity.Rank() > best.Rank() {
				best = f.Severity
			}
			node.Children = append(node.Children, child)
		}
		node.Severity = best
		roots = append(roots, node)
	}
	if roots == nil {
		roots = []model.DefectNode{}
	}
	return roots
}

func StatsOf(findings []model.Finding) model.SeverityStats {
	var s model.SeverityStats
	for _, f := range findings {
		switch f.Severity {
		case model.SeverityCritical:
			s.Critical++
		case model.SeverityHigh:
			s.High++
		case model.SeverityMedium:
			s.Medium++
		case model.SeverityLow:
			s.Low++
		default:
			s.Info++
		}
	}
	return s
}
