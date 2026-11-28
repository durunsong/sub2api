export default { videoStudio: {
        title: 'AI图片/视频工作台', adminTitle: 'AI图片/视频工作台', imageTab: '图片创作', videoTab: '视频创作', open: '打开视频工作台',
        greeting: '今天想要创作什么？', newSession: '新建对话', history: '创作记录', search: '搜索创作记录', noHistory: '还没有创作记录', recordCount: '{count} 条记录', deleteSession: '删除对话',
        preset_product: '生成一个 10 秒的产品展示短片，镜头缓慢推进', preset_landscape: '航拍山间晨雾，阳光穿过云层，电影质感', preset_story: '雨后城市街道，镜头跟随行人，霓虹倒映在水面',
        prompt: '提示词', promptPlaceholder: '描述镜头运动、主体动作、场景氛围和画面风格，输入 {\'@\'} 引用参考图…', selectModel: '选择视频模型', parameters: '生成参数', generate: '生成视频', generating: '正在生成…', pause: '暂停获取', paused: '已暂停获取结果，任务仍可能在上游继续运行。', resume: '继续获取结果', regenerate: '重新生成', reuse: '复用提示词',
        resolution: '清晰度', customResolution: '自定义清晰度', size: '尺寸', width: '宽度', height: '高度', ratio: '比例', auto: '自动', seconds: '视频秒数', secondsInput: '输入秒数', mode: '模式', frames: '首尾帧模式', reference: '全能参考模式', firstFrame: '首帧', lastFrame: '尾帧',
        providerParameters: '参数是否支持由所选渠道和模型决定。任务按现有分组价格计费，实际费用以用量记录为准。',
        references: '参考图', referenceNumber: '图片{n}', addReference: '添加参考图', removeReference: '移除参考图', previewReference: '查看参考图 {n}', referenceHint: '上传、拖入或粘贴图片 · 最多 {count} 张', referenceLimit: '参考图超过当前模式数量或总大小限制，请移除后再生成。', invalidReference: '请选择有效的 PNG、JPEG 或 WebP 图片，每张不超过 10 MB。',
        prompts: '提示词库', promptTitle: '提示词名称', noPrompts: '保存常用提示词，下次创作时直接使用。', useCurrentPrompt: '使用当前提示词', assets: '我的资产', uploadAsset: '上传素材', noAssets: '还没有本地素材', previewAsset: '预览素材：{name}', useReference: '加入参考', addAsset: '加入资产', assetSaved: '已加入本地资产', localOnly: '创作、视频和素材仅保存在当前浏览器，按用户隔离。清理浏览器数据会删除这些内容。',
        download: '下载', completed: '视频生成完成', disabled: '管理员已关闭视频创作，已有本地记录仍可查看。', noModels: '暂无可用的视频模型。请管理员检查 Grok/聚合分组授权与视频配置。',
        invalidResponse: '视频接口返回的数据无法识别，请重新获取任务状态。', invalidVideo: '未获取到有效的视频文件，请稍后继续获取。', generationFailed: '上游视频任务失败或未通过审核。', pollTimeout: '本轮查询已结束，任务尚未完成；可继续获取结果。', lookupFailed: '获取结果失败。任务已保留，可继续获取，无需重新生成。', submitUncertain: '提交未确认。请先检查用量记录；重新生成可能再次计费。', loadFailed: '无法加载视频工作台，请检查网络后刷新。',
        storageFailed: '无法读写本地数据，请检查浏览器存储权限和剩余空间。', storageBlocked: '本地数据库升级被其他标签页阻塞，请关闭旧标签页后重试。', storageQuota: '浏览器空间不足，请先下载并清理不需要的素材。', unsaved: '结果尚未保存到本地，请及时下载或重新保存。', unsavedTask: '任务已提交，但任务编号尚未保存；请重新保存，避免刷新后无法恢复。', saveAgain: '重新保存', imageLibraryFailed: '暂时无法读取图片素材库，视频素材仍可使用。',
        confirmRegenerate: '将按当前记录的模型和参数重新提交视频生成，并可能再次计费。是否继续？', confirmDelete: '删除此对话及全部本地生成记录？删除后无法恢复，也不会取消上游已提交任务或退还费用。', confirmDeleteItem: '删除这条本地记录？此操作无法撤销。',
        adminDescription: '管理视频工作台的可用模型和显示名称。授权、配额与计费沿用现有分组；不会修改图片工作台配置。', adminEnabled: '启用视频工作台', adminPricing: '视频沿用分组计费；Muse 必须先配置对应分辨率的每秒价格。模型须在渠道与分组中可用。', adminProtocol: 'Grok 视频协议', adminHint: '支持 Grok、OpenAI 兼容 Muse 和聚合分组。已配置别名仍须出现在分组实际模型列表中。', adminAdd: '添加视频模型别名',
        fullscreen: '全屏工作台', exitFullscreen: '退出全屏', retryLoad: '重新加载'
    } };
