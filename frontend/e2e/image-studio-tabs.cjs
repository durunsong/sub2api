/* eslint-disable @typescript-eslint/no-var-requires -- CommonJS resolves the host-provided Playwright bundle via NODE_PATH. */
// Native Web Locks + IndexedDB across real tabs; all HTTP responses are fixtures.
const assert = require('node:assert/strict')
const { chromium } = require('playwright')
const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
const user = { id: 900004, email: 'tabs@example.invalid', username: 'Tabs', role: 'user', status: 'active', balance: 100, concurrency: 5, allowed_groups: [7] }
const model = { id: '7:gpt-image-2', group_id: 7, group_name: 'Images', model: 'gpt-image-2', label: 'GPT Image 2', protocol: 'images', sizes: ['1024x1024'], qualities: ['auto'], max_images: 1, max_references: 3, pricing: { currency: 'USD', mode: 'image', image_price_1k: 0.03 } }
;(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const context = await browser.newContext()
    await context.addInitScript(user => {
      localStorage.setItem('auth_token','test-image-studio-tabs-token')
      localStorage.setItem('auth_user',JSON.stringify(user))
      localStorage.setItem('sub2api_locale','zh')
      localStorage.setItem('user_guide_900004_user_v4_interactive','true')
    },user)
    const pending = []
    let generation = 0, image
    await context.route('**/api/v1/**',async route => {
      const path = new URL(route.request().url()).pathname
      let data = []
      if (path.endsWith('/auth/me')) data = user
      else if (path.endsWith('/settings/public')) data = { site_name:'Sub2API', custom_menu_items:[], backend_mode_enabled:false }
      else if (path.endsWith('/image-studio/catalog')) data = {enabled:true,models:[model]}
      else if (path.endsWith('/image-studio/generate')) {
        generation++
        await new Promise(resolve => pending.push(resolve))
        if (route.request().postDataJSON().prompt === 'Failure deletion') { await route.fulfill({status:502,contentType:'application/json',body:JSON.stringify({code:502,message:'Cross-tab test failure'})}); return }
        data = {images:[{b64_json:image,mime_type:'image/png'}]}
      }
      try { await route.fulfill({contentType:'application/json',body:JSON.stringify({code:0,data})}) }
      catch (error) { if (!/closed|canceled|disposed|Target/.test(String(error))) throw error }
    })
    const first = await context.newPage()
    await first.goto(`${base}/image-studio`)
    await first.getByRole('main').getByRole('heading',{name:'生图工作台',exact:true}).waitFor()
    image = await first.evaluate(() => {const canvas=document.createElement('canvas');canvas.width=16;canvas.height=16;canvas.getContext('2d').fillRect(0,0,16,16);return canvas.toDataURL('image/png').split(',')[1]})
    await first.locator('#studio-prompt').fill('Shared task')
    await first.getByRole('button',{name:'开始生成',exact:true}).last().click()
    await first.getByText('生成中',{exact:true}).waitFor()
    const second = await context.newPage()
    await second.goto(`${base}/image-studio`)
    await second.getByText('另一个标签页仍在处理 1 个任务。请在原页面查看进度；完成的图片会自动同步。',{exact:true}).waitFor()
    assert.equal(await second.getByRole('button',{name:'重新生成',exact:true}).count(),0)
    const readTasks = async page => page.evaluate(async () => {const db=(await import('/src/features/image-studio/storage.ts')).createImageStudioDB(900004);try{return await db.listTasks()}finally{db.close()}})
    const liveTasks = await readTasks(second)
    assert.equal(liveTasks[0].status,'running')
    assert.equal(typeof liveTasks[0].ownerID,'string')
    assert.equal(generation,1)
    pending.shift()()
    await second.getByRole('button',{name:'局部重绘',exact:true}).waitFor()
    assert.equal(await second.getByText('另一个标签页仍在处理 1 个任务。请在原页面查看进度；完成的图片会自动同步。',{exact:true}).count(),0)
    await first.locator('#studio-prompt').fill('Interrupted task')
    await first.getByRole('button',{name:'开始生成',exact:true}).last().click()
    await first.getByText('生成中',{exact:true}).waitFor()
    await second.getByText('另一个标签页仍在处理 1 个任务。请在原页面查看进度；完成的图片会自动同步。',{exact:true}).waitFor()
    await first.close()
    pending.shift()()
    await second.getByRole('button',{name:'刷新',exact:true}).first().click()
    await second.getByText('已中断',{exact:true}).waitFor()
    const endedTasks = await readTasks(second)
    assert.equal(endedTasks[0].status,'interrupted')
    assert.equal(generation,2,'opening or refreshing another tab must not replay generation')
    await second.locator('#studio-prompt').fill('Failure deletion')
    await second.getByRole('button',{name:'开始生成',exact:true}).last().click()
    await second.getByText('生成中',{exact:true}).waitFor()
    const observer = await context.newPage()
    await observer.goto(`${base}/image-studio`)
    await observer.getByText('另一个标签页仍在处理 1 个任务。请在原页面查看进度；完成的图片会自动同步。',{exact:true}).waitFor()
    pending.shift()()
    await observer.getByText('Cross-tab test failure',{exact:true}).waitFor()
    await second.getByText('Cross-tab test failure',{exact:true}).waitFor()
    await observer.locator('.result-card').filter({hasText:'Cross-tab test failure'}).getByRole('button',{name:'删除',exact:true}).click()
    await observer.getByRole('dialog').getByRole('button',{name:'确认',exact:true}).click()
    await second.getByText('Cross-tab test failure',{exact:true}).waitFor({state:'detached'})
    await observer.close()
    // The no-Web-Locks fallback is explicit and never overwrites a running record.
    await second.evaluate(async () => {const db=(await import('/src/features/image-studio/storage.ts')).createImageStudioDB(900004);try{const task=(await db.listTasks())[0];await db.saveTask({...task,id:'unknown-owner',status:'running'})}finally{db.close()}})
    const fallback = await context.newPage()
    await fallback.addInitScript(() => Object.defineProperty(navigator,'locks',{value:undefined,configurable:true}))
    await fallback.goto(`${base}/image-studio`)
    await fallback.getByText('当前页面未追踪此任务，它可能仍在其他标签页或上游执行。重新生成会提交新的请求，并可能重复计费。',{exact:true}).waitFor()
    assert.equal((await readTasks(fallback)).find(task=>task.id==='unknown-owner').status,'running')
    assert.equal(generation,3)
    console.log(JSON.stringify({activeTabProtected:true,ownerPersisted:true,completionSynced:true,terminalDeletionSynced:true,closedTabInterrupted:true,noAutomaticReplay:true,noLocksFallbackExplicit:true,noLocksRecordPreserved:true},null,2))
  } finally { await browser.close() }
})().catch(error => {console.error(error);process.exitCode=1})
