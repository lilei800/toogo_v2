package market

import (
	"math"
	"sync"
	"time"
)

// marketStateSmoother 市场状态平滑器（按 platform:symbol 维度缓存）
// 设计目标：
// - 抗抖：避免 1s 级别的状态频繁抖动
// - 轻量：只存最近 N 次，常数内存
// - 并发安全：同一 key 在并发分析时也不会破坏历史序列
type marketStateSmoother struct {
	mu sync.Mutex

	// timeframeHist 保存每个周期的历史状态（仅保留最近 window 条）
	timeframeHist map[string][]string

	// finalHist 保存最终市场状态历史（仅保留最近 window 条）
	finalHist []string

	updatedAt time.Time
}

func newMarketStateSmoother() *marketStateSmoother {
	return &marketStateSmoother{
		timeframeHist: make(map[string][]string),
		finalHist:     make([]string, 0, 16),
		updatedAt:     time.Now(),
	}
}

// getOrCreateStateSmoother 获取/创建指定 key 的平滑器
func (a *MarketAnalyzer) getOrCreateStateSmoother(key string) *marketStateSmoother {
	if v, ok := a.stateSmoothers.Load(key); ok {
		if s, ok2 := v.(*marketStateSmoother); ok2 && s != nil {
			return s
		}
		// 异常类型，清理后重建
		a.stateSmoothers.Delete(key)
	}
	s := newMarketStateSmoother()
	actual, _ := a.stateSmoothers.LoadOrStore(key, s)
	if s2, ok := actual.(*marketStateSmoother); ok && s2 != nil {
		return s2
	}
	return s
}

// pushAndSmoothTimeframe 追加单周期状态并返回平滑结果
func (s *marketStateSmoother) pushAndSmoothTimeframe(interval, state string, window int, minRatio float64) (smoothed string, conf float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := s.timeframeHist[interval]
	h = append(h, state)
	h = trimTail(h, window)
	s.timeframeHist[interval] = h
	s.updatedAt = time.Now()

	return smoothByMajority(state, h, minRatio)
}

// pushAndSmoothFinal 追加最终状态并返回平滑结果
func (s *marketStateSmoother) pushAndSmoothFinal(state string, window int, minRatio float64) (smoothed string, conf float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.finalHist = append(s.finalHist, state)
	s.finalHist = trimTail(s.finalHist, window)
	s.updatedAt = time.Now()

	return smoothByMajority(state, s.finalHist, minRatio)
}

func trimTail(history []string, window int) []string {
	if window <= 0 {
		return history
	}
	if len(history) <= window {
		return history
	}
	return history[len(history)-window:]
}

// smoothByMajority：
// - 在最近 window 的历史中统计出现最多的状态
// - 若占比 >= minRatio，则返回“多数状态”
// - 否则返回当前状态（保持灵敏度），但置信度会较低
func smoothByMajority(current string, history []string, minRatio float64) (smoothed string, conf float64) {
	if len(history) == 0 {
		return current, 1.0
	}

	counts := make(map[string]int, 4)
	maxState := current
	maxCount := 0
	for _, s := range history {
		counts[s]++
		if counts[s] > maxCount {
			maxCount = counts[s]
			maxState = s
		}
	}

	ratio := float64(maxCount) / float64(len(history))
	if minRatio <= 0 {
		minRatio = 0.6
	}

	need := int(math.Ceil(float64(len(history)) * minRatio))
	if maxCount >= need {
		return maxState, ratio
	}

	// 无共识：保持当前状态，但置信度使用“当前状态在窗口中的占比”
	curCount := counts[current]
	return current, float64(curCount) / float64(len(history))
}
