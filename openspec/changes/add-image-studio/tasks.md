# Image Studio Implementation Plan

Goal: 完整交付用户授权的在线 AI 生图工作台。
Spec: specs/image-studio/spec.md；架构见 design.md。
Constraints: 使用 feat/image-studio；不安装依赖、不读秘密、不执行迁移；保留 Fork；新文本 UTF-8 无 BOM。用户现已授权完整补齐并验证后提交推送，只有主代理执行 Git 写操作。

- [x] 1. 检查工作区规则、Fork 差异，创建新分支；调查参考工作台和现有网关。
- [x] 2. 固定后端目录/生成与前端 types 契约；补齐参考功能清单和凭据/计费接线设计。
- [x] 3. 先写失败测试，再实现工作台配置、用户目录、同步生成桥接、授权与隐私；复用既有网关和计费；运行定向 Go 测试。
- [x] 4. 先写失败测试，再实现原生 IndexedDB 仓储及校验，包含 Blob 完整性、账号隔离、事务失败、删除、分类/素材和备份恢复；运行定向前端测试和真实浏览器存储检查。
- [x] 5. 实现 API 与工作台状态（并发、取消、快照、重试、刷新中断、存储失败）并测试；实现参考工作台页面及管理配置页面、中英文、路由和导航。
- [x] 6. 独立评审权限/账务/隐私/存储和完整工作台功能，修复发现并回归。
- [x] 7. 运行最终 Go build/vet/相关测试，前端完整测试/typecheck/build/lint；真实浏览器桌面/移动和深浅色验收。
- [x] 8. 更新 Fork 文档与规则摘要；UTF-8/BOM/乱码、最终 diff 检查；逐项需求证据审计后交付，保留未提交分支。

- [x] 9. 补齐配置操作指南、低高度悬浮生成入口和显影等待动效；覆盖桌面/手机、低动态及排队/完成反馈，更新验证证据。
- [x] 10. 统一上传、生成、历史参考与素材预览，补齐缩放/切图/原始尺寸/触摸/键盘/错误恢复，完成浏览器及自动化验证。

## Task 11: AI 工作流服务

- [x] 11. 补齐授权文本模型目录与 AI 工作流接口，先红后绿测试并完成专项审查。

修改范围：backend 内 image_studio 相关 handler/service、必要的网关隐私保护和测试、user/admin 路由。禁止改前端或提交/push，不派子代理。以现有同步网关/internal:images 身份/NoContentPersistence 为基础。

契约：GET `/api/v1/image-studio/workflow-models` 及管理员 GET `/api/v1/admin/image-studio/workflow-models`，返回 `{enabled:boolean,models:[{id,group_id,group_name,model,label,supports_references:boolean}]}`。用户目录仅含授权分组实际可用的文本模型，排除图片/视频/音频/embedding等；覆盖配置复用 ImageStudioConfig.Models 的隐藏/命名项，管理员可看到隐藏模型。不能因 AllowImageGeneration=false 排除有文本能力的授权分组。未知能力参考图必须保守声明并在提交时验证。

POST `/api/v1/image-studio/workflow` 接收 `{model_id,mode:'workflow'|'series',prompt,references?:string[],count?:number}`。prompt 长度1–16000字符，references<=3内联PNG/JPEG/WebP，每张<=10MiB；series count=1–20（必填），workflow 可省略count。拒绝未知字段、远程URL、不支持参考图、无权限或隐藏模型、禁用工作台。每次提交重验目录及用户权限。固定非流式同步文本路径，原网关按所选分组实际计费，无自动重放，不允许客户端传URL/密钥/原始messages或工具。

返回 `{draft:{title,prompt,category,tags:string[],seriesPrompts:string[]}}`（workflow）或 `{prompts:string[]}`（series）。后端构建严格JSON输出指令，可接受单个markdown JSON围栏但不能容忍任意结构/超限/空结果，title<=120、prompt及单个seriesPrompt<=10000、category<=120、tags<=20且每项<=80、seriesPrompts<=20；series返回恰好count项。工作流以现有{{variable}}变量语法表示，变量可编辑。生成内容不自动执行；错误使用稳定信息，不回显上游内容。参考图用于视觉理解，服务端不得保存任何输入或输出正文；延用安全判定和账务元数据。

必须调查文本网关的日志/内容持久化路径并补齐 NoContentPersistence，不可假定图片已有保护覆盖文本。测试包括真实repo fake授权目录、权限撤销/隐藏、计费身份与固定路由、引用验证、恶意/超限/非JSON输出、上游错误脱敏与取消。验收用定向Go tests、build/vet。报告记录真实DB/供应商未实测的边界。

## Task 12: AI 工作流交互

- [x] 12. 在现有本地工作流弹窗中增加AI创建、审核保存与AI系列草稿生成，覆盖取消/错误/账号切换/移动端，提供管理端文本模型展示控制。

自动发现并默认选择首个文本模型，允许用户选模型；无URL/Key输入。AI创建支持需求、当前参考图、模型、取消、错误恢复；输出先进入现有可编辑模板，确认保存只写IndexedDB。系列模式可以生成1–20张完整提示词并人工编辑/排序/删除/单张或全部生成；复用当前生图模型及参考图，不自动扣生图费用。必须提示文本生成按所选分组计费，取消仍可能已计费。默认保留手工模板模式。API `_retry:true`，关闭/切换账号中止并丢弃过期响应。不新增依赖。单测先红后绿及浏览器桌面/手机/深浅色验收。

## Task 13: 验证、审查与推送

- [x] 13. 独立审查全部功能、修复发现，运行前后端测试/lint/typecheck/build/vet和浏览器回归，更新配置流程/Fork/编码检查，完成全部推送前验收。

交付操作：按用户授权提交并 push `feat/image-studio`，核对本地与远端 SHA；执行结果由最终交付输出记录，不预写尚未发生的远端结果。
