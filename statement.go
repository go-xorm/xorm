package xorm

import (
	"fmt"
)

type Statement struct {
	Table      *Table
	Session    *Session
	Start      int
	LimitN     int
	WhereStr   string
	Params     []interface{}
	OrderStr   string
	JoinStr    string
	GroupByStr string
	HavingStr  string
	ColumnStr  string
	BeanArgs   []interface{}
}

func (statement *Statement) Init() {
	statement.Table = nil
	statement.Session = nil
	statement.Start = 0
	statement.LimitN = 0
	statement.WhereStr = ""
	statement.Params = make([]interface{}, 0)
	statement.OrderStr = ""
	statement.JoinStr = ""
	statement.GroupByStr = ""
	statement.HavingStr = ""
	statement.ColumnStr = ""
	statement.BeanArgs = make([]interface{}, 0)
}

func (statement *Statement) Where(querystring string, args ...interface{}) {
	statement.WhereStr = querystring
	statement.Params = args
}

func (statement *Statement) Limit(limit int, start ...int) {
	statement.LimitN = limit
	if len(start) > 0 {
		statement.Start = start[0]
	}
}

func (statement *Statement) OrderBy(order string) {
	statement.OrderStr = order
}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (statement *Statement) Join(join_operator, tablename, condition string) {
	if statement.JoinStr != "" {
		statement.JoinStr = statement.JoinStr + fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	} else {
		statement.JoinStr = fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	}
}

func (statement *Statement) GroupBy(keys string) {
	statement.GroupByStr = fmt.Sprintf("GROUP BY %v", keys)
}

func (statement *Statement) Having(conditions string) {
	statement.HavingStr = fmt.Sprintf("HAVING %v", conditions)
}

func (statement Statement) generateSql() string {
	columnStr := statement.Table.ColumnStr()
	return statement.genSelectSql(columnStr)
}

func (statement Statement) genCountSql() string {
	return statement.genSelectSql("count(*) as total")
}

func (statement Statement) genExecSql() string {
	return ""
}

func (statement Statement) genSelectSql(columnStr string) (a string) {
	session := statement.Session
	if session.Engine.Protocol == "mssql" {
		if statement.Start > 0 {
			a = fmt.Sprintf("select ROW_NUMBER() OVER(order by %v )as rownum,%v from %v",
				statement.Table.PKColumn().Name,
				columnStr,
				statement.Table.Name)
			if statement.WhereStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
				if statement.ColumnStr != "" {
					a = fmt.Sprintf("%v and %v", a, statement.ColumnStr)
				}
			} else if statement.ColumnStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.ColumnStr)
			}
			a = fmt.Sprintf("select %v from (%v) "+
				"as a where rownum between %v and %v",
				columnStr,
				a,
				statement.Start,
				statement.LimitN)
		} else if statement.LimitN > 0 {
			a = fmt.Sprintf("SELECT top %v %v FROM %v", statement.LimitN, columnStr, statement.Table.Name)
			if statement.WhereStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
				if statement.ColumnStr != "" {
					a = fmt.Sprintf("%v and %v", a, statement.ColumnStr)
				}
			} else if statement.ColumnStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.ColumnStr)
			}
			if statement.GroupByStr != "" {
				a = fmt.Sprintf("%v %v", a, statement.GroupByStr)
			}
			if statement.HavingStr != "" {
				a = fmt.Sprintf("%v %v", a, statement.HavingStr)
			}
			if statement.OrderStr != "" {
				a = fmt.Sprintf("%v ORDER BY %v", a, statement.OrderStr)
			}
		} else {
			a = fmt.Sprintf("SELECT %v FROM %v", columnStr, statement.Table.Name)
			if statement.WhereStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
				if statement.ColumnStr != "" {
					a = fmt.Sprintf("%v and %v", a, statement.ColumnStr)
				}
			} else if statement.ColumnStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.ColumnStr)
			}
			if statement.GroupByStr != "" {
				a = fmt.Sprintf("%v %v", a, statement.GroupByStr)
			}
			if statement.HavingStr != "" {
				a = fmt.Sprintf("%v %v", a, statement.HavingStr)
			}
			if statement.OrderStr != "" {
				a = fmt.Sprintf("%v ORDER BY %v", a, statement.OrderStr)
			}
		}
	} else {
		a = fmt.Sprintf("SELECT %v FROM %v", columnStr, statement.Table.Name)
		if statement.JoinStr != "" {
			a = fmt.Sprintf("%v %v", a, statement.JoinStr)
		}
		if statement.WhereStr != "" {
			a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
			if statement.ColumnStr != "" {
				a = fmt.Sprintf("%v and %v", a, statement.ColumnStr)
			}
		} else if statement.ColumnStr != "" {
			a = fmt.Sprintf("%v WHERE %v", a, statement.ColumnStr)
		}
		if statement.GroupByStr != "" {
			a = fmt.Sprintf("%v %v", a, statement.GroupByStr)
		}
		if statement.HavingStr != "" {
			a = fmt.Sprintf("%v %v", a, statement.HavingStr)
		}
		if statement.OrderStr != "" {
			a = fmt.Sprintf("%v ORDER BY %v", a, statement.OrderStr)
		}
		if statement.Start > 0 {
			a = fmt.Sprintf("%v LIMIT %v, %v", a, statement.Start, statement.LimitN)
		} else if statement.LimitN > 0 {
			a = fmt.Sprintf("%v LIMIT %v", a, statement.LimitN)
		}
	}
	return
}

/*func (statement *Statement) genInsertSQL() string {
	table = statement.Table
	colNames := make([]string, len(table.Columns))
	for idx, col := range table.Columns {
		if col.Name == "" {
			continue
		}
		colNames[idx] = col.Name
	}
	return strings.Join(colNames, ", ")

	colNames := make([]string, len(table.Columns))
	for idx, col := range table.Columns {
		if col.Name == "" {
			continue
		}
		colNames[idx] = "?"
	}
	strings.Join(colNames, ", ")
}*/
