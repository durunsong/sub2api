/* eslint-disable @typescript-eslint/no-var-requires -- Uses the host-provided Playwright runtime. */
const assert = require('node:assert/strict')
const fs = require('node:fs')
const { chromium } = require('playwright')
const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
const output = '/tmp/sub2api-image-studio-batches'
const user = { id: 900011, email: 'batch@example.invalid', username: 'Batch', role: 'user', status: 'active', balance: 100, concurrency: 5, allowed_groups: [7] }
const model = { id: '7:gpt-image-2', group_id: 7, group_name: '图像创作', model: 'gpt-image-2', label: 'GPT Image 2', protocol: 'images', sizes: ['1024x1024'], qualities: ['auto'], max_images: 1, max_references: 3, supports_mask: false, pricing: { currency: 'USD', mode: 'image', rate_multiplier: 1, image_price_1k: 0.03, image_price_2k: 0.06, image_price_4k: 0.12 } }
;(async () => {
  fs.mkdirSync(output, { recursive: true })
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const context = await browser.newContext({ viewport: { width: 1440, height: 900 }, locale: 'zh-CN' })
    await context.addInitScript(user => {
      localStorage.setItem('auth_token', 'fixture-only-token')
      localStorage.setItem('auth_user', JSON.stringify(user))
      localStorage.setItem('sub2api_locale', 'zh')
      localStorage.setItem('user_guide_900011_user_v4_interactive', 'true')
    }, user)
    let requests = 0, fixture = []
    const held = [], errors = []
    await context.route('**/api/v1/**', async route => {
      const path = new URL(route.request().url()).pathname
      let data = []
      if (path.endsWith('/auth/me')) data = user
      else if (path.endsWith('/settings/public')) data = { site_name: 'Sub2API', backend_mode_enabled: false, custom_menu_items: [] }
      else if (path.endsWith('/image-studio/catalog')) data = { enabled: true, models: [model] }
      else if (path.endsWith('/image-studio/workflow-models')) data = { enabled: true, models: [] }
      else if (path.endsWith('/image-studio/generate')) {
        const index = requests++
        assert.equal(route.request().postDataJSON().n, 1)
        await new Promise(resolve => held.push(resolve))
        if (index === 1) return route.fulfill({ status: 502, contentType: 'application/json', body: JSON.stringify({ code: 502, message: 'Fixture image failed' }) })
        data = { images: [{ b64_json: fixture[index % fixture.length], mime_type: 'image/png' }] }
      } else if (path.endsWith('/subscriptions/summary')) data = { subscriptions: [], total: 0 }
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, data }) })
    })
    const page = await context.newPage()
    page.on('pageerror', error => errors.push(error.message))
    await page.goto(`${base}/image-studio`)
    await page.locator('#studio-prompt').waitFor()
    fixture = await page.evaluate(() => ['#b5c9b5', '#b6c6df', '#ead1b3', '#d7c0c9'].map((color, index) => {
      const canvas = document.createElement('canvas'); canvas.width = 800; canvas.height = 600
      const ctx = canvas.getContext('2d'); ctx.fillStyle = color; ctx.fillRect(0, 0, 800, 600)
      ctx.fillStyle = '#263e34'; ctx.beginPath(); ctx.moveTo(0, 600); ctx.lineTo(200 + index * 50, 100); ctx.lineTo(800, 600); ctx.fill()
      ctx.fillStyle = '#fff6d6'; ctx.beginPath(); ctx.arc(620, 120, 45, 0, Math.PI * 2); ctx.fill()
      return canvas.toDataURL('image/png').split(',')[1]
    }))
    await page.locator('#studio-prompt').fill('tooltip fixture')
    await page.locator('#studio-count').fill('1')
    await page.getByRole('button', { name: '开始生成', exact: true }).last().click()
    for (let i = 0; i < 40 && !held.length; i++) await page.waitForTimeout(50)
    for (const release of held.splice(0)) release()
    const actions = page.locator('.result-card .icon-button')
    await actions.first().waitFor()
    const labels = ['保存为素材', '用作参考图', '下载', '删除']
    for (let i = 0; i < 4; i++) {
      const action = actions.nth(i)
      const label = await action.getAttribute('aria-label')
      await action.scrollIntoViewIfNeeded()
      await page.waitForTimeout(250)
      await action.hover()
      const bubble = page.getByRole('tooltip')
      await bubble.waitFor()
      assert.equal(await bubble.textContent(), label)
      assert.equal(await action.getAttribute('aria-describedby'), await bubble.getAttribute('id'))
      const hoverBox = await bubble.boundingBox()
      await page.mouse.move(hoverBox.x + hoverBox.width / 2, hoverBox.y + hoverBox.height / 2)
      await page.waitForTimeout(200)
      assert.ok(await bubble.isVisible(), JSON.stringify(await page.evaluate(() => ({ bubble: document.querySelector('[role=tooltip]')?.outerHTML, hovered: document.querySelector(':hover')?.outerHTML.slice(0,300) }))))
      await page.keyboard.press('Escape')
      assert.equal(await bubble.count(), 0)
      await page.mouse.move(0, 0)
      await action.focus()
      await bubble.waitFor()
      await page.keyboard.press('Escape')
      assert.equal(await bubble.count(), 0)
      await action.evaluate(el => el.blur())
    }
    assert.deepEqual(await actions.evaluateAll(els => els.map(el => el.getAttribute('aria-label'))), labels)
    await page.setViewportSize({ width: 390, height: 844 })
    await page.evaluate(() => document.documentElement.classList.add('dark'))
    await actions.last().scrollIntoViewIfNeeded()
    await actions.last().focus()
    const bubble = page.getByRole('tooltip')
    await bubble.waitFor()
    const box = await bubble.boundingBox()
    assert.ok(box.x >= 0 && box.y >= 0 && box.x + box.width <= 390)
    assert.equal(await bubble.evaluate(el => getComputedStyle(el).backgroundColor), 'rgb(31, 41, 55)')
    await page.screenshot({ path: '/tmp/sub2api-image-studio-tooltips.png' })
    await page.evaluate(() => window.dispatchEvent(new Event('scroll')))
    assert.ok(await bubble.isVisible())
    await page.getByRole('button', { name: '预览图片', exact: true }).last().click()
    await page.getByRole('button', { name: '关闭预览', exact: true }).hover()
    await page.getByRole('tooltip').waitFor()
    const closedTooltip = await page.getByRole('tooltip').getAttribute('id')
    await page.keyboard.press('Escape')
    await page.getByRole('button', { name: '关闭预览', exact: true }).waitFor({ state: 'hidden' })
    assert.equal(await page.locator(`[id="${closedTooltip}"]`).count(), 0)
    await page.keyboard.press('Escape')
    await actions.nth(1).evaluate(el => { el.disabled = true })
    await actions.nth(1).hover()
    await page.getByRole('tooltip').waitFor()
    assert.equal(await page.getByRole('tooltip').textContent(), labels[1])
    assert.deepEqual(errors, [])
    console.log(JSON.stringify({ hover: true, focus: true, escape: true, mobileDark: true, scroll: true, cleanup: true, errors }))
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode = 1 })
