package openrule

import (
	"strconv"
	"strings"
)

func ConvStrToMap(str string, sep string) map[string]struct{} {
	res := make(map[string]struct{})
	splits := strings.Split(str, sep)
	for _, split := range splits {
		res[split] = struct{}{}
	}
	return res
}

func ConvStrToIntMap(str string, sep string) map[int]struct{} {
	res := make(map[int]struct{})
	splits := strings.Split(str, sep)
	for _, split := range splits {
		splitNum, _ := strconv.Atoi(split)
		res[splitNum] = struct{}{}
	}
	return res
}

func isValidCond(fieldType FieldType, cond Cond) bool {
	switch cond.Opr {
	case EQ, NOT_EQ:
		//only these
		switch fieldType {
		case FieldTypeString:
			if cond.Val.ValStr == nil {
				return false
			}
		case FieldTypeInt:
			if cond.Val.ValNum == nil {
				return false
			}
		default:
			return false
		}
	case IN, NOT_IN:
		switch fieldType {
		case FieldTypeString:
			if len(cond.Val.ValStrSet) == 0 {
				return false
			}
		case FieldTypeInt:
			if len(cond.Val.ValNumSet) == 0 {
				return false
			}
		default:
			return false
		}
	case GT, GE, LT, LE:
		if fieldType != FieldTypeInt || cond.Val.ValNum == nil {
			return false
		}
	case INTERSECT, NOT_INTERSECT:
		if fieldType == FieldTypeString && len(cond.Val.ValStrSet) == 0 ||
			fieldType == FieldTypeInt && len(cond.Val.ValNumSet) == 0 {
			return false
		}
	}
	return true
}

func isValidFieldType(fieldType FieldType) bool {
	if fieldType == FieldTypeInt || fieldType == FieldTypeString {
		return true
	}
	return false
}

func numInSet(v int, vals map[int]struct{}) bool {
	_, ok := vals[v]
	return ok
}

func strInSet(v string, vals map[string]struct{}) bool {
	_, ok := vals[v]
	return ok
}
