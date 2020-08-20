package filter

import "github.com/pkg/errors"

type Item interface {
	TypeName() string
}

type Rule interface {
	Check(Item) (bool,error)
}

type Filter interface {
	CheckAll(Item) (bool,error)
}

type ItemFilter struct {
	MustOkRuleList        []Rule
	MustOkRuleListUseCode []bool
	OneOkRuleList         []Rule
	OneOkRuleListUseCode  []bool
}

func (tp ItemFilter) CheckAll(item Item) (bool,error) {
	ok, err := tp.CheckCNF(item,tp.MustOkRuleList,tp.MustOkRuleListUseCode)
	if err != nil {
		return false,err
	}
	if !ok {
		return false,nil
	}
	ok, err = tp.CheckDNF(item,tp.OneOkRuleList,tp.OneOkRuleListUseCode)
	if err != nil {
		return false,err
	}
	if !ok {
		return false,nil
	}
	return true,nil
}

func (tp ItemFilter) CheckCNF(item Item,mustOkRuleList []Rule,mustOkRuleListUselist []bool) (bool,error) {
	for index, tr := range mustOkRuleList {
		if mustOkRuleListUselist[index] == false {
			continue
		}
		ok,err := tr.Check(item)
		if err != nil {
			return false,errors.WithMessage(err, "Pod CNF Check failed;")
		}
		if !ok {
			return false,nil
		}
	}
	return true, nil
}

func (tp ItemFilter) CheckDNF(item Item,oneOkRuleList []Rule,oneOkRuleListUseList []bool) (bool,error) {
	checked := false
	for index, tr := range oneOkRuleList {
		if oneOkRuleListUseList[index] == false {
			continue
		}
		ok,err := tr.Check(item)
		checked = true
		if err != nil {
			return false,errors.WithMessage(err, "Pod DNF Check failed;")
		}
		if ok {
			return true,nil
		}
	}
	if checked {
		return false, nil
	} else {
		return true,nil
	}
}