# 在线生图工作台配置与使用流程

核对日期：2026-11-17。适用于本 Fork 的 `feat/image-studio` 实现。本文不表示已部署或已通过真实供应商验证。

工作台直接使用站内登录身份、现有上游账号、分组授权和网关计费。管理员配置上游凭据；普通用户打开 `/image-studio` 选模型即可，**不需要创建、粘贴或填写用户 API Key、接口地址**。

## 1. 上线前准备

先确认拟发布版本同时包含后端工作台接口、前端页面和导航。只有前端文件不会产生可用的生图服务。本文不执行发布、不修改生产设置、不读取真实凭据，也不代替既有发布审批。

准备以下已获授权的信息：

- 上游账号接入类型、真实支持的图片模型及协议；API Key 或 OAuth 授权由管理员在账号管理页面完成。
- 用户使用余额还是订阅、哪些用户可以访问、各尺寸/模型售价及倍率。这些是经营配置，本文不预填价格或替用户决定。
- 可调用图片模型的上游额度、可用网络/代理，以及网关允许访问的上游地址。
- 测试用户和一次最小生成的费用授权。真实生图、重新生成和账号图片测试都可能产生费用。

本工作台自身没有新增数据库迁移，也不依赖异步生图或对象存储。若从较旧版本升级，仍须按实际版本差异核对既有迁移；“工作台无迁移”不代表整站升级无需迁移。

## 2. 先选对平台与协议

平台按**实际接入协议和账号类型**选择，不能只看模型品牌。

| 用户要使用的模型 | 管理端账号/分组平台 | 工作台接口协议 | 本代码中的模型 ID 示例 | 配置要点 |
| --- | --- | --- | --- | --- |
| GPT 图片 | OpenAI | Images API | `gpt-image-2` | API Key 接入提供 Images API 的上游，或使用现有支持图片的 OpenAI OAuth/Codex 账号；检查尺寸档位与真实上游能力 |
| Grok 图片 | Grok | Images API | `grok-imagine-image`、`grok-imagine-image-quality` | 支持 OAuth 和 API Key 账号；OAuth 另查“媒体生成资格” |
| Gemini 图片，原生接口 | Gemini；现有 Antigravity 接入则用 Antigravity | Gemini generateContent | `gemini-3-pro-image`、`gemini-3-pro-image-preview` | Gemini API Key 使用原生 `generateContent`；Antigravity 使用对应授权/上游，不能把普通 OpenAI 兼容地址填入该路径 |
| Gemini 图片，OpenAI 兼容接口 | OpenAI | Images API | 上游实际支持的 `gemini-…-image…` 或已映射别名 | 必须确认供应商提供 `/images/generations`（编辑还需 `/images/edits`）；只有 Chat Completions 接口不足以证明可用 |

Nano Banana 不是需要另建的独立平台：按实际上游协议选择 Gemini / Antigravity 或 OpenAI，并把别名映射到该账号真正支持的图片模型；具体映射见 3.3。

示例 ID 来自当前代码，用来说明配置关系，不保证某供应商、套餐或账号实际支持。不要用任意新模型名称试图给上游增加能力。

### Base URL 的填写边界

- **OpenAI API Key / OpenAI 兼容 Images：**填供应商 API 根地址，例如 `https://images.example.com/v1`。代码会得到 `https://images.example.com/v1/images/generations` 或 `/images/edits`；根地址不带 `/v1` 时会补标准路径，已有版本前缀会保留。留空走官方 OpenAI 默认地址。优先填写供应商给出的根地址，不填写站内 `/api/v1/image-studio/generate`。
- **Gemini 原生 API Key：**留空走官方 Gemini API；自定义地址填写原生服务根地址，例如 `https://gemini.example.com`。代码在根地址后追加 `/v1beta/models/{model}:generateContent`，因此不要再手工追加 `/v1beta/models` 或完整模型动作路径。
- **Grok：**API Key 默认是 `https://api.x.ai/v1`；OAuth 使用其既有官方端点与授权流程。需要中转时使用账号表单中的上游地址/预设；自定义转发地址不会把 OAuth 授权与刷新改到该地址。不要为解决模型不可见而随意切换端点。
- **Antigravity：**OAuth 按页面授权步骤完成；选择“上游”类型时，填写提供 Antigravity 服务的上游地址与凭据。它与 Gemini API Key、OpenAI 兼容 Images 不是可互换的配置。

