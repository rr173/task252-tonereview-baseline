# 濒危语言声调证据复核台 (task252-tonoreview)

面向田野语言学家的语言声学证据复核后端服务。导入词条发音片段与基频轨迹，按说话人基线归一化后，比较同一音段上的候选声调轮廓，形成最小对立证据簇；研究者可排除噪声录音、裁决对立关系并发布不可变分析版本。

## 业务闭环

导入片段+基频 → 说话人基线归一化 → 最小对立证据簇 → 排除噪声/裁决 → 冻结版本。

## 状态机

- 田野批次：`collecting → reviewing → published → sealed`（封存为终态）
- 发音片段：`pending → usable | noise | excluded`（excluded 可恢复为 usable）
- 声调对立：`candidate → confirmed | rejected | insufficient`（confirmed/rejected 为终态）
- 分析版本：`draft → shared → frozen`（冻结可被新版本替代）

## 运行

```bash
go run ./cmd/tonereview --addr :8080 --db tonoreview.db
go run ./cmd/tonereview --smoke-test   # 自检：创建→导入→归一化→证据→裁决→冻结→封存→重启恢复
```

## 关键不变量

- 录音指纹（audio_fp）幂等，重复导入返回既有片段。
- 说话人基线由可用人声段可靠样本计算，样本不足则置空并令相关片段调型为 unknown。
- 封存批次拒绝一切修改（导入/裁决/版本冻结均守卫）。
- 分析版本冻结时写出不可变快照，旧冻结版本退为 superseded。

## 目录

仓库根目录就是源码根（本地 `env/` 原样推送，远程不再套一层 `env/`）：

```
cmd/tonereview/              入口与 --smoke-test
internal/model               实体、状态枚举、领域错误
internal/store               SQLite 持久化（modernc.org/sqlite）
internal/normalize           基线归一化、轮廓重采样、调型分类
internal/compare             轮廓距离、证据簇聚合、对立得分
internal/adjudicate          裁决规则与状态机守卫
internal/version             版本快照构建
internal/capture             片段导入校验与基频摄入
internal/service             编排层
internal/httpapi             /api JSON 接口
go.mod / go.sum
component-versions.json
Dockerfile / benzhi.Dockerfile / build_benzhi_docker.sh
BENZHI_README.md
```
