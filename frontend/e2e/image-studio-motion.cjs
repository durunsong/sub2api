/* eslint-disable @typescript-eslint/no-var-requires -- Uses the host-provided Playwright runtime. */
// Real Vue / IndexedDB; held HTTP fixtures make waiting states deterministic, with no paid calls.
const assert = require('node:assert/strict')
const fs = require('node:fs')
const { chromium } = require('playwright')
const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
const output = process.env.STUDIO_MOTION_DIR || '/tmp/sub2api-image-studio-motion'
const user = { id: 900004, email: 'motion@example.invalid', username: 'Studio', role: 'user', status: 'active', balance: 100, concurrency: 5, allowed_groups: [7] }
const model = { id: '7:gpt-image-2', group_id: 7, group_name: '图像创作', model: 'gpt-image-2', label: 'GPT Image 2', protocol: 'images', sizes: ['1024x1024','1536x1024','1024x1536'], qualities: ['auto','high','medium','low'], max_images: 1, max_references: 3, supports_custom_size: true, size_multiple: 16, pricing: { currency: 'USD', mode: 'image', rate_multiplier: 1, image_price_1k: 0.03, image_price_2k: 0.06, image_price_4k: 0.12 } }
;(async () => {
  fs.mkdirSync(output, { recursive: true })
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const context = await browser.newContext({ viewport: { width: 1440, height: 640 }, locale: 'zh-CN', recordVideo: { dir: `${output}/video`, size: { width: 1440, height: 640 } } })
    const requests = [], errors = [], held = []
    let fixture
    const installFixtures = async context => {
    await context.addInitScript(user => {
      localStorage.setItem('auth_token','test-image-studio-token')
      localStorage.setItem('auth_user',JSON.stringify(user))
      localStorage.setItem('sub2api_locale','zh')
      localStorage.setItem(`user_guide_${user.id}_user_v4_interactive`,'true')
    }, user)
    await context.route('**/api/v1/**', async route => {
      const url = new URL(route.request().url())
      let data = []
      if (url.pathname.endsWith('/auth/me')) data = user
      else if (url.pathname.endsWith('/settings/public')) data = { site_name: 'Sub2API', site_logo: '', backend_mode_enabled: false, custom_menu_items: [] }
      else if (url.pathname.endsWith('/image-studio/catalog')) data = { enabled: true, models: [model] }
      else if (url.pathname.endsWith('/image-studio/generate')) {
        requests.push(route.request().postDataJSON())
        const outcome = await new Promise(resolve => held.push(resolve))
        if (outcome === 'failed') return route.fulfill({ status: 502, contentType: 'application/json', body: JSON.stringify({ code: 502, message: 'Controlled upstream failure' }) }).catch(() => {})
        data = { images: [{ b64_json: fixture, mime_type: 'image/png' }] }
      } else if (url.pathname.endsWith('/subscriptions/summary')) data = { subscriptions: [], total: 0 }
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, data }) }).catch(() => {})
    })
    }
    await installFixtures(context)
    const page = await context.newPage()
    page.on('pageerror', error => errors.push(error.message))
    await page.goto(`${base}/image-studio`)
    await page.getByRole('main').getByRole('heading', { name: '生图工作台', exact: true }).waitFor()
    const dock = page.locator('[data-generation-dock]')
    await dock.waitFor()
    assert.equal(await dock.locator('[data-generate]').isDisabled(), true)
    fixture = await page.evaluate(() => { const canvas = document.createElement('canvas'); canvas.width = 600; canvas.height = 400; const ctx = canvas.getContext('2d'); ctx.fillStyle = '#e8dfd0'; ctx.fillRect(0,0,600,400); ctx.fillStyle = '#4b5e45'; ctx.beginPath(); ctx.moveTo(0,400); ctx.lineTo(240,80); ctx.lineTo(400,300); ctx.lineTo(500,140); ctx.lineTo(600,400); ctx.fill(); return canvas.toDataURL('image/png').split(',')[1] })
    await page.locator('#studio-prompt').fill('晨雾里的山谷，胶片质感，柔和的自然光')
    await page.locator('#studio-count').fill('4')
    await page.evaluate(() => { document.querySelector('.studio-composer').scrollTop = 0 })
    await dock.waitFor()
    await page.screenshot({ path: `${output}/short-before.png` })
    await dock.locator('[data-generate]').click()
    await page.locator('[data-result-batch] .batch-fan[aria-expanded="false"]').click()
    await page.waitForFunction(() => document.querySelectorAll('.result-card .generation-pending[data-status="running"]').length === 3 && document.querySelectorAll('.result-card .generation-pending[data-status="queued"]').length === 1)
    assert.equal(requests.length, 3)
    assert.ok(requests.every(request => request.n === 1))
    await page.waitForTimeout(400) // Allow only the deliberate scroll transition to settle.
    const pending = page.locator('.result-card .generation-pending').first()
    await pending.scrollIntoViewIfNeeded()
    await page.screenshot({ path: `${output}/short-running.png` })
    const capture = async (name, reduced) => {
      await page.emulateMedia({ reducedMotion: reduced ? 'reduce' : 'no-preference' })
      fs.mkdirSync(`${output}/${name}`, { recursive: true })
      const samples = [], started = Date.now()
      for (const time of [0, 120, 900, 1800, 3000, 4800]) {
        await page.waitForTimeout(Math.max(0, started + time - Date.now()))
        samples.push(await pending.evaluate(el => ({ status: el.dataset.status, motion: el.dataset.motion, fragment: getComputedStyle(el.querySelector('.fragment-one')).transform, animation: getComputedStyle(el.querySelector('.fragment-one')).animationName, scan: getComputedStyle(el.querySelector('.scan-light')).transform })))
        await pending.screenshot({ path: `${output}/${name}/frame-${time}.png` })
      }
      fs.writeFileSync(`${output}/${name}/report.json`, JSON.stringify({ viewport: page.viewportSize(), reducedMotion: reduced, sampleTimesMs: [0,120,900,1800,3000,4800], samples, limitations: ['HTTP fixtures; no upstream generation or billing', 'Chrome simulated viewport; no physical mobile/keyboard or frame-time measurement'] }, null, 2))
      if (reduced) assert.ok(samples.every(sample => sample.animation === 'none'))
      else assert.ok(new Set(samples.map(sample => sample.fragment)).size > 1)
    }
    await capture('normal', false)
    await capture('reduced', true)
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    const sampleMotion = () => pending.evaluate(el => ({ motion: el.dataset.motion, hidden: document.hidden, animations: el.querySelector('.fragment-one').getAnimations().map(animation => ({ state: animation.playState, time: animation.currentTime })) }))
    const pauseSamples = { visible: await sampleMotion() }
    const queued = page.locator('.result-card .generation-pending[data-status="queued"]')
    await queued.scrollIntoViewIfNeeded()
    await queued.screenshot({ path: `${output}/queued-before.png` })
    const queuedAnimations = await queued.evaluate(el => el.getAnimations({ subtree: true }).length)
    await page.waitForTimeout(400)
    await queued.screenshot({ path: `${output}/queued-after.png` })
    assert.equal(queuedAnimations, 0)
    await page.evaluate(() => window.scrollTo(0,document.body.scrollHeight))
    await page.waitForFunction(() => document.querySelector('.result-card .generation-pending').dataset.motion === 'paused')
    assert.equal(await pending.locator('.fragment-one').evaluate(el => getComputedStyle(el).animationPlayState), 'paused')
    await page.waitForTimeout(100)
    pauseSamples.offscreen = await sampleMotion()
    await page.waitForTimeout(300)
    pauseSamples.offscreenLater = await sampleMotion()
    assert.equal(pauseSamples.offscreen.animations[0].time, pauseSamples.offscreenLater.animations[0].time)
    await pending.scrollIntoViewIfNeeded()
    await page.waitForTimeout(100)
    pauseSamples.resumed = await sampleMotion()
    assert.equal(pauseSamples.resumed.animations[0].state, 'running')
    // Headless Chrome keeps background pages visible; simulate the native visibility event only.
    await page.evaluate(() => { Object.defineProperty(document, 'hidden', { value: true, configurable: true }); document.dispatchEvent(new Event('visibilitychange')) })
    await page.waitForTimeout(100)
    pauseSamples.hiddenEvent = await sampleMotion()
    await page.waitForTimeout(300)
    pauseSamples.hiddenEventLater = await sampleMotion()
    assert.equal(pauseSamples.hiddenEvent.animations[0].time, pauseSamples.hiddenEventLater.animations[0].time)
    await page.evaluate(() => { delete document.hidden; document.dispatchEvent(new Event('visibilitychange')) })
    await page.waitForTimeout(100)
    pauseSamples.visibleEvent = await sampleMotion()
    assert.equal(pauseSamples.visibleEvent.animations[0].state, 'running')
    fs.writeFileSync(`${output}/pause-report.json`, JSON.stringify({ samples: pauseSamples, queuedAnimations, limitation: 'Page-hidden is an injected visibility event; headless Chrome did not expose real background-tab visibility.' }, null, 2))
    await page.locator('.studio-results').scrollIntoViewIfNeeded()
    // Dialogs and mobile navigation must take priority over the generation action.
    await page.getByRole('button', { name: '创作工作流', exact: true }).click()
    await page.getByRole('dialog').waitFor()
    assert.equal(await dock.count(), 0)
    await page.keyboard.press('Escape')
    await page.getByRole('dialog').waitFor({ state: 'hidden' })
    await page.setViewportSize({ width: 1440, height: 1100 })
    await page.evaluate(() => { window.scrollTo(0,0); document.querySelector('.studio-composer').scrollTop = 10000 })
    await page.locator('.composer-submit button[type="submit"]').waitFor({ state: 'visible' })
    assert.equal(await dock.count(), 0)
    await page.setViewportSize({ width: 1440, height: 640 })
    await page.evaluate(() => { document.querySelector('.studio-composer').scrollTop = 0 })
    await dock.waitFor()
    assert.equal(Math.round((await dock.boundingBox()).x), 256)
    await page.getByRole('button', { name: '收起', exact: true }).click()
    await page.waitForFunction(() => Math.round(document.querySelector('[data-generation-dock]').getBoundingClientRect().x) === 72)
    await page.getByRole('button', { name: '底部', exact: true }).click()
    await page.evaluate(() => window.scrollTo(0,0))
    await page.waitForTimeout(300)
    assert.equal(await page.getByRole('button', { name: /^继续提交/ }).count(), 1)
    await page.screenshot({ path: `${output}/bottom-layout.png` })
    await page.setViewportSize({ width: 390, height: 650 })
    await page.evaluate(() => { document.documentElement.classList.add('dark'); window.scrollTo(0,0) })
    await dock.waitFor()
    await page.waitForTimeout(400)
    assert.equal(await pending.evaluate(el => getComputedStyle(el).backgroundColor), 'rgb(2, 6, 23)')
    await page.screenshot({ path: `${output}/mobile-running.png` })
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth))
    await page.getByRole('button', { name: '切换菜单', exact: true }).click()
    assert.equal(await dock.count(), 0)
    await page.locator('aside').getByRole('link', { name: '生图工作台', exact: true }).click()
    await dock.waitFor()
    await dock.getByRole('button', { name: '编辑参数', exact: true }).click()
    assert.equal(await page.locator('#studio-prompt').evaluate(el => document.activeElement === el), true)
    await page.setViewportSize({ width: 844, height: 390 })
    await page.evaluate(() => window.scrollTo(0,0))
    await dock.waitFor()
    await page.waitForTimeout(400)
    const rect = await dock.boundingBox()
    assert.ok(rect.y >= 0 && rect.y + rect.height <= 390)
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth))
    await page.screenshot({ path: `${output}/mobile-landscape.png` })
    // A queued cancellation never reaches the upstream fixture; another task can fail independently.
    await page.locator('.result-card').filter({ has: page.locator('[data-status="queued"]') }).getByRole('button', { name: '取消', exact: true }).click()
    assert.equal(requests.length, 3)
    held[0]('failed')
    held[1]('success')
    held[2]('success')
    await page.getByText('Controlled upstream failure', { exact: true }).waitFor()
    await page.waitForFunction(() => document.querySelectorAll('.generation-pending').length === 0)
    assert.equal(requests.length, 3)
    // Isolated storage and a tall side layout keep the same result region unobstructed.
    for (const reduced of [false, true]) {
      const successContext = await browser.newContext({ viewport: { width: 1440, height: 1100 }, locale: 'zh-CN', reducedMotion: reduced ? 'reduce' : 'no-preference' })
      await installFixtures(successContext)
      const successPage = await successContext.newPage()
      successPage.on('pageerror', error => errors.push(error.message))
      await successPage.goto(`${base}/image-studio`)
      await successPage.getByRole('main').getByRole('heading', { name: '生图工作台', exact: true }).waitFor()
      await successPage.locator('#studio-prompt').fill('新作品显影验收')
      await successPage.getByRole('button', { name: '开始生成', exact: true }).click()
      await successPage.locator('.generation-pending').waitFor()
      const region = await successPage.locator('.studio-results').boundingBox()
      const clip = { x: region.x, y: region.y, width: region.width, height: 680 }
      const dir = `${output}/success-${reduced ? 'reduced' : 'normal'}`
      fs.mkdirSync(dir, { recursive: true })
      await successPage.screenshot({ path: `${dir}/before.png`, clip })
      held[held.length - 1]('success')
      await successPage.locator('[data-new-result="true"]').waitFor()
      const started = Date.now(), samples = []
      for (const time of [0, 75, 150, 300, 650, 900]) {
        await successPage.waitForTimeout(Math.max(0, started + time - Date.now()))
        samples.push(await successPage.locator('.result-card').evaluate(el => ({ fresh: el.dataset.newResult, opacity: getComputedStyle(el).opacity, transform: getComputedStyle(el).transform, animation: getComputedStyle(el).animationName })))
        await successPage.screenshot({ path: `${dir}/frame-${time}.png`, clip })
      }
      fs.writeFileSync(`${dir}/report.json`, JSON.stringify({ reducedMotion: reduced, clip, sampleTimesMs: [0,75,150,300,650,900], samples, limitations: ['HTTP fixture image, not an upstream generation', 'Screenshot sampling does not measure frame rate'] }, null, 2))
      assert.equal(samples[samples.length - 1].fresh, 'false')
      if (reduced) assert.ok(samples.every(sample => sample.transform === 'none'))
      await successPage.reload()
      await successPage.getByRole('main').getByRole('heading', { name: '生图工作台', exact: true }).waitFor()
      assert.equal(await successPage.locator('[data-new-result="true"]').count(), 0)
      await successContext.close()
    }
    assert.equal(requests.length, 5)
    assert.deepEqual(errors, [])
    const checks = { shortViewportDock: true, validation: true, tallViewportOriginal: true, bothLayouts: true, sidebarAlignment: true, dialogPriority: true, mobileNavigationPriority: true, mobilePortraitAndLandscape: true, editFocus: true, noOverflow: true, normalMotion: true, reducedMotion: true, offscreenPause: true, simulatedVisibilityPause: true, queuedStatic: true, queuedCancellation: true, independentFailureAndSuccess: true, localSuccessReveal: true, noReplay: true, requests: requests.length, pageErrors: errors, output }
    fs.writeFileSync(`${output}/checks.json`, JSON.stringify(checks, null, 2))
    console.log(JSON.stringify(checks, null, 2))
    await context.close()
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode = 1 })