以上地址仅为语法示例。凭据仅在管理员账号表单中配置，不写入本文、截图、提示词或浏览器工作台草稿。

## 3. 管理员配置分组与上游账号

建议首次验收为每条协议链准备一个容易辨认的分组，先逐组跑通，再接入现有聚合路由。已有合适分组和账号可直接复用。

### 3.1 分组管理 `/admin/groups`

1. 创建或编辑目标分组，选择与上表一致的“平台”，保持分组为启用/活动状态。
2. 在“订阅设置 → 计费类型”中确认“标准（余额）”或“订阅（配额）”。**创建后不能修改计费类型**；不要为了生图随意修改已有订阅设计。
3. 在“图片生成计费”中勾选 **“允许当前分组生图”**（`allow_image_generation`）。工作台目录和真实请求均需要这一权限。
4. 若启用了分组模型白名单，让准备使用的完整请求模型 ID/别名通过白名单。别名需要授权的是用户请求的名字，不只是真实目标模型名。
5. 按第 4 节填写图片价格、逐模型定价和倍率，再保存。

“允许当前分组批量生图”及批量折扣是既有异步批量功能的配置。工作台的多图提交会拆成独立同步请求，不能据此假定工作台享受批量折扣，也不需要为工作台额外开启异步批量能力。

### 3.2 账号管理 `/admin/accounts`

1. 新增或编辑账号，选择平台及真实账号类型，填写名称。
2. API Key 类型填写管理员持有的**上游 Key**和正确 Base URL；OAuth 类型按该平台现有授权流程完成。Gemini 页面也有 OAuth、API Key、Service Account 等入口，必须使用真实可用的类型，不能仅因模型名称是 Gemini 就假设任一账号都支持生图。
3. 在账号表单底部的分组选择中关联目标分组。账号需处于可调度状态，且未过期、禁用、冷却或耗尽上游配额。
4. 配置“模型限制（可选）”中的“模型白名单”或“模型映射”，保留实际支持的模型。白名单为空表示不主动限制模型，不等于上游已支持所有模型。
5. 检查平台专有条件，保存后再进行目录与最小调用验收。

平台专有条件：

| 平台 | 必查字段/行为 |
| --- | --- |
| OpenAI | “生图尺寸档位”需包含预期档位；全选表示不限。当前“端点能力”主要控制文本/Embeddings，没有需要用户寻找的独立 Images 勾选。OAuth/Codex 图片路径与 API Key Images 转发不同，账号类型合格仍不保证上游实际出图。 |
| Grok OAuth | 编辑页“媒体生成资格”：通常保留“自动判断”，查看“当前判定”。“强制禁用”会排除图片/视频调度；“强制启用”只适用于已经确认支持媒体的账号，不会赋予供应商侧权限。Grok API Key 不使用该 OAuth 资格表单。 |
| Gemini / Antigravity | 图片模型应在该账号真实支持范围内，原生协议与账号授权/项目配置需一致。已有混合调度可以复用，但首次验收优先明确账号与分组对应关系。 |

OpenAI 的“生图结果 URL 转 base64”是普通 Images API 的兼容选项。工作台本身会要求可内联的结果，并走既有受限 URL 抓取转换，不需要为工作台配置对象存储。上游返回不可访问的 URL 仍可能导致“已生成但结果无法取回”。

### 3.3 模型映射与别名示例

账号“模型映射”的左侧为“请求模型”，右侧为“实际模型”。例如，在确实支持目标模型的账号上：

| 账号平台 | 请求模型（示例别名） | 实际模型（仅示例） | 工作台协议 |
| --- | --- | --- | --- |
| OpenAI | `studio-gpt` | `gpt-image-2` | Images API |
| Grok | `studio-grok` | `grok-imagine-image` | Images API |
| Gemini / Antigravity | `studio-gemini` | `gemini-3-pro-image` | Gemini generateContent |

映射不会新建上游模型。为了自动发现，自定义别名应是完整名字；通配符模式本身不会作为工作台选项展示。

