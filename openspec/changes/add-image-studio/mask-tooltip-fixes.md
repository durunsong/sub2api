# 局部编辑与图标提示修复

## 目标与验收

用户报告局部遮罩无效、纯图标含义不清。保留当前分支既有六项修复；不部署、安装依赖或调用真实收费模型。开发完成后用户已授权审查通过再提交并推送当前修复分支。

根因：MaskEditor把蓝色覆盖合进唯一参考图，useImageStudio未发送已有API支持的mask；擦除只是清除标记，界面缺少说明。纯图标只有aria-label，视觉用户没有气泡提示。

1. 支持mask的模型：原图保持原字节，另生成同尺寸PNG遮罩（涂抹区透明、其他区不透明），提交mask，不把蓝色画进原图。任务/历史/备份保存遮罩，明确重试仍使用同一区域；普通生成不串用旧遮罩。
2. 不支持mask但支持多参考的模型：明确标为“标记引导编辑”，发送原图+独立标记图和位置说明，不谎称精确遮罩，不发送不支持的mask。输入不足时明确阻止。
3. 画笔与擦除、重置、选区空检测正确；擦除仅清除标记，删除物体用画笔圈定并写删除要求。笔刷反馈与可见范围一致；重复提交、关闭/账号切换不触发迟到生成。沿用已有任务时二次确认。
4. 工作台纯图标：悬停与键盘聚焦显示中文/英文气泡，Escape可关闭，跟随滚动/边界，深浅色可读，卸载清理；复用aria-label不改变原动作。

## 实施

- [x] 原图/遮罩导出、模型能力与单次编辑参数传输，先红后绿测试。
- [x] 图标气泡及可访问性，先红后绿测试。
- [x] 浏览器真实canvas像素/请求/擦除/重试/存储与tooltip测试。
- [x] 独立审查、相关前后端回归、全量前端与构建、文档及编码检查。

## 2026-11-18 验证与审查

- `pnpm exec vitest run`：432文件、3505测试通过，退出0。
- `pnpm run lint:check`、`pnpm typecheck`、`pnpm run build`：退出0；build包含语言键、`vue-tsc -b`和Vite生产构建。仍有原Browserslist数据库过期与大chunk提醒，未更新依赖。
- `go test ./internal/handler -run '^TestImageStudio' -count=1`：通过；新增multipart测试验证原图和mask字节分别写入image[]/mask，n=1，非遮罩模型拒绝mask。没有修改后端生产逻辑。
- `node e2e/image-studio-mask.cjs`：先因原图被蓝色污染失败；修后真实canvas导出1024×768，painted alpha=0、erased/outside alpha=255，原图data URL逐字节一致。自动化验证完全擦光及reset零请求、48px笔刷范围、刷新不重放、手动重试带原遮罩、备份导入保留mask、普通生成无mask、非遮罩模型两张参考、390宽手机不溢出且底部按钮可达，4次fixture请求，pageErrors=[]。
- 擦光测试额外暴露抗锯齿残边：原alpha>0判定会错误提交。导出改为alpha>=128二值化，仅处理getImageData副本；原图Blob与画布标记不改。测试再次通过。
- `node e2e/image-studio-ui.cjs`、`image-studio-preview.cjs`、`image-studio-tooltips.cjs`：全部退出0；原素材、系列、分类、尺寸、局部编辑、预览缩放/触摸/嵌套/焦点行为保留。tooltip验证准确文案、hover/focus/Escape、移入气泡停留、禁用按钮悬停、深色窄屏、滚动和卸载，无页面错误。
- 独立审查：规格PASS、质量PASS；中文关闭气泡及hoverable两项P2均关闭。额外回归修复tooltip Escape吞掉预览关闭的问题；现在Escape同时执行既有弹窗行为。BaseDialog默认不开气泡，仅工作台opt-in。
- 最终Impeccable检测 `[]`；UTF-8、无BOM、无异常替换字符检查、`git diff --check`通过。桌面遮罩、手机编辑和深色气泡截图已人工查看。

原生遮罩契约参考：[OpenAI Images编辑文档](https://developers.openai.com/api/docs/guides/image-generation)，并以本仓库现有catalog的supports_mask与网关mask解析链为实现依据。模型能力不由品牌猜测，不保证实际供应商逐像素保留选区外画面。

限制：没有调用真实收费供应商、修改生产配置或部署；未验证物理手机Safari。所有浏览器HTTP供应商响应均为fixture，canvas/IndexedDB/Blob和请求结构使用实际应用实现。保留原先六项修复；交付分支为 `fix/image-studio-interactions`。

## 推送前复核

2026-11-18 用户明确授权审查无阻塞后提交、推送当前修复分支。重新运行全量前端：432文件3505测试、lint及build全部退出0；后端 `go test ./internal/handler -run '^TestImageStudio' -count=1` 退出0。构建保留已记录的既有非阻塞提醒。检查全部39个变更文件，无秘密/生成产物/新增个人绝对路径，UTF-8与diff检查通过；独立最终审查另行复核全部增量。
