package openrule

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ruleEngine = make(map[string]*Rule)
)

//LoadConfig
func LoadConfig(scene string, defs []FieldDef) error {
	if len(defs) == 0 || scene == "" {
		return errors.New("no def")
	}
	if _, ok := ruleEngine[scene]; ok {
		return errors.New("already defined")
	}

	//每个字段都排好序
	fields := make(map[string]FieldDef)
	for _, def := range defs {
		if !isValidFieldType(def.Type) {
			return errors.New("not valid fieldType")
		}
		fields[def.Name] = def
	}

	r := &Rule{
		Name:      scene,
		FieldDefs: defs,
		Fields:    fields,
		Entities:  [2]RuleEntitySlice{},
		curIdx:    0,
	}
	ruleEngine[scene] = r
	return nil
}

//GetSceneRule
func GetSceneRule(scene string) *Rule {
	return ruleEngine[scene]
}

func (r *Rule) InsertSingleCond(entity *RuleEntity, cond Cond) error {
	fieldDef, ok := r.Fields[cond.Key]
	if !ok {
		panic(fmt.Errorf("fields not defined, condKey=%v", cond.Key))
	}
	if !isValidCond(fieldDef.Type, cond) {
		panic(fmt.Errorf("invalid FieldType, fieldType=%v", fieldDef.Type))
	}
	cond.priority = fieldDef.Priority //把优先级赋值，便于后期排序
	entity.Conds = append(entity.Conds, cond)
	return nil
}

//InsertRuleEntity
func (r *Rule) InsertRuleEntity(entity *RuleEntity) error {
	sort.Sort(entity)
	var unUseIdx int
	if r.curIdx == 0 { //插入时，插入另一个暂时不用的slice
		unUseIdx = 1
	}
	r.Entities[unUseIdx] = append(r.Entities[unUseIdx], entity)
	return nil
}

//MatchRules 规则命中的规则（多条）
func (r *Rule) MatchRules(fact map[string]interface{}) ([]*RuleEntity, error) {
	var matchEntities []*RuleEntity
	usingIdx := r.curIdx //idx表示目前正在用的slice
	if r == nil || len(r.Entities[usingIdx]) == 0 {
		return nil, errors.New("rule entity empty,check")
	}
	for _, entity := range r.Entities[usingIdx] {
		if len(entity.Conds) == 0 { //兜底规则，没有条件，默认匹配上了
			matchEntities = append(matchEntities, entity)
			continue
		}
		var match bool
		var err error
		for _, cond := range entity.Conds {
			factVal, ok := fact[cond.Key]
			if !ok {
				return nil, fmt.Errorf("fact not contain val[%v]", cond.Key)
			}
			fieldDef := r.Fields[cond.Key]
			match, err = matchSingleCond(factVal, fieldDef.Type, cond)
			if err != nil {
				return nil, err
			}
			if !match {
				break
			}
		}
		if match {
			matchEntities = append(matchEntities, entity)
		}
	}
	return matchEntities, nil
}

//GetWinnerRuleEntity 根据优先级 取出最高优先级的规则
func GetWinnerRuleEntity(entities []*RuleEntity) (*RuleEntity, error) {
	if len(entities) == 0 {
		return nil, errors.New("entities nil")
	}
	winner := &RuleEntity{}
	var priority = -1
	for _, entity := range entities {
		if entity.Priority > priority {
			priority = entity.Priority
			winner = entity
		}
	}
	return winner, nil
}

//FinishLoad 0 -> 1  1-> 0
func FinishLoad(scene string) {
	var unUseIdx int
	if ruleEngine[scene].curIdx == 0 {
		unUseIdx = 1
	}
	sort.Sort(ruleEngine[scene].Entities[unUseIdx])
	ruleEngine[scene].curIdx = unUseIdx
	return
}

func matchSingleCond(fieldVal interface{}, fieldType FieldType, cond Cond) (bool, error) {
	var flag bool
	switch cond.Opr {
	case IN, NOT_IN:
		if fieldType == FieldTypeInt {
			fieldValNum, ok := fieldVal.(int)
			if !ok {
				return false, fmt.Errorf("key[%v] should be int", cond.Key)
			}
			flag = numInSet(fieldValNum, cond.Val.ValNumSet)
		}
		if fieldType == FieldTypeString {
			fieldValStr, ok := fieldVal.(string)
			if !ok {
				return false, fmt.Errorf("key[%v] should be string", cond.Key)
			}
			flag = strInSet(fieldValStr, cond.Val.ValStrSet)
		}
		if cond.Opr == NOT_IN {
			return !flag, nil
		}
		return flag, nil
	case EQ, NOT_EQ:
		if fieldType == FieldTypeInt {
			fieldValNum, ok := fieldVal.(int)
			if !ok {
				return false, fmt.Errorf("key[%v] should be int", cond.Key)
			}
			flag = fieldValNum == *cond.Val.ValNum
		}
		if fieldType == FieldTypeString {
			fieldValStr, ok := fieldVal.(string)
			if !ok {
				return false, fmt.Errorf("key[%v] should be string", cond.Key)
			}
			flag = fieldValStr == *cond.Val.ValStr
		}
		if cond.Opr == NOT_EQ {
			return !flag, nil
		}
		return flag, nil
	case LT, GT:
		fieldValNum, ok := fieldVal.(int)
		if !ok {
			return false, fmt.Errorf("key[%v] should be int", cond.Key)
		}
		flag = fieldValNum > *cond.Val.ValNum
		if cond.Opr == LT {
			return !flag, nil
		}
		return flag, nil
	case LE, GE:
		fieldValNum, ok := fieldVal.(int)
		if !ok {
			return false, fmt.Errorf("key[%v] should be int", cond.Key)
		}
		flag = fieldValNum >= *cond.Val.ValNum
		if cond.Opr == LT {
			return !flag, nil
		}
		return flag, nil
	case INTERSECT, NOT_INTERSECT:
		if fieldType == FieldTypeInt {
			fieldValNums, ok := fieldVal.([]int)
			if !ok {
				return false, fmt.Errorf("key[%v] should be []int", cond.Key)
			}
			for _, num := range fieldValNums {
				if _, ok = cond.Val.ValNumSet[num]; ok {
					flag = true
					break
				}
			}
		}
		if fieldType == FieldTypeString {
			fieldValStrs, ok := fieldVal.([]string)
			if !ok {
				return false, fmt.Errorf("key[%v] should be []string", cond.Key)
			}
			for _, str := range fieldValStrs {
				if _, ok = cond.Val.ValStrSet[str]; ok {
					flag = true
					break
				}
			}
		}
		if cond.Opr == NOT_INTERSECT {
			return !flag, nil
		}
		return flag, nil
	default:
		return false, fmt.Errorf("unknown opr [%v]", cond.Opr)
	}
}