同一分组下所有可能承接该别名的合格账号应映射到相同、明确的图片目标。若一个账号映射到文本模型、没有明确映射，或多个账号映射到不同目标，工作台不会自动把这个别名认定为图片模型。尤其避免把同一个别名在同一组内映射为不同供应商/图片规格。

渠道映射、聚合路由也可能参与模型解析；需要复用它们时，先核对既有真实路由和计费模型来源，再配置工作台展示。

## 4. 配置售价与计费口径

工作台没有独立钱包或价格表，沿用所选分组既有计费。页面估价已包含对应倍率，**不要再手工乘一次**。管理端预览不代表每个用户的专属售价，最终核对目标用户的工作台报价及用量账单。

### 4.1 要操作的字段

在 `/admin/groups` 目标分组中：

- **图片生成计费 → `1K ($)`、`2K ($)`、`4K ($)`：**分组通用基础图片单价，单位 USD。留空表示继承后续价源；填 `0` 是显式零价，不是“未配置”。
- **“生图倍率独立” / “生图独立倍率”：**控制按张费用是否使用独立倍率。
- **“按张计费”：**OpenAI / Composite 分组可见；用于专用图片端点，强制按实际出图张数计费，不把混合文本的 Responses/Chat/WS 一律改为按张。
- **“分组逐模型定价 → 添加模型价格”：**需要不同模型不同售价时使用。按现有价格卡选择计费模式、匹配模型和尺寸价格；不要只有空卡或不匹配的模型名。

已有渠道定价在 `/admin/channels` 管理。修改分组价前应检查是否已有匹配价格卡，避免以为填写了平面尺寸价就必然生效。

### 4.2 实际优先顺序

先判定计费模式，再看价格，最后应用倍率：

1. 通用模型价格解析先匹配**分组逐模型定价**，再看**渠道定价**，最后是模型价格目录/内置兜底。
2. 对图片输出，如果没有强制按张且匹配的分组/渠道价格卡为 **token 模式**，保持 token 结算；工作台显示“按实际 token 用量计费”。不能仅靠填写 `1K/2K/4K` 就断定覆盖了 token 模式。
3. 进入按张结算后，顺序是：**匹配的分组逐模型图片/按次价 → 当前尺寸的分组 `1K/2K/4K` 显式价 → 渠道图片/按次价 → 默认图片价**。强制按张时，不用 token 价卡计算零 token 费用，而是回落到图片价路径。
4. 默认图片价由代码中的 Grok 专用默认值或模型价格目录与内置兜底决定，不等于供应商实时成本。经营上需要固定售价时，应明确配置并验账。

按张倍率：

- 未开“生图倍率独立”：**用户对该分组的专属倍率 > 分组默认倍率 > 系统默认倍率**；这里是覆盖优先级，不是三者相乘。
- 开启后：直接使用“生图独立倍率”，替代上面的普通有效倍率；不再额外乘用户专属倍率。
- 当前按张图片倍率不叠加分组高峰因子；token 模式使用 token 倍率及适用的高峰规则。

可以用符号核算而不预设价格：某张图片命中的基础价为 `P`，普通有效倍率为 `R`，独立倍率为 `I`；未独立时约为 `P × R`，独立时约为 `P × I`。实际张数、模型映射、尺寸及请求时刻以结算记录为准。

尺寸档位按最长边划分：`≤1024 → 1K`、`≤2048 → 2K`、更大为 `4K`。例如 `1536×1024` 属于 2K，不因短边是 1024 就按 1K；`auto` 的实际结果可能改变计费档位。

别名的计费模型还受渠道“计费模型来源”、实际上游模型和无价兜底影响。不要仅根据工作台“显示名称”创建价卡；该名称只是标签，模型 ID 和账单中的实际计费模型才是核对依据。

## 5. 给用户授予访问权限

目录可见和能完成扣费是两层检查。管理员目录会看活动分组，普通用户只看到自己有权使用的分组；管理员看到某模型，不代表普通用户自动有权限。

