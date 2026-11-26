/* eslint-disable @typescript-eslint/no-var-requires */
// Local browser fixture: never calls a provider or production backend.
const assert=require('node:assert/strict')
const fs=require('node:fs')
const {chromium}=require('playwright')
const base=process.env.STUDIO_TEST_URL||'http://127.0.0.1:5179'
const output=process.env.STUDIO_SCREENSHOT_DIR||'/tmp/sub2api-video-studio-ui'
const video=fs.readFileSync(process.env.STUDIO_VIDEO_FIXTURE||'/tmp/sub2api-video-fixture.mp4')
const user={id:900021,email:'video@example.invalid',username:'Video',role:'admin',status:'active',balance:100,concurrency:5,allowed_groups:[7]}
const videoModel={id:'7:grok-imagine-video',group_id:7,group_name:'Video Group',model:'grok-imagine-video',label:'Grok Video',protocol:'grok'}
const imageModel={id:'7:gpt-image-2',group_id:7,group_name:'Images',model:'gpt-image-2',label:'GPT Image 2',protocol:'images',sizes:['1024x1024'],qualities:['auto'],max_images:1,max_references:3,supports_mask:true,pricing:{currency:'USD',mode:'image',rate_multiplier:1,image_price_1k:0.03,image_price_2k:0.06,image_price_4k:0.12}}
;(async()=>{
 const browser=await chromium.launch({headless:true,channel:'chrome'})
 try{
  const context=await browser.newContext({viewport:{width:1512,height:1000},locale:'zh-CN'})
  await context.addInitScript(user=>{
   localStorage.setItem('auth_token','video-fixture-token');localStorage.setItem('auth_user',JSON.stringify(user));localStorage.setItem('sub2api_locale','zh')
   localStorage.setItem('admin_guide_900021_admin_v4_interactive','true')
  },user)
  let generates=0,polls=0,completed=false,savedConfig={enabled:true,models:[]}
  const bodies=[],writes=[],errors=[]
  await context.route('**/api/v1/**',async route=>{
   const req=route.request(),url=new URL(req.url());let data=[]
   if(req.method()!=='GET')writes.push(url.pathname)
   if(url.pathname.endsWith('/auth/me'))data=user
   else if(url.pathname.endsWith('/settings/public'))data={site_name:'Sub2API',site_logo:'',backend_mode_enabled:false,custom_menu_items:[],available_channels_enabled:true}
   else if(url.pathname.endsWith('/image-studio/videos/catalog'))data={enabled:true,models:[videoModel]}
   else if(url.pathname.endsWith('/image-studio/videos/config')){if(req.method()==='PUT')savedConfig=req.postDataJSON();data=savedConfig}
   else if(url.pathname.endsWith('/image-studio/videos/generate')){generates++;bodies.push(req.postDataJSON());data={id:`task-${generates}`,status:'pending'}}
   else if(url.pathname.endsWith('/image-studio/videos/status')){polls++;data={id:url.searchParams.get('task_id'),status:completed?'completed':'pending',progress:completed?100:25}}
   else if(url.pathname.endsWith('/image-studio/videos/content')){await route.fulfill({status:200,contentType:'video/mp4',body:video});return}
   else if(url.pathname.endsWith('/image-studio/catalog'))data={enabled:true,models:[imageModel]}
   else if(url.pathname.endsWith('/image-studio/config'))data={enabled:true,models:[]}
   else if(url.pathname.endsWith('/image-studio/workflow-models'))data={enabled:true,models:[]}
   else if(url.pathname.endsWith('/admin/groups/all'))data=[{id:7,name:'Video Group',platform:'grok',status:'active'}]
   else if(url.pathname.includes('/subscriptions/summary'))data={subscriptions:[],total:0}
   else if(url.pathname.includes('/version'))data={version:'test'}
   await route.fulfill({status:200,contentType:'application/json',body:JSON.stringify({code:0,data})})
  })
  const page=await context.newPage();page.on('pageerror',e=>{errors.push(e.message);console.error('PAGEERROR',e.message)})
  const records=()=>page.evaluate(async()=>{const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('sub2api-video-studio-900021');r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error)});try{return await new Promise((resolve,reject)=>{const r=db.transaction('records').objectStore('records').getAll();r.onsuccess=()=>resolve(r.result.map(v=>({...v,video:v.video?{size:v.video.size,type:v.video.type}:null})));r.onerror=()=>reject(r.error)})}finally{db.close()}})
  await page.goto(`${base}/image-studio`)
  await page.locator('#studio-prompt').fill('keep image draft')
  await page.getByRole('tab',{name:'视频创作',exact:true}).click()
  await page.locator('[data-video-studio] textarea').waitFor().catch(async e=>{console.error(await page.locator('body').innerText());throw e})
  await page.locator('[data-video-studio] textarea').fill('A camera glides across a mountain lake')
  const png=Buffer.from(await page.evaluate(()=>{const canvas=document.createElement('canvas');canvas.width=16;canvas.height=16;canvas.getContext('2d').fillRect(0,0,16,16);return canvas.toDataURL('image/png').split(',')[1]}),'base64')
  await page.locator('[data-video-studio] input[type=file]').first().setInputFiles([{name:'first.png',mimeType:'image/png',buffer:png},{name:'last.png',mimeType:'image/png',buffer:png}])
  await page.waitForFunction(()=>document.querySelectorAll('.video-reference').length===2)
  await page.getByRole('button',{name:'全屏工作台',exact:true}).click()
  await page.locator('[data-video-studio] textarea').focus();await page.keyboard.press('Escape')
  await page.getByRole('button',{name:'全屏工作台',exact:true}).waitFor()
  assert.equal(await page.getByRole('button',{name:'生成视频',exact:true}).isEnabled(),true)
  await page.getByRole('button',{name:'生成参数',exact:true}).click()
  await page.getByRole('button',{name:'16:9',exact:true}).click()
  await page.keyboard.press('Escape')
  await page.getByRole('tab',{name:'图片创作',exact:true}).click()
  assert.equal(await page.locator('#studio-prompt').inputValue(),'keep image draft')
  await page.getByRole('tab',{name:'视频创作',exact:true}).click()
  assert.equal(await page.locator('[data-video-studio] textarea').inputValue(),'A camera glides across a mountain lake')
  fs.mkdirSync(output,{recursive:true});await page.screenshot({path:`${output}/desktop-empty.png`,fullPage:true})
  await page.getByRole('button',{name:'生成视频',exact:true}).click()
  await page.waitForFunction(()=>document.querySelector('[data-video-record]'))
  await page.waitForTimeout(600)
  assert.equal(generates,1);assert.equal(bodies[0].size,'1280x720');assert.equal(bodies[0].references.length,2);assert.equal(bodies[0].mode,'frames')
  assert.equal((await records())[0].taskID,'task-1')
  const before=polls;completed=true
  await page.reload()
  await page.locator('[data-video-record] video').waitFor({timeout:15000})
  assert.equal(generates,1,'refresh must not repeat a paid create')
  assert(polls>before)
  const rows=await records();assert.equal(rows[0].status,'success');assert.equal(rows[0].video.type,'video/mp4');assert(rows[0].video.size>0);assert.equal(rows[0].width,320);assert.equal(rows[0].height,180);assert.equal(rows[0].videoDurationMs,1000)
  await page.getByRole('button',{name:'加入资产',exact:true}).click()
  await page.getByRole('button',{name:'我的资产',exact:true}).click();await page.getByRole('dialog').waitFor();await page.keyboard.press('Escape')
  await page.screenshot({path:`${output}/desktop-result.png`,fullPage:true})
  await page.evaluate(()=>document.documentElement.classList.add('dark'));assert((await page.locator('.video-workspace').evaluate(el=>getComputedStyle(el).backgroundImage)).includes('0.16'));await page.screenshot({path:`${output}/desktop-dark.png`,fullPage:true})
  await page.setViewportSize({width:390,height:844});await page.waitForTimeout(500);await page.screenshot({path:`${output}/mobile.png`,fullPage:true})
  assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),'mobile must not overflow')
  await page.setViewportSize({width:1512,height:1000});await page.waitForTimeout(500)
  await page.getByRole('button',{name:'创作记录',exact:true}).click();await page.getByRole('button',{name:'删除对话',exact:true}).click();await page.getByRole('button',{name:'确认',exact:true}).click()
  await page.waitForTimeout(200);assert.equal((await records()).length,0)
  await page.goto(`${base}/admin/image-studio?tab=video`)
  await page.getByRole('heading',{name:'视频创作',exact:true}).waitFor()
  await page.getByRole('textbox',{name:'显示名称 grok-imagine-video'}).fill('Video Custom').catch(async()=>{await page.locator('#studio-panel-video table input[type=text]').fill('Video Custom')})
  await page.locator('#studio-panel-video').getByRole('button',{name:'保存',exact:true}).click()
  await page.waitForTimeout(200);assert.equal(savedConfig.models[0]?.label,'Video Custom')
  assert(!writes.includes('/api/v1/admin/image-studio/config'),'video settings must not save image settings')
  assert.equal(errors.length,0,errors.join('\n'))
  console.log(JSON.stringify({passed:true,generates,polls,savedVideos:rows.length,checks:['tabs preserve drafts','fullscreen Escape','reference upload','decoded metadata','settings payload','IndexedDB video blob','refresh resumes without create','assets','history delete','admin independent save','mobile overflow','light/dark'],screenshots:output}))
 }finally{await browser.close()}
})().catch(error=>{console.error(error);process.exitCode=1})
