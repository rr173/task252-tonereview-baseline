package service

import (
	"task252-tonereview/internal/capture"
	"task252-tonereview/internal/model"
	"task252-tonereview/internal/normalize"
)

// speakerBaselineOf 由可用人声段样本计算说话人基线（log2(F0) 均值）。
func speakerBaselineOf(samples []model.F0Sample) (float64, bool) {
	return normalize.SpeakerBaseline(samples)
}

// toneTypeOfSamples 在已知基线时对片段分类调型。
func toneTypeOfSamples(samples []model.F0Sample, hasBaseline bool, baselineLog float64) string {
	return capture.ToneTypeOf(samples, hasBaseline, baselineLog)
}