| 分组类型 | 管理员操作 | 用户还需满足 |
| --- | --- | --- |
| 标准、公开分组 | 默认对用户开放；在 `/admin/users` 检查“用户分组配置” | 若开启“限制可访问的公开分组”，必须勾选该组；生成还需足够余额及未耗尽的相关配额 |
| 标准、专属分组 | 在 `/admin/users` 的“设置允许分组”/“用户分组配置”授予目标组 | 账号活动、足够余额、相关额度/并发限制未超限 |
| 订阅分组 | 在 `/admin/subscriptions` 按已确定方案分配有效订阅，或让用户按站点既有购买流程取得订阅 | 订阅未到期且日/周/订阅周期总额度未耗尽；仅加入标准允许分组不代替订阅 |

用户专属倍率也在“用户分组配置”设置。余额调整、订阅分配和额度变更涉及真实权益，应按已有运营授权执行，不为验证目录随意充值或发放订阅。

## 6. 配置工作台展示 `/admin/image-studio`

1. 打开“生图工作台配置”，点击“刷新”。确认目标分组的模型已自动出现。
2. 勾选“开放在线生图”。无历史覆盖配置时后端默认开放，但仍受所有分组、账号和用户权限约束。
3. 按“计费分组”筛选，检查“模型 ID”“显示名称”“接口协议”。同名模型可分别出现在不同分组，费用和权限按所选分组计算。
4. 用“展示”隐藏不想开放的模型；用“显示名称”提供易读标签。点击“保存”，看到“配置已保存，新请求将使用最新配置”。修改标签不修改模型路由或价格。
5. 自定义别名未自动识别时，点击“添加模型”，填写“计费分组”“模型 ID”“显示名称”“接口协议”，再保存。Images 选 `Images API`；原生 Gemini/Antigravity 选 `Gemini generateContent`。
6. 取消覆盖时点“恢复自动”并保存：这会移除该行覆盖，让自动发现规则重新生效，不保证不符合规则的模型仍会出现。

“添加模型”只添加展示/协议覆盖，**不能增加分组授权、绕过白名单、创建账号映射或赋予上游能力**。条目必须已经存在于该分组的实际模型目录中，且协议与分组平台一致；仅在管理页看到手动添加的行不证明用户端可生成。

隐藏只影响工作台，不改变公共模型 API 或其他客户端能力。关闭整个工作台会阻止新的工作台提交；不要把它当作终止已经发往上游的请求或退费开关。

## 7. 用户首次使用与验账

使用目标普通用户登录，打开侧栏“生图工作台”或 `/image-studio`：

1. 确认“模型”列表包含获授权分组中的目标模型；核对分组和显示的费用提示。
2. 首次只选 **1 张**，使用页面已有默认尺寸/质量，无参考图，以清晰简单提示词提交一次“开始生成”。先跑通最小链路，再测更高分辨率或编辑。
3. 等待图片显示，检查预览与下载。出现“尚未保存”时先下载，再用“重试本地保存”；这与重新生成不同，重存不应重新调用上游。
4. 打开用户 `/usage`，管理员同步检查 `/admin/usage`，按用户、时间、分组及模型找到该次调用。核对请求数量、计费模型、图片数量/尺寸或 token 用量、倍率、费用与余额/订阅用量变化。账务身份名称为 `Image Studio`；它不需要也不会在常规用户密钥列表暴露。
5. 记录实际结果与费用后，再测一次受控参考图编辑及需要的尺寸。只做账号连通性或文本测试不足以证明图片链路可用。

点击参考图、生成结果、历史记录上的参考图或素材封面即可预览。预览支持切换图片、缩略图、适应窗口、原始尺寸、缩放拖动与下载；手机支持双指缩放。键盘 ←/→ 切图、+/− 缩放、Shift+方向键移动、Esc 关闭。素材图片点击只预览，使用时点击“用作参考图”；预览和下载本地图片不会重新生成或计费。

参考图支持上传、拖入和 Ctrl/Cmd+V，PNG/JPEG/WebP，每张不超过 10MB；实际可用数量以当前模型限制为准。页面只展示代码认为支持的尺寸/质量，自定义尺寸也有校验；仍需供应商实际接受。

