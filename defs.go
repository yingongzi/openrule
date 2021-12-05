package openrule

type Operator string

const (
	IN            Operator = "in"
	NOT_IN                 = "not in"
	EQ                     = "="
	NOT_EQ                 = "!="
	GT                     = ">"
	GE                     = ">="
	LT                     = "<"
	LE                     = "<="
	INTERSECT              = "intersect"
	NOT_INTERSECT          = "not intersect"
)

//CondVal 为了便于比较，只有 int 和 string 类型,得用指针，判断有没有设置值,同时只会有一个有值
type CondVal struct {
	ValNum    *int                `json:"val_num,omitempty"`
	ValNumSet map[int]struct{}    `json:"val_num_set,omitempty"`
	ValStr    *string             `json:"val_str,omitempty"`
	ValStrSet map[string]struct{} `json:"val_str_set,omitempty"`
}

//Cond
type Cond struct {
	Key      string // tag
	Opr      Operator
	Val      CondVal
	priority int
}

type RuleEntity struct {
	Conds     []Cond
	ID        int
	ExtraData map[string]interface{}
	Priority  int //优先级，正数
}

type RuleEntitySlice []*RuleEntity

type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeInt    FieldType = "int"
)

//FieldDef 定义好每个字段的类型和优先级
type FieldDef struct {
	Name     string
	Type     FieldType
	Priority int //条件的优先级，用户自定义，高的先进行匹配
}

type Rule struct {
	Name      string
	Initialed bool
	FieldDefs []FieldDef          //优先级排序后的Demands
	Fields    map[string]FieldDef //defs里面的名称
	Entities  [2]RuleEntitySlice  //双buffer，保证并发安全
	curIdx    int
}

//将entity的condition按照优先级排序
func (e *RuleEntity) Len() int      { return len(e.Conds) }
func (e *RuleEntity) Swap(i, j int) { e.Conds[i], e.Conds[j] = e.Conds[j], e.Conds[i] }

// 按照间隔进行排序
func (e *RuleEntity) Less(i, j int) bool {
	return e.Conds[i].priority < e.Conds[j].priority
}

//将entity的condition按照优先级排序
func (e RuleEntitySlice) Len() int      { return len(e) }
func (e RuleEntitySlice) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

// 按照间隔进行排序
func (e RuleEntitySlice) Less(i, j int) bool {
	return e[i].Priority < e[j].Priority
}
