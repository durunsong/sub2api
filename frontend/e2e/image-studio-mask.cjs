/* eslint-disable @typescript-eslint/no-var-requires -- Existing host Playwright runtime. */
const assert = require('node:assert/strict')
const fs = require('node:fs')
const { chromium } = require('playwright')
const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
const output = '/tmp/sub2api-image-studio-mask'
const user = { id: 900015, email: 'mask@example.invalid', username: 'Mask', role: 'user', status: 'active', balance: 100, concurrency: 5, allowed_groups: [7] }
const model = { id: 'mask-model', group_id: 7, group_name: 'Images', model: 'gpt-image-2', label: 'GPT Image 2', protocol: 'images', sizes: ['1024x1024'], qualities: [], max_images: 1, max_references: 3, supports_mask: true, pricing: { currency: 'USD', mode: 'image', image_price_1k: .03 } }
;(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } })
    await context.addInitScript(user => {
      localStorage.setItem('auth_token', 'mask-fixture-token'); localStorage.setItem('auth_user', JSON.stringify(user))
      localStorage.setItem('sub2api_locale', 'zh'); localStorage.setItem(`user_guide_${user.id}_user_v4_interactive`, 'true')
    }, user)
    let original, fail = true
    const requests = [], errors = []
    await context.route('**/api/v1/**', async route => {
      const path = new URL(route.request().url()).pathname
      let data = []
      if (path.endsWith('/auth/me')) data = user
      else if (path.endsWith('/settings/public')) data = { site_name: 'Sub2API', backend_mode_enabled: false, custom_menu_items: [] }
      else if (path.endsWith('/image-studio/catalog')) data = { enabled: true, models: [model] }
      else if (path.endsWith('/image-studio/generate')) {
        requests.push(route.request().postDataJSON())
        if (fail) { fail = false; return route.fulfill({ status: 502, contentType: 'application/json', body: JSON.stringify({ code: 502, message: 'Controlled edit failure' }) }) }
        data = { images: [{ b64_json: original.split(',')[1], mime_type: 'image/png' }] }
      } else if (route.request().method() !== 'GET') throw new Error(`Unexpected write: ${path}`)
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ code: 0, data }) })
    })
    const page = await context.newPage(); page.on('pageerror', e => errors.push(e.message))
    await page.goto(`${base}/image-studio`)
    await page.locator('#studio-prompt').waitFor()
    original = await page.evaluate(async () => {
      const canvas = document.createElement('canvas'); canvas.width = 1024; canvas.height = 768
      const ctx = canvas.getContext('2d'); ctx.fillStyle = '#b7c9aa'; ctx.fillRect(0, 0, 1024, 768); ctx.fillStyle = '#f4b66d'; ctx.fillRect(256, 192, 512, 384)
      const url = canvas.toDataURL('image/png'), blob = await (await fetch(url)).blob()
      const { createImageStudioDB } = await import('/src/features/image-studio/storage.ts')
      const db = createImageStudioDB(900015)
      await db.saveHistory({ id: 'source', createdAt: 1, status: 'success', snapshot: { groupID: 7, modelID: 'mask-model', model: 'gpt-image-2', prompt: 'Original image', size: '1024x1024', quality: 'auto', count: 1 }, images: [blob], references: [], favorite: false, categoryID: null })
      db.close(); return url
    })
    await page.reload(); await page.getByRole('button', { name: '局部重绘', exact: true }).waitFor()
    await page.locator('#studio-count').fill('3')
    await page.getByRole('button', { name: '局部重绘', exact: true }).click()
    let dialog = page.getByRole('dialog')
    await page.waitForFunction(() => document.querySelector('[role="dialog"] canvas')?.width === 1024)
    await dialog.getByRole('button', { name: 'AI 修改', exact: true }).click()
    assert.equal(requests.length, 0)
    await dialog.getByRole('textbox').fill('将选中区域改成红色，保留其他内容')
    const canvas = dialog.locator('canvas'), rect = await canvas.boundingBox()
    await dialog.getByRole('slider').fill('48')
    const draw = async fraction => { await page.mouse.move(rect.x + rect.width * fraction, rect.y + rect.height * .5); await page.mouse.down(); await page.mouse.up() }
    await draw(.4); await draw(.6)
    await dialog.getByRole('button', { name: '擦除', exact: true }).click(); await draw(.6)
    await draw(.4)
    await dialog.getByRole('button', { name: 'AI 修改', exact: true }).click()
    assert.equal(requests.length, 0, 'fully erased selection must not submit')
    await dialog.getByRole('button', { name: '画笔', exact: true }).click(); await draw(.4)
    await dialog.getByRole('button', { name: '重置标记', exact: true }).click()
    await dialog.getByRole('button', { name: 'AI 修改', exact: true }).click()
    assert.equal(requests.length, 0, 'reset selection must not submit')
    await draw(.4)
    assert.equal(await dialog.locator('[data-brush-cursor]').evaluate(el => el.getBoundingClientRect().width), 48)
    fs.mkdirSync(output, { recursive: true }); await page.screenshot({ path: `${output}/native-mask.png` })
    await dialog.getByRole('button', { name: 'AI 修改', exact: true }).click()
    await page.getByText('Controlled edit failure', { exact: true }).first().waitFor()
    assert.equal(requests.length, 1); assert.equal(requests[0].n, 1)
    assert.equal(requests[0].references[0], original, 'original bytes must not contain blue paint')
    assert.ok(requests[0].mask?.startsWith('data:image/png;base64,'), 'native edit must send a separate mask')
    const mask = await page.evaluate(async url => {
      const img = new Image(); img.src = url; await img.decode()
      const canvas = document.createElement('canvas'); canvas.width = img.width; canvas.height = img.height
      const ctx = canvas.getContext('2d'); ctx.drawImage(img, 0, 0)
      return { width: img.width, height: img.height, painted: ctx.getImageData(Math.floor(img.width * .4), img.height / 2, 1, 1).data[3], erased: ctx.getImageData(Math.floor(img.width * .6), img.height / 2, 1, 1).data[3], outside: ctx.getImageData(10, 10, 1, 1).data[3] }
    }, requests[0].mask)
    assert.deepEqual(mask, { width: 1024, height: 768, painted: 0, erased: 255, outside: 255 })
    await page.reload()
    await page.getByText('Controlled edit failure', { exact: true }).first().waitFor()
    assert.equal(requests.length, 1)
    await page.locator('.result-card').filter({ hasText: 'Controlled edit failure' }).getByRole('button', { name: '重新生成', exact: true }).click()
    await page.getByRole('dialog').getByRole('button', { name: '确认', exact: true }).click()
    await page.waitForFunction(() => document.querySelectorAll('.favorite-button').length === 2)
    assert.equal(requests.length, 2); assert.equal(requests[1].mask, requests[0].mask)
    const persisted = await page.evaluate(async () => {
      const { createImageStudioDB } = await import('/src/features/image-studio/storage.ts')
      const db = createImageStudioDB(900015), rows = await db.listHistory()
      const edit = rows.find(row => row.id !== 'source'), backup = await db.exportBackup()
      await db.importBackup(backup); const restored = (await db.listHistory()).find(row => row.id === edit.id)
      db.close(); return { mask: restored.snapshot.mask?.type, bytes: restored.snapshot.mask?.size }
    })
    assert.equal(persisted.mask, 'image/png'); assert.ok(persisted.bytes > 0)
    // A normal generation must not inherit the previous mask.
    await page.locator('#studio-count').fill('1'); await page.locator('#studio-prompt').fill('A fresh generation')
    await page.locator('form').getByRole('button', { name: '开始生成', exact: true }).click()
    await page.waitForFunction(() => document.querySelectorAll('.favorite-button').length === 3)
    assert.equal(requests[2].mask, undefined)
    // Models without native masks receive both the clean original and a separate guide.
    model.supports_mask = false; model.model = 'grok-imagine-image'; model.label = 'Grok'
    await page.reload(); await page.getByRole('button', { name: '局部重绘', exact: true }).first().click()
    dialog = page.getByRole('dialog'); await dialog.getByText('标记引导编辑', { exact: true }).waitFor()
    await page.waitForFunction(() => document.querySelector('[role="dialog"] canvas')?.width === 1024)
    await dialog.getByRole('textbox').fill('删除选中的物体')
    const r = await dialog.locator('canvas').boundingBox()
    await page.mouse.click(r.x + r.width / 2, r.y + r.height / 2)
    await dialog.getByRole('button', { name: 'AI 修改', exact: true }).click()
    await page.waitForFunction(() => document.querySelectorAll('.favorite-button').length === 4)
    assert.equal(requests[3].mask, undefined); assert.equal(requests[3].references.length, 2)
    assert.equal(requests[3].references[0], original); assert.notEqual(requests[3].references[1], original)
    assert.ok(requests[3].prompt.includes('第一张')); assert.ok(requests[3].prompt.includes('第二张'))
    await page.setViewportSize({ width: 390, height: 844 })
    await page.getByRole('button', { name: '局部重绘', exact: true }).first().click()
    dialog = page.getByRole('dialog')
    await dialog.getByRole('button', { name: 'AI 修改', exact: true }).scrollIntoViewIfNeeded()
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true)
    await page.screenshot({ path: `${output}/mobile-editor.png` })
    assert.deepEqual(errors, [])
    console.log(JSON.stringify({ nativeMaskPixels: mask, originalUnchanged: true, eraseWorks: true, singleEdit: true, retryAfterReload: true, maskBackup: persisted, ordinaryNoMask: true, guidedTwoReferences: true, requests: requests.length, errors }, null, 2))
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode = 1 })