“生成张数”可选择 1–10，前端拆为每次一张的独立收费请求，最多 3 个并行任务，还受服务端用户/账号并发和额度限制。一个失败不会自动免费补图。“重新生成”会再次提交；客户端不会自动重放收费请求。刷新或取消不能保证已经提交到供应商的任务免费或可恢复，先查用量再决定是否重试。

### 本地内容保存

图片、参考图、历史、分类、收藏、素材、模板和草稿保存在当前站点、当前浏览器、当前用户隔离的 IndexedDB：`sub2api-image-studio-{userID}`。它们不随登录自动同步到另一台设备或另一个浏览器。

- 使用“导出备份”定期保存本地 JSON 备份；换设备/浏览器或清理网站数据前先导出。
- “导入备份”写入当前用户的本地存储；相同 ID 可能替换，先导出当前内容。备份含创作内容，应按作品资料保管。
- 浏览器清理网站数据、无痕模式退出、存储空间不足或 IndexedDB 被禁用都可能导致历史不可用。
- 服务端保留必要账务、安全判定/计数和配置元数据，但工作台创作内容不作为云端图库持久化；提示词和参考图仍须临时发送到实际上游完成生成，不能理解为不出浏览器。

### AI 创建工作流与系列提示词

管理员先为用户配置至少一个可用文本模型：在账号管理接入已有文本模型，把账号关联到用户可访问的分组，并在分组白名单中允许实际请求模型。文本分组不需要开启“允许当前分组生图”；它可以与图片分组独立。文本请求按照所选分组既有 token/模型定价、倍率和余额/订阅规则结算，不套用图片按张价。

已配置的文本别名会按实际目标模型识别能力，包括 `team/writer` 这样的公开名字和 Composite 显式路由别名；白名单、隐藏设置和实际请求仍使用公开名字。未显式映射的 DeepSeek、Kimi、智谱、MiniMax 使用项目已登记的对应平台文本候选，不会套用 Claude 默认名单；目录不是实时上游健康检查。

在 `/admin/image-studio`，文本候选显示为“文本 · 创作助手”。可以隐藏不希望展示的模型、修改名称或恢复自动。全局“开放在线生图”关闭时，创作助手也停止接收新请求。用户无需填写接口地址或API Key；没有授权文本模型时显示明确空态，仍可手工创建模板。

普通用户操作：

1. 按需在工作台上传、粘贴或拖入参考图，然后打开“创作工作流 → AI 创建工作流”。
2. 选择“创作助手模型”，描述要制作的工作流。已确认支持视觉输入的文本模型可以使用当前参考图（最多3张、每张10MB）；其他模型只发送文字需求，界面明确提示。
3. 点击“生成工作流草稿”。生成结果进入模板编辑页，可以修改名称、分类、标签、变量提示词和系列要求；点击“保存模板”才写入本机IndexedDB，不创建服务器模板记录，也不自动生图。
4. 对已保存模板点击“使用模板”，填写变量。可用“预览生成提示词”组装手工系列，或点击“AI 生成系列提示词”，选择1–20张及文本模型，编辑系列要求，再生成草稿。
5. 系列草稿是完整提示词，可以逐条编辑、排序、删除或追加要求。确认后点击“生成这张”或“生成全部”，才使用工作台当前图片模型和参数发起独立生图请求。

AI创建和AI系列各是一次文本计费请求，后续生图另行计费。取消、关闭弹窗、切换账号会终止客户端等待，迟到结果不会覆盖当前内容；已经产生的上游用量不保证退款。失败不会自动重试，手动再次生成可能再次计费。修改模板变量会使基于旧变量的在途系列请求失效。

## 8. 排障顺序

