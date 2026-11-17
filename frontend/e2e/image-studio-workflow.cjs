/* eslint-disable @typescript-eslint/no-var-requires -- Uses the existing host Playwright runtime. */
const assert = require('node:assert/strict')
const fs = require('node:fs')
const { chromium } = require('playwright')
const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
const output = process.env.STUDIO_SCREENSHOT_DIR || '/tmp/sub2api-image-studio-workflow'
const user = { id: 900006, email: 'workflow@example.invalid', username: 'Workflow', role: 'user', status: 'active', balance: 100, concurrency: 5, allowed_groups: [7,8] }
const model = { id: 'image', group_id: 7, group_name: '生图分组', model: 'gpt-image-2', label: 'GPT Image 2', protocol: 'images', sizes: ['1024x1024'], qualities: ['auto'], max_images: 1, max_references: 3, pricing: { currency: 'USD', mode: 'image', image_price_1k: 0.03 } }
const textModel = { id: 'text', group_id: 8, group_name: '创作文本分组', model: 'gpt-4o', label: 'GPT 创作助手', supports_references: true }
;(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } })
    await context.addInitScript(user => {
      localStorage.setItem('auth_token', 'workflow-fixture-token')
      localStorage.setItem('auth_user', JSON.stringify(user))
      localStorage.setItem('sub2api_locale', 'zh')
      localStorage.setItem('user_guide_900006_user_v4_interactive', 'true')
    }, user)
    const writes = [], textRequests = [], imageRequests = [], errors = []
    let image, pending, noModels = false
    await context.route('**/api/v1/**', async route => {
      const request = route.request(), path = new URL(request.url()).pathname
      let data = []
      if (request.method() !== 'GET') writes.push(path)
      if (path.endsWith('/auth/me')) data = user
      else if (path.endsWith('/settings/public')) data = { site_name: 'Sub2API', custom_menu_items: [], backend_mode_enabled: false }
      else if (path.endsWith('/image-studio/catalog')) data = { enabled: true, models: [model] }
      else if (path.endsWith('/image-studio/workflow-models')) data = { enabled: true, models: noModels ? [] : [textModel] }
      else if (path.endsWith('/image-studio/workflow')) {
        const body = request.postDataJSON(); textRequests.push(body)
        assert.equal(body.model_id, textModel.id)
        if (body.prompt === '等待取消') await new Promise(resolve => { pending = resolve })
        if (body.prompt === '失败测试') { await route.fulfill({ status: 502, contentType: 'application/json', body: JSON.stringify({ code: 502, message: 'Fixture text provider unavailable' }) }); return }
        data = body.mode === 'workflow'
          ? { draft: { title: '自然光产品摄影', prompt: '{{产品}}，自然光，统一构图', category: '产品', tags: ['自然光'], seriesPrompts: ['产品正面', '材质细节'] } }
          : { prompts: ['白色咖啡杯，窗边正面全景', '白色咖啡杯，手柄材质特写'] }
      } else if (path.endsWith('/image-studio/generate')) {
        imageRequests.push(request.postDataJSON())
        data = { images: [{ b64_json: image, mime_type: 'image/png' }] }
      }
      try { await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ code: 0, data }) }) }
      catch (error) { if (!/closed|canceled|disposed|Target/.test(String(error))) throw error }
    })
    const page = await context.newPage()
    page.on('pageerror', error => errors.push(error.message))
    await page.goto(`${base}/image-studio`)
    await page.getByRole('main').getByRole('heading', { name: '生图工作台', exact: true }).waitFor()
    image = await page.evaluate(() => { const c = document.createElement('canvas'); c.width = 80; c.height = 80; const x = c.getContext('2d'); x.fillStyle = '#8fa59a'; x.fillRect(0,0,80,80); return c.toDataURL('image/png').split(',')[1] })
    await page.locator('input[type="file"][accept^="image"]').first().setInputFiles({ name: 'reference.png', mimeType: 'image/png', buffer: Buffer.from(image, 'base64') })
    await page.getByRole('button', { name: '移除参考图 1', exact: true }).waitFor()
    await page.getByRole('button', { name: '创作工作流', exact: true }).click()
    let dialog = page.getByRole('dialog')
    await dialog.getByRole('button', { name: 'AI 创建工作流', exact: true }).click()
    await dialog.getByLabel('描述你的工作流需求', { exact: true }).fill('为咖啡杯设计两张风格统一的宣传图工作流')
    await dialog.getByText('GPT 创作助手 · 创作文本分组', { exact: true }).waitFor()
    fs.mkdirSync(output, { recursive: true })
    await dialog.screenshot({ path: `${output}/create-desktop.png`, animations: 'disabled' })
    await dialog.getByRole('button', { name: '生成工作流草稿', exact: true }).click()
    await dialog.getByLabel('模板名称', { exact: true }).waitFor()
    assert.equal(imageRequests.length, 0)
    assert.equal(textRequests[0].references.length, 1)
    await dialog.getByLabel('模板名称', { exact: true }).fill('已审核的产品摄影')
    await dialog.getByRole('button', { name: '保存模板', exact: true }).click()
    await dialog.getByText('已审核的产品摄影', { exact: true }).waitFor()
    assert.equal(writes.length, 1, 'saving a template must stay local')
    await page.reload()
    await page.getByRole('button', { name: '创作工作流', exact: true }).click()
    dialog = page.getByRole('dialog')
    await dialog.getByText('已审核的产品摄影', { exact: true }).waitFor()
    await dialog.getByRole('button', { name: '使用模板', exact: true }).click()
    await dialog.getByLabel('产品', { exact: true }).fill('白色咖啡杯')
    await dialog.getByRole('button', { name: 'AI 生成系列提示词', exact: true }).click()
    await dialog.getByLabel('系列张数', { exact: true }).fill('2')
    await dialog.getByRole('button', { name: '生成系列草稿', exact: true }).click()
    await dialog.getByRole('button', { name: '生成全部 2 张', exact: true }).waitFor()
    assert.equal(imageRequests.length, 0, 'series results require explicit image submission')
    await dialog.locator('textarea').nth(1).fill('白色咖啡杯，审核后的细节图')
    await dialog.screenshot({ path: `${output}/series-desktop.png`, animations: 'disabled' })
    await page.setViewportSize({ width: 390, height: 844 })
    await page.evaluate(() => document.documentElement.classList.add('dark'))
    await dialog.screenshot({ path: `${output}/series-mobile-dark.png`, animations: 'disabled' })
    assert.equal(await dialog.evaluate(el => el.scrollWidth <= el.clientWidth), true)
    await dialog.getByRole('button', { name: '生成全部 2 张', exact: true }).click()
    await page.waitForFunction(() => document.querySelectorAll('.result-card img').length >= 2)
    assert.deepEqual(imageRequests.map(item => item.prompt).sort(), ['白色咖啡杯，窗边正面全景', '白色咖啡杯，审核后的细节图'].sort())
    await page.getByRole('button', { name: '创作工作流', exact: true }).click()
    await dialog.getByRole('button', { name: 'AI 创建工作流', exact: true }).click()
    await dialog.getByLabel('描述你的工作流需求', { exact: true }).fill('等待取消')
    await dialog.getByRole('button', { name: '生成工作流草稿', exact: true }).click()
    await dialog.getByRole('button', { name: '取消', exact: true }).waitFor()
    await dialog.screenshot({ path: `${output}/waiting-mobile-dark.png`, animations: 'disabled' })
    await dialog.getByRole('button', { name: '取消', exact: true }).click()
    assert.equal(typeof pending, 'function')
    pending()
    await dialog.getByText('已停止等待。上游可能已产生用量，再次生成会提交新的计费请求。', { exact: true }).waitFor()
    await dialog.getByLabel('描述你的工作流需求', { exact: true }).fill('失败测试')
    await dialog.getByRole('button', { name: '生成工作流草稿', exact: true }).click()
    await dialog.getByRole('alert').filter({ hasText: 'Fixture text provider unavailable' }).waitFor()
    assert.equal(textRequests.length, 4, 'errors and cancellation must not automatically retry')
    noModels = true
    await dialog.getByRole('button', { name: '返回', exact: true }).click()
    await dialog.getByRole('button', { name: 'AI 创建工作流', exact: true }).click()
    await dialog.getByText('暂无可用文本模型。请管理员为你的账号授权文本模型分组，并检查工作台模型展示设置。', { exact: true }).waitFor()
    assert.equal(await dialog.getByRole('button', { name: '生成工作流草稿', exact: true }).isDisabled(), true)
    assert.equal(writes.every(path => ['/api/v1/image-studio/workflow', '/api/v1/image-studio/generate'].includes(path)), true)
    assert.deepEqual(errors, [])
    console.log(JSON.stringify({ aiTemplateReviewed: true, templateLocalReload: true, referencesForwarded: true, aiSeriesReviewed: true, explicitImages: imageRequests.length, cancellation: true, manualRetryOnly: true, emptyModelState: true, mobileOverflow: false, pageErrors: errors, screenshots: output }, null, 2))
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode = 1 })
