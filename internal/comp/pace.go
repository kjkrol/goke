package comp

import (
	"fmt"
	"reflect"
)

// Pace is how a query walked by its pace goes over its blocks: the runs numbered by Step, a block
// carrying a ruled component going as its rule says, every other block every run.
type Pace struct {
	Step  func() uint64
	Rules []PaceRule
}

// PaceRule is the period of the blocks carrying the component ID: every Period-th run, none at 0.
type PaceRule struct {
	ID     ID
	Period uint64
}

// PaceRuleOpt resolves a rule against the index of component types.
type PaceRuleOpt func(*DefIndex) (PaceRule, error)

// PacedBy paces the query by step and rules; a query takes one pace.
func PacedBy(step func() uint64, rules ...PaceRuleOpt) AccessOpt {
	return func(s *AccessSpec, mi *DefIndex) error {
		if s.Pace != nil {
			return fmt.Errorf("the query is paced already")
		}
		if step == nil {
			return fmt.Errorf("a pace without a step")
		}
		p := &Pace{Step: step}
		for _, opt := range rules {
			rule, err := opt(mi)
			if err != nil {
				return err
			}
			p.Rules = append(p.Rules, rule)
		}
		s.Pace = p
		return nil
	}
}

// Every is the rule of the blocks carrying T: every n-th run.
func Every[T any](n uint64) PaceRuleOpt {
	return func(mi *DefIndex) (PaceRule, error) {
		t := reflect.TypeFor[T]()
		if n == 0 {
			return PaceRule{}, fmt.Errorf("%s every 0 runs: Never is the rule of none", t)
		}
		return PaceRule{ID: mi.Intern(t).ID, Period: n}, nil
	}
}

// Never is the rule of the blocks carrying T: no run.
func Never[T any]() PaceRuleOpt {
	return func(mi *DefIndex) (PaceRule, error) {
		return PaceRule{ID: mi.Intern(reflect.TypeFor[T]()).ID}, nil
	}
}

// Period is the period of a block of mask: 0 where a rule says never, else the longest of the
// rules it meets, 1 where it meets none or there is no pace.
func (p *Pace) Period(mask Mask) uint64 {
	period := uint64(1)
	if p == nil {
		return period
	}
	for _, r := range p.Rules {
		if !mask.IsSet(r.ID) {
			continue
		}
		if r.Period == 0 {
			return 0
		}
		period = max(period, r.Period)
	}
	return period
}
