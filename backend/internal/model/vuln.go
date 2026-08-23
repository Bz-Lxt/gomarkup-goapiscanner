package model

type VulnClass string

const (
	ClassSQLi      VulnClass = "sql_injection"
	ClassTimeBlind VulnClass = "time_blind_sqli"
	ClassXSS       VulnClass = "xss"
	ClassUnauth    VulnClass = "unauthorized"
	ClassTraversal VulnClass = "path_traversal"
	ClassCMDi      VulnClass = "command_injection"
)

func (c VulnClass) Title() string {
	switch c {
	case ClassSQLi:
		return "SQL 注入"
	case ClassTimeBlind:
		return "时序盲注"
	case ClassXSS:
		return "跨站脚本 (XSS)"
	case ClassUnauth:
		return "未授权访问"
	case ClassTraversal:
		return "路径遍历"
	case ClassCMDi:
		return "命令注入"
	default:
		return string(c)
	}
}

func KnownClasses() []VulnClass {
	return []VulnClass{ClassSQLi, ClassTimeBlind, ClassXSS, ClassUnauth, ClassTraversal, ClassCMDi}
}
