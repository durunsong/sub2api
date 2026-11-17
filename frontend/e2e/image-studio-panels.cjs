/* eslint-disable @typescript-eslint/no-var-requires -- Uses the existing host Playwright runtime. */
const assert = require('node:assert/strict')
const fs = require('node:fs')
const { chromium } = require('playwright')
const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
const output = process.env.STUDIO_SCREENSHOT_DIR || '/tmp/sub2api-image-studio-panels'
const user = { id: 900009, email: 'panels@example.invalid', username: 'Panels', role: 'user', status: 'active', balance: 100, concurrency: 5, allowed_groups: [7] }
const model = { id: 'image', group_id: 7, group_name: '图像', model: 'gpt-image-2', label: 'Image', protocol: 'images', sizes: ['1024x1024'], qualities: ['auto'], max_images: 1, max_references: 3, pricing: { currency: 'USD', mode: 'image', rate_multiplier: 1, image_price_1k: 0.03 } }
;(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const context = await browser.newContext({ viewport: { width: 1280, height: 720 } })
    await context.addInitScript(user => {
      localStorage.setItem('auth_token', 'panels-fixture-token')
      localStorage.setItem('auth_user', JSON.stringify(user))
      localStorage.setItem('sub2api_locale', 'zh')
      localStorage.setItem('user_guide_900009_user_v4_interactive', 'true')
    }, user)
    const writes = [], errors = []
    await context.route('**/api/v1/**', async route => {
      const request = route.request(), path = new URL(request.url()).pathname
      if (request.method() !== 'GET') writes.push(path)
      const data = path.endsWith('/auth/me') ? user
        : path.endsWith('/settings/public') ? { site_name: 'Sub2API', custom_menu_items: [], backend_mode_enabled: false }
          : path.endsWith('/image-studio/catalog') ? { enabled: true, models: [model] }
            : path.endsWith('/image-studio/workflow-models') ? { enabled: true, models: [] } : []
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ code: 0, data }) })
    })
    const page = await context.newPage()
    page.on('pageerror', error => errors.push(error.message))
    await page.goto(`${base}/image-studio`)
    await page.getByRole('button', { name: '创作工作流', exact: true }).click()
    const dialog = page.getByRole('dialog')
    await dialog.getByRole('button', { name: '新建模板', exact: true }).click()
    await dialog.getByLabel('模板名称', { exact: true }).fill('长工作流滚动验收')
    const variables = Array.from({ length: 35 }, (_, i) => `{{变量${i}}}`).join(' ')
    await dialog.getByLabel('提示词模板', { exact: true }).fill(`${variables}\n${'自然光产品摄影，保留材质与构图细节。'.repeat(150)}`)
    await dialog.getByLabel('多图系列要求（可选）', { exact: true }).fill(Array.from({ length: 20 }, (_, i) => `构图${i}`).join('\n---\n'))
    await dialog.getByRole('button', { name: '保存模板', exact: true }).click()
    await dialog.getByRole('button', { name: '使用模板', exact: true }).click()
    const left = dialog.getByLabel('变量0', { exact: true }).locator('..').locator('..')
    const right = dialog.locator('p.whitespace-pre-wrap').locator('..')
    const back = dialog.getByRole('button', { name: '返回', exact: true })
    for (const height of [720, 480]) {
      await page.setViewportSize({ width: 1280, height })
      const headerBefore = await back.boundingBox()
      assert.ok(headerBefore && headerBefore.y >= 0 && headerBefore.y + headerBefore.height <= height, 'runner back button remains visible')
      for (const [name, panel] of [['variables', left], ['content', right]]) {
        const metrics = await panel.evaluate(el => ({ overflow: getComputedStyle(el).overflowY, client: el.clientHeight, scroll: el.scrollHeight }))
        assert.equal(metrics.overflow, 'auto', `${name} independently scrolls on desktop`)
        assert.ok(metrics.scroll > metrics.client, `${name} contains long scrollable content`)
        await panel.evaluate(el => { el.scrollTop = el.scrollHeight })
        assert.ok(await panel.evaluate(el => el.scrollTop > 0), `${name} actually scrolls`)
      }
      assert.equal((await back.boundingBox()).y, headerBefore.y, 'scrolling columns does not move the runner title/back')
      assert.equal(await dialog.locator('.modal-body').evaluate(el => el.scrollHeight <= el.clientHeight + 1), true, 'desktop has no outer nested scroll')
    }
    for (let i = 0; i < 35; i++) await dialog.getByLabel(`变量${i}`, { exact: true }).fill(`产品${i}`)
    await dialog.getByRole('button', { name: '预览生成提示词', exact: true }).click()
    assert.equal(await right.locator('textarea').count(), 20)
    assert.equal(await right.locator('div').evaluateAll(elements => elements.some(el => ['auto', 'scroll'].includes(getComputedStyle(el).overflowY))), false, 'series drafts share the right column scroll')
    fs.mkdirSync(output, { recursive: true })
    await dialog.screenshot({ path: `${output}/workflow-desktop-low.png`, animations: 'disabled' })
    await page.setViewportSize({ width: 390, height: 844 })
    assert.equal(await left.evaluate(el => getComputedStyle(el).overflowY), 'visible')
    assert.equal(await right.evaluate(el => getComputedStyle(el).overflowY), 'visible')
    assert.equal(await dialog.evaluate(el => el.scrollWidth <= el.clientWidth), true)
    await dialog.getByRole('button', { name: 'AI 生成系列提示词', exact: true }).scrollIntoViewIfNeeded()
    assert.equal(await dialog.getByRole('button', { name: 'AI 生成系列提示词', exact: true }).isVisible(), true)
    await dialog.getByRole('button', { name: '生成全部 20 张', exact: true }).scrollIntoViewIfNeeded()
    await dialog.screenshot({ path: `${output}/workflow-mobile.png`, animations: 'disabled' })
    assert.deepEqual(writes, [], 'scrolling and local template saves never submit billable requests')
    assert.deepEqual(errors, [])
    console.log(JSON.stringify({ desktopIndependentScroll: true, lowHeightHeaderReachable: true, mobileSingleScroll: true, pageErrors: errors, screenshots: output }, null, 2))
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode = 1 })
