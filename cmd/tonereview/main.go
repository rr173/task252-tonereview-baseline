// 命令入口：濒危语言声调证据复核台。
// 提供两种运行模式：
//   - 服务模式：--addr :端口 --db 路径，启动 /api HTTP 服务（JSON 复核接口）。
//   - 自检模式：--smoke-test，真实创建批次→导入片段→归一化基线→构建证据簇→
//     裁决对立→冻结版本→封存批次→关闭并重开数据库验证持久化与重启恢复，最后以 0 退出。
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"task252-tonereview/internal/capture"
	"task252-tonereview/internal/httpapi"
	"task252-tonereview/internal/service"
	"task252-tonereview/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP 监听地址")
	dbPath := flag.String("db", "tonoreview.db", "SQLite 数据库路径")
	smoke := flag.Bool("smoke-test", false, "运行自检后退出，验证持久化与重启恢复")
	flag.Parse()

	if *smoke {
		if err := runSmokeTest(); err != nil {
			log.Fatalf("smoke-test FAILED: %v", err)
		}
		fmt.Println("smoke-test PASSED")
		return
	}

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	api := httpapi.New(svc)
	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler())
	// 极简根路径：以 JSON 形式列出可用端点（非前端页面，仅便捷入口）。
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, `{"service":"task252-tonereview","note":"声调证据复核后端服务，端点前缀 /api","examples":["POST /api/batches","POST /api/speakers","POST /api/batches/{id}/segments","POST /api/oppositions/{id}/evidence","POST /api/oppositions/{id}/adjudicate","POST /api/versions/{id}/freeze","GET /api/selfcheck"]}`)
	})

	server := &http.Server{Addr: *addr, Handler: mux}
	fmt.Printf("task252-tonereview listening on %s (db=%s)\n", *addr, *dbPath)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// genF0 生成基频采样序列（时间严格递增，频率线性插值）。
func genF0(startMs, stepMs int64, n int, f0Start, f0End float64) []capture.F0Point {
	out := make([]capture.F0Point, 0, n)
	for i := 0; i < n; i++ {
		t := startMs + int64(i)*stepMs
		f := f0Start + (f0End-f0Start)*float64(i)/float64(n-1)
		out = append(out, capture.F0Point{TMs: t, F0Hz: f})
	}
	return out
}

