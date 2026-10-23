# 本 fork 携带的本地补丁

本 fork 服务于 Float Desktop（`lrbmike/Float-Desktop`）的本地依赖开发模式：
桌面项目通过 `go mod replace` 引用本仓库检出，`main` 分支只做上游
（`oneo-me/onevra`）的镜像，**永不承载改动**；所有为桌面需求做的修改
都走 `float-local` 分支。

## 分支拓扑

```
upstream/main ──ff-only──> main（镜像）        float-local（桌面引用的稳定面）
                          └──cherry-pick──> xxx-主题分支（单个改动）
```

## 同步与冲突规则

1. 上游推进后：`main` 用 `git merge --ff-only upstream/main` 快进。
2. `float-local` 同步上游：`git merge main`（用 merge 不用 rebase，保留
   冲突解决的历史），冲突按语义裁决——
   - 纯上游演进相关 → 取**上游**；
   - 与 Float 需求相关 → 取**本 fork**。
3. 新增补丁：先从 `upstream/main` 切主题分支提交，再 cherry-pick 进
   `float-local`；优先"加文件/加可选参数"而不是改热点路径，把冲突面
   压到底。
4. 上游以其他写法吸收了某个主题后：从 `upstream/main` + 剩余主题重建
   `float-local`，作废对应主题分支。

## 补丁清单

| 主题分支 | 内容 | 为什么桌面需要 | 上游状态 |
|---|---|---|---|
| `sidebar-item-face` | 为 `SidebarItemOptions.Face`：`SidebarItem` 系列组件支持可选参数，覆盖选中项的默认 `Text.Alpha(.1)` 底色 | Float 的侧栏浮在有色背景上，选中项需要"凸起浮雕"而不是加深的 alpha 块 | 未提 PR；上游若以其它写法支持该能力，本主题作废并删除 |
| `primary-hover` | `ButtonPrimary` 悬停不再压暗强调色底，改为按钮下方一圈强调色淡投影（"浮起"）；按下才压暗一档 | Float 的强调色是偏深的橙：任何黑混加深都读成棕"脏色"，投影抬升是颜色不变的主流反馈 | 未提 PR；主要影响悬停观感，若上游改了悬停机制需按语义合并 |
