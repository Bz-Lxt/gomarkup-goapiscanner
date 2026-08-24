package swagger

import "strings"

type ParamIn string

const (
	InQuery  ParamIn = "query"
	InPath   ParamIn = "path"
	InHeader ParamIn = "header"
	InBody   ParamIn = "body"
)

type Param struct {
	Name     string
	In       ParamIn
	Type     string
	Required bool
}

type Endpoint struct {
	Method      string
	Path        string
	OperationID string
	Params      []Param
}

func (e Endpoint) Key() string {
	return strings.ToUpper(e.Method) + " " + e.Path
}

func NormalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func JoinBase(basePath, path string) string {
	basePath = strings.TrimRight(basePath, "/")
	path = NormalizePath(path)
	if basePath == "" || basePath == "/" {
		return path
	}
	return basePath + path
}