// runSmokeTest 验证完整业务闭环与重启恢复。
func runSmokeTest() error {
	dir, err := os.MkdirTemp("", "tonoreview-smoke")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	dbFile := dir + "/smoke.db"

	// ---- 第一次会话：创建→导入→归一化→证据→裁决→冻结→封存 ----
	st, err := store.Open(dbFile)
	if err != nil {
		return err
	}
	svc := service.New(st)

	// 两个说话人：A 发高平调，B 发升调。
	spA, err := svc.CreateSpeaker("spk-A", "dialect-X", "f", 1950)
	if err != nil {
		return fmt.Errorf("create speaker A: %w", err)
	}
	spB, err := svc.CreateSpeaker("spk-B", "dialect-X", "m", 1948)
	if err != nil {
		return fmt.Errorf("create speaker B: %w", err)
	}

	batch, err := svc.CreateBatch("smoke-batch", "烟雾测试：同音段 ma 的高低平/升对立")
	if err != nil {
		return fmt.Errorf("create batch: %w", err)
	}
	bid := batch.ID

	// A 侧（高平调）两个相似片段；B 侧（升调）两个相似片段。
	segA1, err := svc.ImportSegment(bid, capture.SegmentInput{
		LexicalItem: "ma-high", PhoneticSeg: "ma", SpeakerID: spA.ID, AudioFP: "fp-a1",
		DurationMs: 400, RecordedAt: 1, F0: genF0(0, 50, 8, 200, 200),
	})
	if err != nil {
		return fmt.Errorf("import a1: %w", err)
	}
	segA2, err := svc.ImportSegment(bid, capture.SegmentInput{
		LexicalItem: "ma-high", PhoneticSeg: "ma", SpeakerID: spA.ID, AudioFP: "fp-a2",
		DurationMs: 400, RecordedAt: 2, F0: genF0(0, 50, 8, 202, 198),
	})
	if err != nil {
		return fmt.Errorf("import a2: %w", err)
	}
	segB1, err := svc.ImportSegment(bid, capture.SegmentInput{
		LexicalItem: "ma-rising", PhoneticSeg: "ma", SpeakerID: spB.ID, AudioFP: "fp-b1",
		DurationMs: 400, RecordedAt: 3, F0: genF0(0, 50, 8, 150, 220),
	})
	if err != nil {
		return fmt.Errorf("import b1: %w", err)
	}
	segB2, err := svc.ImportSegment(bid, capture.SegmentInput{
		LexicalItem: "ma-rising", PhoneticSeg: "ma", SpeakerID: spB.ID, AudioFP: "fp-b2",
		DurationMs: 400, RecordedAt: 4, F0: genF0(0, 50, 8, 155, 215),
	})
	if err != nil {
		return fmt.Errorf("import b2: %w", err)
	}

	// 可用化片段并重建说话人基线（触发调型分类）。
	for _, id := range []string{segA1.ID, segA2.ID, segB1.ID, segB2.ID} {
		if err := svc.VerifySegment(id); err != nil {
			return fmt.Errorf("verify %s: %w", id, err)
		}
	}
	if err := svc.RecomputeBaseline(spA.ID); err != nil {
		return fmt.Errorf("baseline A: %w", err)
	}
	if err := svc.RecomputeBaseline(spB.ID); err != nil {
		return fmt.Errorf("baseline B: %w", err)
	}
	fmt.Println("  [smoke] baselines computed; tone types classified")

	// 创建候选对立并加入两侧证据。
	opp, err := svc.CreateOpposition(bid, "ma-high", "ma", "ma-rising")
	if err != nil {
		return fmt.Errorf("create opposition: %w", err)
	}
	for _, e := range []struct {
		seg, side string
	}{
		{segA1.ID, "a"}, {segA2.ID, "a"}, {segB1.ID, "b"}, {segB2.ID, "b"},
	} {
		if err := svc.AddEvidence(opp.ID, e.seg, e.side); err != nil {
			return fmt.Errorf("add evidence %s/%s: %w", e.seg, e.side, err)
		}
	}
	stats, err := svc.RecomputeCluster(opp.ID)
	if err != nil {
		return fmt.Errorf("recompute cluster: %w", err)
	}
	if stats.Score < 1.0 {
		return fmt.Errorf("expected high opposition score, got %f", stats.Score)
	}
	fmt.Printf("  [smoke] cluster score=%.3f (between=%.3f, avg_within=%.3f)\n", stats.Score, stats.Between, stats.AvgWithin)

	// 裁决为确认。
	if err := svc.Adjudicate(opp.ID, "confirmed", "组间差异显著大于组内方差"); err != nil {
		return fmt.Errorf("adjudicate: %w", err)
	}

	// 批次流转：整理中→待复核→已发布。
	if err := svc.StartReview(bid); err != nil {
		return fmt.Errorf("start review: %w", err)
	}
	if err := svc.PublishBatch(bid); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	// 创建分析版本并冻结（不可变快照）。
	ver, err := svc.CreateVersion(bid, "首版声调证据")
	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}
	if err := svc.ShareVersion(ver.ID); err != nil {
		return fmt.Errorf("share version: %w", err)
	}
	if err := svc.FreezeVersion(ver.ID); err != nil {
		return fmt.Errorf("freeze version: %w", err)
	}

	// 封存批次（终态）。
	if err := svc.SealBatch(bid); err != nil {
		return fmt.Errorf("seal batch: %w", err)
	}
	fmt.Println("  [smoke] batch sealed, version frozen")

	if err := st.Close(); err != nil {
		return err
	}

	// ---- 第二次会话：重开同一数据库，验证持久化与重启恢复 ----
	st2, err := store.Open(dbFile)
	if err != nil {
		return fmt.Errorf("reopen db: %w", err)
	}
	defer st2.Close()
	svc2 := service.New(st2)

	b2, err := svc2.GetBatch(bid)
	if err != nil {
		return fmt.Errorf("reopen get batch: %w", err)
	}
	if b2.Status != "sealed" {
		return fmt.Errorf("after restart expected sealed, got %s", b2.Status)
	}
	opp2, err := svc2.GetOpposition(opp.ID)
	if err != nil {
		return fmt.Errorf("reopen get opposition: %w", err)
	}
	if opp2.Status != "confirmed" {
		return fmt.Errorf("after restart expected confirmed, got %s", opp2.Status)
	}
	evs, err := svc2.ListOppositions(bid, "")
	if err != nil {
		return err
	}
	if len(evs) != 1 {
		return fmt.Errorf("after restart expected 1 opposition, got %d", len(evs))
	}
	vers, err := svc2.ListVersions(bid)
	if err != nil {
		return err
	}
	if len(vers) != 1 || vers[0].Status != "frozen" {
		return fmt.Errorf("after restart version not frozen: %+v", vers)
	}
	segs, err := svc2.ListSegments(store.SegmentFilter{BatchID: bid})
	if err != nil {
		return err
	}
	if len(segs) != 4 {
		return fmt.Errorf("after restart expected 4 segments, got %d", len(segs))
	}
	fmt.Println("  [smoke] restart recovery verified (batch/opposition/version/segments persisted)")
	return nil
}
