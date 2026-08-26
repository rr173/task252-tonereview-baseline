package httpapi

import (
	"net/http"

	"task252-tonereview/internal/service"
)

// New 构造 API（持有服务编排层）。
func New(svc *service.Service) *API { return &API{svc: svc} }

// Handler 返回注册全部 /api 路由的 http.Handler。
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()

	// 批次
	m.HandleFunc("POST /api/batches", a.handleCreateBatch)
	m.HandleFunc("GET /api/batches", a.handleListBatches)
	m.HandleFunc("GET /api/batches/{id}", a.handleGetBatch)
	m.HandleFunc("GET /api/batches/{id}/summary", a.handleBatchSummary)
	m.HandleFunc("POST /api/batches/{id}/review", a.handleStartReview)
	m.HandleFunc("POST /api/batches/{id}/publish", a.handlePublishBatch)
	m.HandleFunc("POST /api/batches/{id}/seal", a.handleSealBatch)

	// 说话人
	m.HandleFunc("POST /api/speakers", a.handleCreateSpeaker)
	m.HandleFunc("GET /api/speakers", a.handleListSpeakers)
	m.HandleFunc("GET /api/speakers/{id}", a.handleGetSpeaker)
	m.HandleFunc("POST /api/speakers/{id}/baseline", a.handleRecomputeBaseline)

	// 发音片段
	m.HandleFunc("POST /api/batches/{id}/segments", a.handleImportSegment)
	m.HandleFunc("GET /api/segments", a.handleListSegments)
	m.HandleFunc("GET /api/segments/{id}", a.handleGetSegment)
	m.HandleFunc("POST /api/segments/{id}/verify", a.handleVerifySegment)
	m.HandleFunc("POST /api/segments/{id}/noise", a.handleMarkNoise)
	m.HandleFunc("POST /api/segments/{id}/exclude", a.handleExcludeSegment)
	m.HandleFunc("POST /api/segments/{id}/restore", a.handleRestoreSegment)
	m.HandleFunc("POST /api/segments/{id}/f0", a.handleAddF0)
	m.HandleFunc("GET /api/segments/{id}/f0", a.handleListF0)

	// 声调对立
	m.HandleFunc("POST /api/oppositions", a.handleCreateOpposition)
	m.HandleFunc("GET /api/oppositions", a.handleListOppositions)
	m.HandleFunc("GET /api/oppositions/{id}", a.handleGetOpposition)
	m.HandleFunc("POST /api/oppositions/{id}/evidence", a.handleAddEvidence)
	m.HandleFunc("POST /api/oppositions/{id}/recompute", a.handleRecomputeCluster)
	m.HandleFunc("GET /api/oppositions/{id}/cluster", a.handleGetCluster)
	m.HandleFunc("POST /api/oppositions/{id}/adjudicate", a.handleAdjudicate)
	m.HandleFunc("POST /api/compare", a.handleCompare)

	// 分析版本
	m.HandleFunc("POST /api/versions", a.handleCreateVersion)
	m.HandleFunc("POST /api/versions/{id}/share", a.handleShareVersion)
	m.HandleFunc("POST /api/versions/{id}/freeze", a.handleFreezeVersion)
	m.HandleFunc("GET /api/versions", a.handleListVersions)
	m.HandleFunc("GET /api/versions/{id}", a.handleGetVersion)

	// 自检 / 统计
	m.HandleFunc("GET /api/health", a.handleHealth)
	m.HandleFunc("GET /api/stats", a.handleStats)
	m.HandleFunc("GET /api/selfcheck", a.handleSelfCheck)

	return m
}
