/* eslint-disable @typescript-eslint/no-var-requires */
// Local browser fixture: never calls a provider or production backend.
const assert=require('node:assert/strict')
const fs=require('node:fs')
const {chromium}=require('playwright')
const base=process.env.STUDIO_TEST_URL||'http://127.0.0.1:5179'
const output=process.env.STUDIO_SCREENSHOT_DIR||'/tmp/sub2api-studio-polish'
const video=fs.readFileSync(process.env.STUDIO_VIDEO_FIXTURE||'/tmp/sub2api-video-fixture.mp4')
const user={id:900021,email:'video@example.invalid',username:'Video',role:'admin',status:'active',balance:100,concurrency:5,allowed_groups:[7]}
const videoModel={id:'7:grok-imagine-video',group_id:7,group_name:'Video Group',model:'grok-imagine-video',label:'Grok Video',protocol:'grok'}
const imageModel={id:'7:gpt-image-2',group_id:7,group_name:'Images',model:'gpt-image-2',label:'GPT Image 2',protocol:'images',sizes:['1024x1024'],qualities:['auto'],max_images:1,max_references:3,supports_mask:true,pricing:{currency:'USD',mode:'image',rate_multiplier:1,image_price_1k:0.03,image_price_2k:0.06,image_price_4k:0.12}}
;(async()=>{
 const browser=await chromium.launch({headless:true,channel:'chrome'})
 try{
  const context=await browser.newContext({viewport:{width:1512,height:1000},locale:'zh-CN'})
  await context.addInitScript(user=>{
   if(!localStorage.getItem('theme'))localStorage.setItem('theme','light');localStorage.setItem('auth_token','video-fixture-token');localStorage.setItem('auth_user',JSON.stringify(user));localStorage.setItem('sub2api_locale','zh')
   localStorage.setItem('admin_guide_900021_admin_v4_interactive','true')
  },user)
  let generates=0,polls=0,completed=false,failed=false,catalogMode='ready',savedConfig={enabled:true,models:[]}
  const bodies=[],writes=[],errors=[]
  await context.route('**/api/v1/**',async route=>{
   const req=route.request(),url=new URL(req.url());let data=[]
   if(req.method()!=='GET')writes.push(url.pathname)
   if(url.pathname.endsWith('/auth/me'))data=user
   else if(url.pathname.endsWith('/settings/public'))data={site_name:'Sub2API',site_logo:'',backend_mode_enabled:false,custom_menu_items:[],available_channels_enabled:true}
   else if(url.pathname.endsWith('/image-studio/videos/catalog')){if(catalogMode==='offline'&&!url.pathname.includes('/admin/')){await route.abort();return}data={enabled:catalogMode!=='disabled',models:catalogMode==='empty'?[]:[videoModel]}}
   else if(url.pathname.endsWith('/image-studio/videos/config')){if(req.method()==='PUT')savedConfig=req.postDataJSON();data=savedConfig}
   else if(url.pathname.endsWith('/image-studio/videos/generate')){generates++;bodies.push(req.postDataJSON());data={id:`task-${generates}`,status:'pending'}}
   else if(url.pathname.endsWith('/image-studio/videos/status')){polls++;data={id:url.searchParams.get('task_id'),status:failed?'failed':completed?'completed':'pending',progress:completed?100:25}}
   else if(url.pathname.endsWith('/image-studio/videos/content')){await route.fulfill({status:200,contentType:'video/mp4',body:video});return}
   else if(url.pathname.endsWith('/image-studio/catalog'))data={enabled:true,models:[imageModel]}
   else if(url.pathname.endsWith('/image-studio/config'))data={enabled:true,models:[]}
   else if(url.pathname.endsWith('/image-studio/workflow-models'))data={enabled:true,models:[]}
   else if(url.pathname.endsWith('/admin/groups/all'))data=[{id:7,name:'Video Group',platform:'grok',status:'active'}]
   else if(url.pathname.includes('/subscriptions/summary'))data={subscriptions:[],total:0}
   else if(url.pathname.includes('/version'))data={version:'test'}
   await route.fulfill({status:200,contentType:'application/json',body:JSON.stringify({code:0,data})})
  })
  const page=await context.newPage();page.setDefaultTimeout(10000);page.on('pageerror',e=>{errors.push(e.message);console.error('PAGEERROR',e.message)})
  const sourceRequests=[]
  await context.route('https://raw.githubusercontent.com/yukkcat/image-prompts/main/dist/sources/*.json',async route=>{
    sourceRequests.push(route.request().headers())
    const name=new URL(route.request().url()).pathname.split('/').pop()
    const local=process.env.STUDIO_PROMPT_FIXTURE_DIR && `${process.env.STUDIO_PROMPT_FIXTURE_DIR}/${name}`
    if (local && !fs.existsSync(local)) { await route.abort(); return }
    const body=local&&fs.existsSync(local)?fs.readFileSync(local):JSON.stringify([{id:'one',title:'苹果风格海报',prompt:'A calm product poster',tags:['工作','海报']},{id:'two',title:'自然山水',prompt:'A quiet mountain lake',tags:['自然']}])
    await route.fulfill({contentType:'application/json',body})
  })
  fs.mkdirSync(output,{recursive:true});const shots=[]
  const shot=async(name,caption)=>{await page.waitForTimeout(350);await page.screenshot({path:`${output}/${name}.png`,fullPage:true});shots.push({name,caption});console.log('SCREENSHOT',name)}
  const click=async name=>page.getByRole('button',{name,exact:true}).click()
  const tab=async name=>page.getByRole('tab',{name,exact:true}).click()
  const popup=()=>page.locator('.video-settings-popover')
  await page.goto(`${base}/image-studio`);await page.locator('#studio-prompt').waitFor()
  assert.equal(await page.getByRole('heading',{name:'生图工作台',exact:true}).count(),0)
  const dimensions=await page.locator('.studio-toolbar').evaluate(el=>({height:el.getBoundingClientRect().height,width:el.getBoundingClientRect().width}))
  assert(dimensions.height<85,JSON.stringify(dimensions))
  await shot('01-image-toolbar','图片工作台 · 合并工具栏')
  await page.locator('#studio-prompt').fill('保留图片草稿')
  await click('提示词库');await page.locator('.prompt-card').first().waitFor();await page.waitForTimeout(1500)
  if(process.env.STUDIO_PROMPT_FIXTURE_DIR) await page.waitForFunction(()=>Array.from(document.querySelectorAll('.prompt-card img')).slice(0,5).every(img=>img.complete),{},{timeout:45000}).catch(()=>{})
  assert.equal(await page.getByRole('tab',{name:'精选提示词'}).getAttribute('aria-selected'),'true')
  assert.equal(await page.locator('.category-list button').count(),8)
  await shot('02-prompt-gallery','精选提示词 · 七个来源与缩略图')
  await click('Banana Prompt Quicker');await page.getByRole('textbox',{name:'按标题查询'}).fill('苹果风格')
  await shot('03-prompt-search','分类与关键词组合筛选')
  await page.getByRole('button',{name:'查看详情：苹果风格海报',exact:true}).click()
  const exactPrompt=await page.locator('.library-detail p.whitespace-pre-wrap').innerText()
  await shot('04-prompt-detail','详情 · 完整提示词与来源')
  await click('一键使用');assert.equal(await page.locator('#studio-prompt').inputValue(),exactPrompt);assert.equal(generates,0)
  await click('提示词库');await tab('我的自定义');await shot('05-image-custom','图片自定义 · 原有工作流')
  await page.keyboard.press('Escape');await tab('视频创作');await page.locator('[data-video-studio] textarea').waitFor()
  await shot('06-video-empty','视频空白会话')
  await click('生成参数');await popup().waitFor();await popup().getByRole('button',{name:'1080p',exact:true}).click();assert.equal(await popup().count(),1)
  await page.locator('[data-video-studio] textarea').click();assert.equal(await popup().count(),0)
  await click('生成参数');await shot('07-parameters','视频参数 · 视口内浮层')
  await tab('图片创作');assert.equal(await popup().count(),0);await tab('视频创作')
  await page.locator('[data-video-studio] textarea').fill('航拍山间晨雾，镜头缓缓掠过湖面，柔和的晨光照亮远山。')
  await click('生成视频');await page.locator('.video-pending[role="status"]').waitFor()
  await page.evaluate(()=>{window.__studioComposer=document.querySelector('.video-composer')})
  await shot('08-video-running','生成中 · 紧凑进度卡与独立输入区')
  await page.locator('.video-feed').evaluate(el=>{el.scrollTop=0});await page.waitForTimeout(10500)
  assert.equal(await page.locator('.video-feed').evaluate(el=>el.scrollTop),0,'polling must preserve scroll position')
  assert(await page.evaluate(()=>window.__studioComposer===document.querySelector('.video-composer')));assert.equal(generates,1)
  failed=true;await page.getByText('上游视频任务失败或未通过审核。',{exact:true}).waitFor({timeout:15000})
  await shot('09-video-failed','失败状态 · 清晰说明与操作')
  failed=false;await click('生成视频');await page.locator('.video-pending[role="status"]').waitFor()
  await shot('10-video-multiple','多条记录 · 统一卡片间距')
  await page.locator('.video-feed').evaluate(el=>{el.scrollTop=0});await page.waitForTimeout(5500)
  assert.equal(await page.locator('.video-feed').evaluate(el=>el.scrollTop),0)
  completed=true;await page.locator('.video-entry video').waitFor({timeout:15000})
  await shot('11-video-complete','完成结果 · 播放与下载')
  await click('提示词库');await page.locator('.prompt-card').first().waitFor();await tab('我的自定义')
  await page.getByRole('textbox',{name:'提示词名称',exact:true}).fill('本地镜头');await page.getByRole('dialog').getByRole('textbox',{name:'提示词',exact:true}).fill('一个缓慢推进的镜头')
  await click('保存');await shot('12-video-custom','视频自定义 · 本地保存')
  await page.keyboard.press('Escape');await click('深色模式');await shot('13-dark-video','深色主题 · 视频工作台')
  await click('提示词库');await shot('14-dark-gallery','深色主题 · 提示词库');await page.keyboard.press('Escape');await click('浅色模式')
  await page.setViewportSize({width:390,height:844});await shot('15-mobile-video','手机 · 视频结果与输入区')
  assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth))
  await click('提示词库');await shot('16-mobile-gallery','手机 · 筛选与双列图库')
  assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth))
  await tab('我的自定义');await shot('17-mobile-custom','手机 · 自定义提示词');await page.keyboard.press('Escape')
  await tab('图片创作');await shot('18-mobile-image','手机 · 紧凑图片工具栏')
  await page.setViewportSize({width:1512,height:1000})
  await context.route('https://raw.githubusercontent.com/yukkcat/image-prompts/main/dist/sources/*.json',route=>route.abort())
  await page.reload();await page.locator('#studio-prompt').waitFor();await click('提示词库');await page.locator('.prompt-card').first().waitFor()
  await page.getByRole('dialog').getByRole('button',{name:'刷新',exact:true}).click();await page.getByText('部分来源暂时无法更新，',{exact:false}).waitFor()
  await shot('19-offline-cache','断网 · 缓存可用与重试提示');await page.keyboard.press('Escape')
  await page.goto(`${base}/admin/image-studio?tab=video`);await page.getByText('启用视频工作台',{exact:true}).waitFor()
  await shot('20-admin','管理员 · 图片与视频配置 Tab')
  assert(sourceRequests.length>=7);assert(sourceRequests.every(h=>!h.authorization&&!h.cookie));assert.equal(errors.length,0,errors.join('\n'))
  fs.writeFileSync(`${output}/manifest.json`,JSON.stringify(shots,null,2))
  fs.writeFileSync(`${output}/index.html`,`<!doctype html><meta charset="utf-8"><title>工作台整改验收</title><style>body{font:16px system-ui;background:#f3f4f6;margin:32px;color:#111827}main{display:grid;grid-template-columns:repeat(auto-fit,minmax(420px,1fr));gap:24px}article{background:white;padding:16px;border-radius:12px}img{width:100%;height:auto}</style><h1>工作台整改验收 · ${shots.length} 张截图</h1><p>视频使用本地模拟 API，无真实付费调用。</p><main>${shots.map(s=>`<article><h2>${s.caption}</h2><a href="${s.name}.png"><img src="${s.name}.png" loading="lazy"></a></article>`).join('')}</main>`)
  console.log(JSON.stringify({screenshots:shots.length,generates,polls,sourceRequests:sourceRequests.length,errors,output}))
 }finally{await browser.close()}
})().catch(error=>{console.error(error);process.exit(1)})
