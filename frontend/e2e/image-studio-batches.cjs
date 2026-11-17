/* eslint-disable @typescript-eslint/no-var-requires -- Uses the host-provided Playwright runtime. */
const assert = require('node:assert/strict')
const fs = require('node:fs')
const { execFileSync } = require('node:child_process')
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
    await page.locator('#studio-prompt').fill('四季山野的四张摄影作品，清晰的自然光')
    await page.locator('#studio-count').fill('4')
    await page.getByRole('button', { name: '开始生成', exact: true }).last().click()
    await page.waitForFunction(() => document.querySelectorAll('[data-result-batch]').length === 1)
    await page.waitForTimeout(150)
    assert.equal(requests, 3)
    const batch = page.locator('[data-result-batch]')
    assert.equal(await batch.locator('.result-card').count(), 0)
    // Both regular submit and the short-viewport dock share confirmation; cancel makes no call.
    await page.setViewportSize({ width: 1440, height: 620 })
    await page.locator('[data-generation-dock] [data-generate]').click()
    await page.getByRole('dialog').getByRole('heading', { name: '确认继续生成' }).waitFor()
    assert.equal(requests, 3)
    await page.getByRole('dialog').getByRole('button', { name: '取消', exact: true }).click()
    assert.equal(requests, 3)
    for (const release of held.splice(0)) release()
    await page.waitForFunction(() => document.querySelector('[data-result-batch]')?.textContent.includes('1 张处理中'))
    for (let i = 0; i < 40 && !held.length; i++) await page.waitForTimeout(50)
    assert.equal(held.length, 1); held.shift()()
    await page.locator('[data-studio-notice]').filter({ hasText: '生成完成 · 成功 3 张' }).waitFor()
    assert.equal(requests, 4)
    assert.ok((await page.locator('[data-studio-notice]').innerText()).includes('失败 1 张'))
    await page.waitForTimeout(5100)
    assert.equal(await page.locator('[data-studio-notice]').count(), 0)
    await batch.scrollIntoViewIfNeeded()
    await page.screenshot({ path: `${output}/fan-desktop.png`, animations: 'disabled' })
    await batch.locator('.batch-fan').hover()
    await page.waitForTimeout(450)
    await batch.screenshot({ path: `${output}/fan-hover.png` })
    await batch.getByRole('button', { name: '展开这一组', exact: true }).last().click()
    assert.equal(await batch.locator('.result-card').count(), 4)
    await batch.getByRole('button', { name: '收藏', exact: true }).first().click()
    const favorite = batch.getByRole('button', { name: '取消收藏', exact: true })
    await favorite.waitFor()
    assert.ok((await favorite.textContent()).includes('已收藏'))
    const contrast = await favorite.evaluate(el => ({ foreground: getComputedStyle(el).color, background: getComputedStyle(el).backgroundColor, fill: getComputedStyle(el.querySelector('svg')).fill }))
    assert.notEqual(contrast.foreground, contrast.background)
    assert.equal(contrast.fill, 'none')
    await page.screenshot({ path: `${output}/expanded-desktop.png`, animations: 'disabled' })
    // Download one original, then selected originals as one ZIP. No additional generation.
    await batch.getByRole('button', { name: '选择图片下载', exact: true }).click()
    let dialog = page.getByRole('dialog', { name: '选择图片下载', exact: true })
    assert.equal(await page.locator('[data-generation-dock]').count(), 0)
    await dialog.getByRole('button', { name: '全不选', exact: true }).click()
    await dialog.locator('[data-download-choice]').first().check()
    let downloading = page.waitForEvent('download')
    await dialog.locator('[data-download-selected]').click()
    const single = await downloading
    assert.ok(single.suggestedFilename().endsWith('.png'))
    await single.saveAs(`${output}/single.png`)
    await batch.getByRole('button', { name: '选择图片下载', exact: true }).click()
    dialog = page.getByRole('dialog', { name: '选择图片下载', exact: true })
    await dialog.locator('[data-download-choice]').last().uncheck()
    downloading = page.waitForEvent('download')
    await dialog.locator('[data-download-selected]').click()
    const multiple = await downloading
    await multiple.saveAs(`${output}/selected.zip`)
    execFileSync('python3', ['-c', 'import sys,zipfile; z=zipfile.ZipFile(sys.argv[1]); assert len(z.namelist())==2; assert z.testzip() is None; assert all(z.read(n).startswith(bytes([137,80,78,71,13,10,26,10])) for n in z.namelist())', `${output}/selected.zip`])
    assert.equal(requests, 4)
    await page.reload()
    await page.locator('[data-result-batch]').waitFor()
    assert.equal(await page.locator('[data-result-batch]').count(), 1)
    assert.equal(await page.locator('[data-studio-notice]').count(), 0)
    assert.equal(requests, 4)
    await page.setViewportSize({ width: 390, height: 844 })
    await page.evaluate(() => document.documentElement.classList.add('dark'))
    await page.emulateMedia({ reducedMotion: 'reduce' })
    await batch.scrollIntoViewIfNeeded()
    await page.screenshot({ path: `${output}/fan-mobile-dark.png`, animations: 'disabled' })
    const dark = await batch.locator('.batch-fan').evaluate(el => getComputedStyle(el).backgroundColor)
    assert.equal(dark, 'rgb(23, 30, 36)')
    assert.equal(await batch.locator('.fan-sheet').first().evaluate(el => getComputedStyle(el).transitionDuration), '0s')
    await batch.getByRole('button', { name: '展开这一组', exact: true }).last().click()
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true)
    await page.screenshot({ path: `${output}/expanded-mobile-dark.png`, animations: 'disabled' })
    // External deletion and backup round trip retain only real local images and their batch IDs.
    const local = await page.evaluate(async () => {
      const { createImageStudioDB } = await import('/src/features/image-studio/storage.ts')
      const db = createImageStudioDB(900011)
      const rows = await db.listHistory()
      const batchID = rows[0].batch.id
      if (!rows.every(row => row.batch.id === batchID)) throw new Error('batch lost')
      const backup = await db.exportBackup()
      await db.importBackup(backup)
      await db.deleteHistory(rows[0].id)
      const remaining = await db.listHistory(); db.close()
      window.dispatchEvent(new Event('focus'))
      return { count: remaining.length, batchID: remaining[0].batch.id }
    })
    assert.equal(local.count, 2)
    await page.waitForFunction(() => document.querySelectorAll('[data-batch-entries] .favorite-button').length === 2)
    assert.equal(requests, 4)
    assert.deepEqual(errors, [])
    console.log(JSON.stringify({ batchGrouping: true, confirmationCancelNoReplay: true, mixedCompletion: true, noticeFiveSeconds: true, favoriteContrast: contrast, singleDownload: true, selectedZipValidated: true, refreshNoReplay: true, mobileDarkReducedMotion: true, backupAndDeletion: true, requests, errors }, null, 2))
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode = 1 })
