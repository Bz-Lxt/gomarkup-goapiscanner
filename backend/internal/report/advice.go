package report

import "github.com/alkaid/goapiscanner/internal/model"

func AdviceFor(c model.VulnClass) string {
	switch c {
	case model.ClassSQLi:
		return "使用参数化查询 / PreparedStatement，禁止拼接 SQL；对输入做白名单校验，最小权限连接数据库。"
	case model.ClassTimeBlind:
		return "关闭数据库延时函数对不可信输入的可达路径；统一查询超时；与报错注入一并改为参数绑定。"
	case model.ClassXSS:
		return "按上下文做 HTML/JS/URL 编码输出；启用 Content-Security-Policy；避免将用户输入写入 innerHTML。"
	case model.ClassUnauth:
		return "默认拒绝：管理接口必须鉴权与鉴权失败 401/403；补充对象级授权，禁止仅靠隐藏 URL。"
	case model.ClassTraversal:
		return "禁止直接拼接用户路径；使用白名单文件 ID；realpath 后校验前缀落在允许目录内。"
	case model.ClassCMDi:
		return "禁止把用户输入交给 shell；使用参数数组执行；必要命令改为内部 API。"
	default:
		return "按最小权限与输入校验原则修复，并补充回归测试。"
	}
}

func AdviceList(findings []model.Finding) []string {
	seen := map[model.VulnClass]struct{}{}
	out := make([]string, 0, 8)
	for _, f := range findings {
		if _, ok := seen[f.Class]; ok {
			continue
		}
		seen[f.Class] = struct{}{}
		out = append(out, f.Class.Title()+"："+AdviceFor(f.Class))
	}
	if len(out) == 0 {
		out = append(out, "本次扫描未命中已知指纹，建议扩大接口覆盖后复测。")
	}
	return out
}