| 现象 | 按顺序检查 |
| --- | --- |
| 整个工作台显示未启用 | `/admin/image-studio` 的“开放在线生图”是否勾选且已保存；前后端是否为同一完整功能版本 |
| 用户“暂无可用模型” | 分组活动状态 → “允许当前分组生图” → 用户公开/专属/订阅权限 → 分组模型白名单 → 账号分组关联与模型映射 → 工作台是否隐藏及协议是否匹配 |
| 创作助手“暂无可用文本模型” | 用户文本分组权限 → 账号可调度状态与文本模型支持 → 分组模型白名单 → 工作台文本模型是否被隐藏。图片模型不会作为文本助手；未知能力或不明确的映射不会被当作已支持 |
| 管理员也看不到目标自动模型 | 分组平台是否在 OpenAI/Grok/Gemini/Antigravity/Composite 支持范围；请求模型是否进入分组目录；别名映射是否明确且一致；通配符不能直接做工作台选项。目录可能回退默认模型，所以“看得到”也不是健康检查 |
| 隐藏后用户页面仍留着旧选项 | 保存后让用户刷新目录/页面；服务器每次生成都会重新验目录，旧选项不能绕过隐藏。先前已提交请求另行查用量 |
| 保存配置提示旧分组不可用 | 已保存的停用/删除分组覆盖可以原样保留，不影响保存全局开关或其他模型；这些旧项可“恢复自动”移除，但不能新增或修改失效分组覆盖。需要编辑时先恢复分组可用状态 |
| 添加别名后仍不可用 | 管理页覆盖不创建上游模型；先检查账号“请求模型 → 实际模型”、分组白名单和实际目录，再确认协议；必要时“恢复自动”重新核对 |
| 401/403、未授权或余额不足 | 登录有效性、用户活动状态、分组授权、余额/在途预留、订阅有效性及额度；工作台错误消息做了脱敏，不能仅凭文案区分是哪一项 |
| 429、容量或配额耗尽 | 用户/分组限速与并发、订阅日/周/全期额度、平台额度、上游账号额度/冷却；有 `Retry-After` 时按提示等待，勿连续重复提交 |
| 有模型但没有可用账号 | 上游账号是否关联目标组、活动且可调度；OpenAI 尺寸档位；Grok OAuth 媒体资格；模型映射与上游实际支持情况。目录展示不是实时上游探测 |
| 400、供应商拒绝参数 | 先一张、默认尺寸/质量、无参考图；核对实际模型、协议、尺寸、质量和参考图限制；不要将 Images URL 配到 Gemini 原生账号或反向混用 |
| 5xx、响应无图片或图片无法取回 | 上游/代理连通性、Base URL 规则、真正的图片响应、上游结果 URL 可抓取性和响应大小；先看该次是否已计费，再决定重新生成 |
| 价格与预期不同 | 逐模型计费模式 → 强制按张开关 → 分组尺寸价 → 渠道价 → 别名/计费模型来源 → 用户专属或独立倍率 → 实际尺寸/张数/token。最终以用量账单为准 |
| 生成成功但历史丢失/未保存 | 是否同一站点/浏览器/用户，IndexedDB 和剩余存储空间，是否清理过网站数据；先下载，再重试本地保存或从备份恢复，不要直接再次生成 |

工作台故意不回显可能包含创作内容的完整上游错误，也不把创作内容写入审核、Ops、账号错误或普通请求日志。诊断优先用时间、用户、分组、状态、请求标识与用量元数据；不要为排障打开提示词/图片持久化。

## 9. 最小上线验收记录

以下由发布负责人在获授权环境执行并记录；本文编写时未调用真实上游，不能预填通过：

- [ ] 目标前后端版本均已发布，两个页面可访问；普通用户不能访问管理配置接口。
- [ ] 每条准备开放的协议链至少完成一张真实图片：GPT、Grok、Gemini 原生/Antigravity、Gemini Images 兼容按实际使用范围勾验。
- [ ] 普通用户只看见授权且展示开启的模型；隐藏/停用后的新提交被阻止。
- [ ] 余额用户/订阅用户按实际开放类型验账，价格、倍率、额度变动正确。
- [ ] 一张图片成功保存、下载、刷新后可见；备份导出/导入与账号隔离符合预期。
- [ ] 上游失败、取消/刷新、未保存情况下有明确提示，重试不被误认为免费。
- [ ] AI创建工作流、系列草稿和文本分组计费各做一次受控验收；保存模板不新增服务器内容记录，用户确认后才生图。

## 10. 源码依据

链接均相对本仓库；关键符号便于版本变动后重新定位。

