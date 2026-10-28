# 本 fork 携带的本地补丁

本 fork 服务于本地桌面项目（Float Desktop、Calico Studio）的本地依赖开发
模式：桌面项目通过 `go mod replace` 引用本仓库检出，`main` 分支只做上游
（`oneo-me/onevra`）的镜像，**永不承载改动**；所有为桌面需求做的修改都走
`local` 分支。`local` 是所有项目共用的唯一稳定面，不按项目分叉。

## 分支拓扑

```
upstream/main ──ff-only──> main（镜像）        local（各项目 replace 的稳定面）
                          └──cherry-pick──> xxx-主题分支（单个改动）
```

**主 checkout 常驻 `local`**：`replace` 认目录不认分支，两个桌面项目读的就是
`~/workspaces/mine/onevra` 当前检出的内容。停在 `main` 会丢掉所有补丁（Float
直接编译失败，Calico 丢 `ConversationState.Width` 与 `inputHeight` 修复）；
停在某个主题分支则丢掉别的主题。

写主题分支时用独立 worktree，让主 checkout 一直待在 `local`：

```sh
git worktree add ../onevra-topic -b conversation-width upstream/main
# 在 ../onevra-topic 里改、提交
git checkout local && git cherry-pick conversation-width   # 主 checkout 只做这一步
git worktree remove ../onevra-topic
git push -u origin conversation-width
```

只读地看上游代码不必切分支：`git show upstream/main:<path>`、`git log -p`。

## 补丁准入规则

`replace` 指向的是一个工作树，同一时刻只能检出一个分支，所以**分支按「是否
改变他人行为」切，不按项目切**：

1. **可选参数优先**：加字段、加选项、加回调，零值等于旧行为。这类改动放
   `local`，其他项目无感。
2. **不改共享默认值**：某个项目嫌 720 窄就给自己的 `ConversationState`
   传 860，而不是把 `ContentWidth` 改成 860——那会同时改掉所有使用方。
3. **项目独有逻辑留在项目仓库**：界面编排、业务流程、文案在 app 里用现有
   组件组装，不进本库。
4. **确实无法参数化时才考虑分叉**：同一组件在两个项目里必须行为不同、且
   抽象表达不了时，先改抽象（插槽 / hook / 回调）；仍做不到才开项目专属
   分支，并且必须配独立 worktree（`git worktree add`，一份 checkout 只能
   待在一个分支）+ 在下方补丁清单里记明它与 `local` 的合并责任。

## 动手前与落地后的检查

主题分支从 `upstream/main` 切、`local` 才装全部补丁，两边因此可能错开版本。
下面三步把"在别人没跑的代码上写补丁"挡在提交之前：

1. **先对齐上游再切分支**（顺序不能反）：

   ```sh
   git checkout main && git merge --ff-only upstream/main
   git checkout local && git merge main
   git log --oneline local..upstream/main   # 必须为空，非空说明还没同步干净
   ```

   在落后于 `local` 的上游上写补丁，cherry-pick 只按文本判断冲突：上游改了同一
   段逻辑但行位置错开时，它会"干净通过"，语义却是错的。

2. **看目标文件已被哪些补丁动过**，避免与已有主题踩在同一段代码上：

   ```sh
   git diff upstream/main local -- source/components/chat/chat.go
   ```

   有重叠就先想清楚是扩展既有参数，还是新增一个（准入规则第 1 条）。

3. **cherry-pick 后每个桌面项目各自跑全量测试**——文本冲突 Git 会报，语义回归
   只有项目自己的测试能发现（Float Desktop 与 Calico Studio 都是
   `go test -count=1 ./internal/app/`，Calico 另加 `-race`）。

## 同步与冲突规则

1. 上游推进后：`main` 用 `git merge --ff-only upstream/main` 快进。
2. `local` 同步上游：`git merge main`（用 merge 不用 rebase，保留
   冲突解决的历史），冲突按语义裁决——
   - 纯上游演进相关 → 取**上游**；
   - 与桌面需求相关 → 取**本 fork**。
3. 新增补丁：先从 `upstream/main` 切主题分支提交，再 cherry-pick 进
   `local`；优先"加文件/加可选参数"而不是改热点路径，把冲突面压到底。
4. 上游以其他写法吸收了某个主题后：从 `upstream/main` + 剩余主题重建
   `local`，作废对应主题分支。

## 补丁清单

| 主题分支 | 内容 | 为什么桌面需要 | 上游状态 |
|---|---|---|---|
| `sidebar-item-face` | 为 `SidebarItemOptions.Face`：`SidebarItem` 系列组件支持可选参数，覆盖选中项的默认 `Text.Alpha(.1)` 底色 | Float 的侧栏浮在有色背景上，选中项需要"凸起浮雕"而不是加深的 alpha 块 | 未提 PR；上游若以其它写法支持该能力，本主题作废并删除 |
| `primary-hover` | `ButtonPrimary` 悬停不再压暗强调色底，改为按钮下方一圈强调色淡投影（"浮起"）；按下才压暗一档 | Float 的强调色是偏深的橙：任何黑混加深都读成棕"脏色"，投影抬升是颜色不变的主流反馈 | 未提 PR；主要影响悬停观感，若上游改了悬停机制需按语义合并 |
| `conversation-width` | `ConversationState.Width` 可选覆盖正文列宽（零值沿用 `ContentWidth`） | Calico Studio 的输入框比 720 宽，正文列需要跟着放宽才能对齐 | 未提 PR；上游若把列宽做成 `Conversation` 的参数，本主题作废并删除 |
