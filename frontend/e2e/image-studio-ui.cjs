/* eslint-disable @typescript-eslint/no-var-requires -- CommonJS resolves the host-provided Playwright bundle via NODE_PATH. */
// Browser integration against Vite, with deterministic HTTP fixtures and real IndexedDB.
const assert = require('node:assert/strict')
const fs = require('node:fs')
const { chromium } = require('playwright')
const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
const output = process.env.STUDIO_SCREENSHOT_DIR || '/tmp/sub2api-image-studio-ui'
const user = { id: 900003, email: 'studio@example.invalid', username: 'Studio', role: 'user', status: 'active', balance: 100, concurrency: 5, allowed_groups: [7] }
const model = { id: '7:gpt-image-2', group_id: 7, group_name: '图像创作', model: 'gpt-image-2', label: 'GPT Image 2', protocol: 'images', sizes: ['1024x1024','1536x1024','1024x1536'], qualities: ['auto','high','medium','low'], max_images: 1, max_references: 3, supports_mask: true, supports_custom_size: true, size_multiple: 16, pricing: { currency: 'USD', mode: 'image', rate_multiplier: 1, image_price_1k: 0.03, image_price_2k: 0.06, image_price_4k: 0.12 } }
;(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const context = await browser.newContext({ viewport: { width: 1512, height: 1100 }, locale: 'zh-CN' })
    await context.addInitScript(user => {
      localStorage.setItem('auth_token','test-image-studio-token')
      localStorage.setItem('auth_user',JSON.stringify(user))
      localStorage.setItem('sub2api_locale','zh')
      localStorage.setItem('user_guide_900003_user_v4_interactive','true')
      localStorage.setItem('admin_guide_900003_admin_v4_interactive','true')
    },user)
    let fixture, generation = 0
    const writes = [], requests = [], generationBodies = []
    await context.route('**/api/v1/**', async route => {
      const request = route.request(), url = new URL(request.url())
      requests.push(url.pathname)
      let data = [], status = 200
      if (request.method() !== 'GET') writes.push(url.pathname)
      if (url.pathname.endsWith('/auth/me')) data = user
      else if (url.pathname.endsWith('/settings/public')) data = { site_name: 'Sub2API', site_logo: '', backend_mode_enabled: false, custom_menu_items: [], available_channels_enabled: true }
      else if (url.pathname.endsWith('/image-studio/catalog')) data = { enabled: true, models: [model] }
      else if (url.pathname.endsWith('/image-studio/workflow-models')) data = { enabled: true, models: [] }
      else if (url.pathname.endsWith('/image-studio/generate')) {
        generation++
        const taskIndex = generation
        const body = request.postDataJSON()
        generationBodies.push(body)
        assert.equal(body.n, 1)
        assert.equal(body.model_id, model.id)
        assert.equal(body.aspect_ratio, undefined)
        assert.equal(body.resolution, undefined)
        await new Promise(resolve => setTimeout(resolve, taskIndex % 2 ? 250 : 80))
        if (taskIndex === 2) { status = 502; data = { code: 502, message: 'Test upstream rejected the image' } }
        else data = { images: [{ b64_json: fixture, mime_type: 'image/png' }] }
      } else if (url.pathname.endsWith('/subscriptions/summary')) data = { subscriptions: [], total: 0 }
      else if (url.pathname.includes('/version')) data = { version: 'test' }
      await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(status === 200 ? { code: 0, data } : data) })
    })
    const page = await context.newPage()
    const waitForLocal = async (check, timeout = 10000) => {
      const start = Date.now()
      while (Date.now() - start < timeout) {
        const records = await page.evaluate(async () => {
          const db = await new Promise((resolve, reject) => { const request = indexedDB.open('sub2api-image-studio-900003'); request.onsuccess = () => resolve(request.result); request.onerror = () => reject(request.error) })
          try { return await new Promise((resolve, reject) => { const request = db.transaction('history').objectStore('history').getAll(); request.onsuccess = () => resolve(request.result); request.onerror = () => reject(request.error) }) }
          finally { db.close() }
        })
        if (check(records)) return records
        await page.waitForTimeout(100)
      }
      throw new Error('Timed out waiting for local history')
    }
    const expandBatches = async () => {
      const collapsed = page.locator('[data-result-batch] .batch-fan[aria-expanded="false"]')
      while (await collapsed.count()) await collapsed.first().click()
    }
    const errors = []
    page.on('pageerror', error => errors.push(error.message))
    await page.goto(`${base}/image-studio`)
    await page.getByRole('main').getByRole('tab',{name:'图片创作',exact:true}).waitFor()
    fixture = await page.evaluate(() => { const canvas=document.createElement('canvas');canvas.width=600;canvas.height=400;const ctx=canvas.getContext('2d');ctx.fillStyle='#e8dfd0';ctx.fillRect(0,0,600,400);ctx.fillStyle='#4b5e45';ctx.beginPath();ctx.moveTo(0,350);ctx.lineTo(240,80);ctx.lineTo(400,300);ctx.lineTo(500,140);ctx.lineTo(600,400);ctx.fill();return canvas.toDataURL('image/png').split(',')[1] })
    // Exercise clipboard/drag payloads without touching the host clipboard.
    await page.locator('[data-reference-dropzone]').focus()
    const pasted = await page.evaluate(async encoded => {
      const blob = await (await fetch(`data:image/png;base64,${encoded}`)).blob()
      const data = new DataTransfer(); data.items.add(new File([blob], 'pasted.png', { type: blob.type }))
      const event = new ClipboardEvent('paste', { clipboardData: data, bubbles: true, cancelable: true })
      document.activeElement.dispatchEvent(event)
      return event.defaultPrevented
    }, fixture)
    assert.equal(pasted, true)
    await page.getByRole('button', { name: '移除参考图 1', exact: true }).waitFor()
    const textPreserved = await page.locator('#studio-prompt').evaluate(element => {
      const data = new DataTransfer(); data.setData('text/plain', 'ordinary prompt')
      const event = new ClipboardEvent('paste', { clipboardData: data, bubbles: true, cancelable: true })
      element.dispatchEvent(event)
      return !event.defaultPrevented
    })
    assert.equal(textPreserved, true)
    const transfer = await page.evaluateHandle(async encoded => {
      const blob = await (await fetch(`data:image/png;base64,${encoded}`)).blob()
      const data = new DataTransfer()
      data.items.add(new File([blob], 'dragged-1.png', { type: blob.type }))
      data.items.add(new File([blob], 'dragged-2.png', { type: blob.type }))
      return data
    }, fixture)
    const referenceZone = page.locator('[data-reference-dropzone]')
    await referenceZone.dispatchEvent('dragenter', { dataTransfer: transfer })
    assert.equal(await referenceZone.getAttribute('data-drag-active'), 'true')
    await page.getByText('松开鼠标，添加为参考图', { exact: true }).waitFor()
    fs.mkdirSync(output, { recursive: true })
    await referenceZone.screenshot({ path: `${output}/reference-drop.png` })
    await referenceZone.dispatchEvent('drop', { dataTransfer: transfer })
    await page.getByRole('button', { name: '移除参考图 3', exact: true }).waitFor()
    assert.equal(await referenceZone.getAttribute('data-drag-active'), 'false')
    await referenceZone.dispatchEvent('drop', { dataTransfer: transfer })
    await page.getByRole('alert').filter({ hasText: '参考图超出当前模型限制，请减少数量后重试。' }).waitFor()
    assert.equal(await referenceZone.getByRole('button', { name: /^移除参考图/ }).count(), 3)
    await transfer.dispose()
    while (await referenceZone.getByRole('button', { name: /^移除参考图/ }).count()) await referenceZone.getByRole('button', { name: /^移除参考图/ }).first().click()
    assert.equal(generation, 0)
    await page.locator('#studio-prompt').fill('山间晨雾，柔和自然光，宁静的横向构图')
    await page.locator('#studio-count').fill('3')
    await page.getByRole('button',{name:'开始生成',exact:true}).click()
    await page.locator('[data-result-batch]').waitFor()
    await expandBatches()
    await page.waitForFunction(() => document.querySelectorAll('.result-card').length >= 3)
    await page.getByText('Test upstream rejected the image',{exact:true}).first().waitFor()
    await page.waitForFunction(async () => { const db=await new Promise(resolve=>{const r=indexedDB.open('sub2api-image-studio-900003');r.onsuccess=()=>resolve(r.result)}); const count=await new Promise(resolve=>{const r=db.transaction('history').objectStore('history').count();r.onsuccess=()=>resolve(r.result)});db.close();return count===2 })
    assert.equal(generation,3)
    await page.waitForFunction(() => ![...document.querySelectorAll('.pending-image')].some(e => e.textContent.includes('生成中')))
    // Reuse a generated image as a local asset and then as a reference.
    await page.getByRole('button',{name:'保存为素材',exact:true}).first().click()
    await page.getByRole('button',{name:'我的素材',exact:true}).click()
    await page.getByRole('dialog').getByRole('button',{name:'用作参考图',exact:true}).first().waitFor()
    await page.getByRole('dialog').getByRole('button',{name:'用作参考图',exact:true}).first().click()
    await page.getByRole('button',{name:'移除参考图 1',exact:true}).waitFor()
    // A mask edit needs instructions and paint, then sends a separate mask and clean original.
    await page.getByRole('button',{name:'局部重绘',exact:true}).first().click()
    await page.getByRole('dialog').getByRole('button',{name:'AI 修改',exact:true}).click()
    await page.getByText('请涂抹需要修改的区域，并填写修改要求。',{exact:true}).waitFor()
    assert.equal(generation,3)
    await page.getByRole('dialog').getByRole('textbox').fill('将标记区域改为蓝色陶瓷')
    const drawing = page.getByRole('dialog').locator('canvas')
    const rect = await drawing.boundingBox()
    await page.mouse.move(rect.x + rect.width * .4, rect.y + rect.height * .4)
    await page.mouse.down()
    await page.mouse.move(rect.x + rect.width * .6, rect.y + rect.height * .6, { steps: 8 })
    await page.mouse.up()
    await page.getByRole('dialog').getByRole('button',{name:'AI 修改',exact:true}).click()
    await waitForLocal(records => records.length === 3)
    assert.equal(generationBodies[3].references.length,1)
    assert.ok(generationBodies[3].prompt.includes('透明遮罩'))
    assert.ok(generationBodies[3].references[0].startsWith('data:image/png;base64,'))
    assert.ok(generationBodies[3].mask.startsWith('data:image/png;base64,'))
    // Create and review a local variable template, then generate its two-image series.
    await page.getByRole('button',{name:'创作工作流',exact:true}).click()
    await page.getByRole('dialog').getByRole('button',{name:'新建模板',exact:true}).click()
    await page.getByRole('dialog').getByLabel('模板名称',{exact:true}).fill('本地产品系列')
    await page.getByRole('dialog').getByLabel('提示词模板',{exact:true}).fill('{{产品}}，自然光产品摄影')
    await page.getByRole('dialog').getByLabel('多图系列要求（可选）',{exact:true}).fill('正面视角\n---\n俯拍视角')
    await page.getByRole('dialog').getByRole('button',{name:'保存模板',exact:true}).click()
    await page.getByRole('dialog').getByRole('button',{name:'使用模板',exact:true}).click()
    assert.equal(await page.getByRole('dialog').getByRole('button',{name:'预览生成提示词',exact:true}).isDisabled(),true)
    await page.getByRole('dialog').getByLabel('产品',{exact:true}).fill('白色咖啡杯')
    await page.getByRole('dialog').getByRole('button',{name:'预览生成提示词',exact:true}).click()
    await page.getByRole('dialog').getByRole('button',{name:'生成全部 2 张',exact:true}).click()
    await waitForLocal(records => records.length === 5)
    assert.equal(generation,6)
    // Concurrent tasks can reach the fixture in either order; both prompts must arrive once.
    assert.deepEqual(generationBodies.slice(4,6).map(body => body.prompt).sort(), ['白色咖啡杯，自然光产品摄影\n\n正面视角','白色咖啡杯，自然光产品摄影\n\n俯拍视角'].sort())
    // Arbitrary allowed dimensions survive alignment, generation and refresh.
    await page.getByRole('button',{name:'尺寸',exact:true}).click()
    await page.getByRole('option',{name:'自定义尺寸',exact:true}).click()
    await page.getByLabel('宽度 W',{exact:true}).fill('1050')
    await page.getByLabel('高度 H',{exact:true}).click()
    assert.equal(await page.getByLabel('宽度 W',{exact:true}).inputValue(),'1056')
    await page.locator('#studio-prompt').fill('自定义尺寸测试')
    await page.locator('#studio-count').fill('1')
    await page.getByRole('button',{name:'开始生成',exact:true}).click()
    await waitForLocal(records => records.length === 6)
    assert.equal(generationBodies[6].size,'1056x1024')
    // Multi-category assignment is local and preserves both choices.
    await page.getByRole('button',{name:'分类展示',exact:true}).click()
    for (const name of ['产品','灵感']) {
      await page.getByRole('button',{name:'新建分类',exact:true}).click()
      await page.getByRole('dialog').getByLabel('分类名称',{exact:true}).fill(name)
      await page.getByRole('dialog').getByRole('button',{name:'保存',exact:true}).click()
    }
    await page.getByRole('button',{name:'全部展示',exact:true}).click()
    await page.locator('.result-card').filter({has:page.getByRole('button',{name:'局部重绘',exact:true})}).first().getByRole('button',{name:'分类',exact:true}).click()
    await page.getByRole('dialog').getByLabel('产品',{exact:true}).check()
    await page.getByRole('dialog').getByLabel('灵感',{exact:true}).check()
    await page.getByRole('dialog').getByRole('button',{name:'保存',exact:true}).click()
    await waitForLocal(records => records.some(item => item.categoryIDs?.length === 2))
    fs.mkdirSync(output,{recursive:true})
    await page.getByRole('dialog').waitFor({state:'hidden'})
    await page.screenshot({path:`${output}/desktop-light.png`,fullPage:true})
    await page.getByRole('button',{name:'底部',exact:true}).click()
    await page.waitForTimeout(500)
    await page.reload()
    await page.getByRole('main').getByRole('tab',{name:'图片创作',exact:true}).waitFor()
    await page.locator('.layout-bottom').waitFor()
    await page.locator('[data-result-batch]').first().waitFor()
    await expandBatches()
    assert.equal(await page.locator('.result-card').count(),7)
    assert.equal(await page.getByLabel('宽度 W',{exact:true}).inputValue(),'1056')
    assert.equal(generation,7,'reload must not replay generation')
    await page.evaluate(() => document.documentElement.classList.add('dark'))
    await page.screenshot({path:`${output}/desktop-dark-bottom.png`,fullPage:true})
    await page.setViewportSize({width:390,height:844})
    await page.waitForFunction(() => getComputedStyle(document.querySelector('main').parentElement).marginLeft === '0px')
    await page.screenshot({path:`${output}/mobile-dark.png`,fullPage:true})
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth),true,'no mobile horizontal overflow')
    // External local deletion must be reflected after refresh.
    await page.evaluate(async () => { const db=await new Promise(resolve=>{const r=indexedDB.open('sub2api-image-studio-900003');r.onsuccess=()=>resolve(r.result)});await new Promise(resolve=>{const tx=db.transaction('history','readwrite');tx.objectStore('history').clear();tx.oncomplete=resolve});db.close() })
    await page.getByRole('button',{name:'刷新',exact:true}).first().click()
    await page.waitForFunction(() => document.querySelectorAll('.result-card').length === 1)
    assert.equal(writes.every(path => path === '/api/v1/image-studio/generate'),true,'no server history or asset writes')
    // Finished failed tasks can be dismissed locally and stay deleted after reload.
    await page.locator('.result-card').getByRole('button',{name:'删除',exact:true}).click()
    await page.getByRole('dialog').getByRole('button',{name:'确认',exact:true}).click()
    await page.waitForFunction(() => document.querySelectorAll('.result-card').length === 0)
    await page.reload()
    await page.getByRole('main').getByRole('tab',{name:'图片创作',exact:true}).waitFor()
    assert.equal(await page.locator('.result-card').count(),0)
    // Admin settings use fixture-backed configuration only.
    const admin = { ...user, role: 'admin' }
    const adminContext = await browser.newContext({ viewport: { width: 1512, height: 1100 }, locale: 'zh-CN' })
    await adminContext.addInitScript(admin => {
      localStorage.setItem('auth_token','test-image-studio-admin-token')
      localStorage.setItem('auth_user',JSON.stringify(admin))
      localStorage.setItem('sub2api_locale','zh')
      localStorage.setItem('admin_guide_900003_admin_v4_interactive','true')
    },admin)
    let config = { enabled: true, models: [] }, configWrites = 0
    await adminContext.route('**/api/v1/**',async route => {
      const request = route.request(), path = new URL(request.url()).pathname
      let data = []
      if (path.endsWith('/auth/me')) data = admin
      else if (path.endsWith('/settings/public')) data = { site_name: 'Sub2API', backend_mode_enabled: false, custom_menu_items: [] }
      else if (path.endsWith('/admin/image-studio/config')) { if (request.method() === 'PUT') { config = request.postDataJSON(); configWrites++ } data = config }
      else if (path.endsWith('/admin/image-studio/catalog')) data = { enabled: true, models: [model] }
      else if (path.endsWith('/admin/image-studio/workflow-models')) data = { enabled: true, models: [{ id:'text', group_id:7, group_name:'图像创作', model:'gpt-4o', label:'GPT', supports_references:true }] }
      else if (path.endsWith('/admin/groups/all')) data = [{ id:7, name:'图像创作' }]
      else if (path.includes('/version')) data = { version:'test' }
      await route.fulfill({contentType:'application/json',body:JSON.stringify({code:0,data})})
    })
    const adminPage = await adminContext.newPage()
    adminPage.on('pageerror',error => errors.push(error.message))
    await adminPage.goto(`${base}/admin/image-studio`)
    await adminPage.getByRole('main').getByRole('heading',{name:'生图工作台配置',exact:true}).waitFor()
    await adminPage.getByLabel('展示 gpt-image-2',{exact:true}).uncheck()
    await adminPage.getByLabel('显示名称 gpt-image-2',{exact:true}).fill('专业产品图')
    await adminPage.getByLabel('展示 gpt-4o',{exact:true}).uncheck()
    await adminPage.getByLabel('显示名称 gpt-4o',{exact:true}).fill('创作助手')
    await adminPage.getByRole('button',{name:'添加模型',exact:true}).click()
    await adminPage.getByRole('dialog').getByRole('button',{name:'计费分组',exact:true}).click()
    await adminPage.getByRole('option',{name:'图像创作',exact:true}).click()
    await adminPage.getByRole('dialog').getByLabel('模型 ID',{exact:true}).fill('configured-image-alias')
    await adminPage.getByRole('dialog').getByLabel('显示名称',{exact:true}).fill('已配置图像别名')
    await adminPage.getByRole('dialog').getByRole('button',{name:'添加模型',exact:true}).click()
    await adminPage.getByLabel('开放在线生图',{exact:true}).uncheck()
    await adminPage.getByRole('button',{name:'保存',exact:true}).click()
    await adminPage.getByText('配置已保存，新请求将使用最新配置。',{exact:true}).waitFor()
    assert.equal(configWrites,1)
    assert.equal(config.enabled,false)
    assert.equal(config.models.find(item=>item.model==='gpt-image-2').enabled,false)
    assert.equal(config.models.find(item=>item.model==='gpt-image-2').label,'专业产品图')
    assert.equal(config.models.find(item=>item.model==='configured-image-alias').group_id,7)
    assert.equal(config.models.find(item=>item.model==='gpt-4o').enabled,false)
    assert.equal(config.models.find(item=>item.model==='gpt-4o').protocol,undefined)
    await adminPage.reload()
    await adminPage.getByLabel('显示名称 configured-image-alias',{exact:true}).waitFor()
    assert.equal(await adminPage.getByLabel('开放在线生图',{exact:true}).isChecked(),false)
    assert.equal(await adminPage.getByLabel('展示 gpt-image-2',{exact:true}).isChecked(),false)
    assert.equal(await adminPage.getByLabel('显示名称 gpt-image-2',{exact:true}).inputValue(),'专业产品图')
    assert.equal(await adminPage.getByLabel('展示 gpt-4o',{exact:true}).isChecked(),false)
    assert.equal(await adminPage.getByLabel('显示名称 gpt-4o',{exact:true}).inputValue(),'创作助手')
    await adminPage.screenshot({path:`${output}/admin-settings.png`,fullPage:true})
    assert.deepEqual(errors,[])
    console.log(JSON.stringify({ directImagePaste: pasted, plainTextPastePreserved: textPreserved, referenceDropAndLimit: true, independentTasks: generation, mixedOutcomes: true, dismissedFailure: true, adminSettings: true, localAssets: true, maskEditing: true, variableSeries: true, customDimensions: true, multiCategory: true, persistedLayout: true, reloadWithoutBilling: true, externalDeletion: true, mobileOverflow: false, noContentUploads: true, screenshots: output, pageErrors: errors },null,2))
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode=1 })