| 说明 | 源码 |
| --- | --- |
| Fork 范围、无新增迁移、隐私边界 | [FORK_VS_UPSTREAM.md §8.8](FORK_VS_UPSTREAM.md#88-在线-ai-生图工作台fork-新增2026-11-17) |
| 页面路径 | [router/index.ts](../frontend/src/router/index.ts) |
| 工作台目录/协议/能力/覆盖/提交重验 | [handler/image_studio.go](../backend/internal/handler/image_studio.go)：`imageStudioProtocol`、`studioModel`、`catalog`、`Generate` |
| AI工作流目录、输出校验与文本转发 | [handler/image_studio_workflow.go](../backend/internal/handler/image_studio_workflow.go)、[service/image_studio_workflow.go](../backend/internal/service/image_studio_workflow.go) |
| AI模板与系列审核界面 | [WorkflowAssistant.vue](../frontend/src/features/image-studio/components/WorkflowAssistant.vue)、[WorkflowLibrary.vue](../frontend/src/features/image-studio/components/WorkflowLibrary.vue) |
| 内部账务身份、别名一致性、用户报价 | [service/image_studio.go](../backend/internal/service/image_studio.go)：`EnsureImageStudioKey`、`ImageStudioMappedModel`、`ImageStudioPricing` |
| 可用模型发现和默认回退 | [gateway_service.go](../backend/internal/service/gateway_service.go)：`GetAvailableModels` |
| 分组和用户授权 | [api_key_service.go](../backend/internal/service/api_key_service.go)：`GetAvailableGroups`、`canUserBindGroup`；[user.go](../backend/internal/service/user.go)：`CanBindGroup` |
| 管理页面字段 | [GroupsView.vue](../frontend/src/views/admin/GroupsView.vue)、[CreateAccountModal.vue](../frontend/src/components/account/CreateAccountModal.vue)、[EditAccountModal.vue](../frontend/src/components/account/EditAccountModal.vue)、[ImageStudioSettingsView.vue](../frontend/src/views/admin/ImageStudioSettingsView.vue) |
| 中文字段标签 | [admin/overview.ts](../frontend/src/i18n/locales/zh/admin/overview.ts)、[admin/accounts.ts](../frontend/src/i18n/locales/zh/admin/accounts.ts)、[studioAdmin.ts](../frontend/src/i18n/locales/zh/studioAdmin.ts)、[imageStudio.ts](../frontend/src/i18n/locales/zh/imageStudio.ts) |
| OpenAI URL 组装 | [openai_images.go](../backend/internal/service/openai_images.go)、[openai_images_test.go](../backend/internal/service/openai_images_test.go)：`TestBuildOpenAIImagesURL_HandlesVersionedBaseURL` |
| Gemini 原生 URL | [gemini_upstream_url.go](../backend/internal/service/gemini_upstream_url.go)：`buildGeminiAIStudioModelActionURL` |
| 账号类型、Grok 资格及端点 | [account.go](../backend/internal/service/account.go)：`SupportsOpenAIImageCapability`、`GetGrokBaseURL`、Grok media eligibility |
| 价格优先级与结算 | [model_pricing_resolver.go](../backend/internal/service/model_pricing_resolver.go)、[gateway_usage_billing.go](../backend/internal/service/gateway_usage_billing.go)：`calculateImageCost`、`recordUsageMultipliers`、`recordUsageBillingModel`；[billing_service.go](../backend/internal/service/billing_service.go)：`getImageUnitPrice` |
| 独立倍率、按张范围、高峰规则 | [image_billing_multiplier.go](../backend/internal/service/image_billing_multiplier.go)、[media_price_config.go](../backend/internal/service/media_price_config.go)、[group.go](../backend/internal/service/group.go)：`computePeakAwareMultipliers` |
| 工作台报价已有自动化用例 | [image_studio_pricing_test.go](../backend/internal/service/image_studio_pricing_test.go) |
| 同步提交、不重放、并发与本地保存 | [api.ts](../frontend/src/features/image-studio/api.ts)、[useImageStudio.ts](../frontend/src/features/image-studio/useImageStudio.ts)、[storage.ts](../frontend/src/features/image-studio/storage.ts) |

本文仅交付配置操作说明，未更改代码行为、后台设置或用户权益；未安装依赖、执行数据库迁移、访问生产数据或进行真实上游生成。
