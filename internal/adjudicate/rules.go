// Package adjudicate 实现声调对立的裁决规则：基于对立得分的建议状态，
// 以及状态流转合法性校验。
package adjudicate

import "task252-tonereview/internal/model"

// 对立得分阈值：得分越高代表组间差异相对组内差异越大，越可能是真对立。
const (
	confirmThreshold  = 1.6 // 达到此值建议「确认」
	candidateCeil     = 1.05 // 高于此值但不足确认阈值，维持「候选」
	// 低于 candidateCeil 视为「证据不足」
)

// SuggestStatus 根据对立得分给出建议状态（不改变既有终态）。
func SuggestStatus(score float64) string {
	switch {
	case score >= confirmThreshold:
		return model.OppConfirmed
	case score >= candidateCeil:
		return model.OppCandidate
	default:
		return model.OppInsufficient
	}
}

// CanTransition 校验声调对立状态流转合法性。
// 候选/证据不足可相互流转或转为确认/否决；确认/否决为终态，不可再改。
func CanTransition(from, to string) bool {
	switch from {
	case model.OppCandidate, model.OppInsufficient:
		return to == model.OppConfirmed || to == model.OppRejected ||
			to == model.OppCandidate || to == model.OppInsufficient
	case model.OppConfirmed, model.OppRejected:
		return false
	}
	return false
}

// BatchTransitions 给定批次状态，返回允许的下一步（用于状态机守卫）。
func BatchTransitions(from string) []string {
	switch from {
	case model.BatchCollecting:
		return []string{model.BatchReviewing}
	case model.BatchReviewing:
		return []string{model.BatchPublished}
	case model.BatchPublished:
		return []string{model.BatchSealed}
	case model.BatchSealed:
		return []string{}
	}
	return []string{}
}

// CanBatchTransition 校验批次状态流转。
func CanBatchTransition(from, to string) bool {
	for _, n := range BatchTransitions(from) {
		if n == to {
			return true
		}
	}
	return false
}

// VersionTransitions 版本状态流转守卫：草稿→共享→冻结。
func CanVersionTransition(from, to string) bool {
	switch from {
	case model.VerDraft:
		return to == model.VerShared
	case model.VerShared:
		return to == model.VerFrozen
	case model.VerFrozen, model.VerSuperseded:
		return false
	}
	return false
}
