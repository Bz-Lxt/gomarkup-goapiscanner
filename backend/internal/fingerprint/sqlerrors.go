package fingerprint

// CommonSQLErrors is a hand-maintained signature table used by the matcher.
// Entries are compared case-insensitively against response bodies.
var CommonSQLErrors = []string{
	"you have an error in your sql syntax",
	"sql_error",
	"warning: mysql_",
	"mysql_fetch",
	"mysqli_sql_exception",
	"unclosed quotation mark after the character string",
	"quoted string not properly terminated",
	"sqlite error",
	"sqlite3.operationalerror",
	"pg::syntaxerror",
	"postgresql query failed",
	"org.postgresql.util.psqlexception",
	"ora-01756",
	"ora-00933",
	"ora-00936",
	"microsoft ole db provider for sql server",
	"odbc sql server driver",
	"jdbc.sql.sqlexception",
	"syntax error at or near",
	"unterminated quoted string",
	"sqlstate[",
	"sql syntax",
}

func LooksLikeSQLError(body string) string {
	return bodyHasAny(body, CommonSQLErrors)
}
