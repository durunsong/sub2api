/* eslint-disable @typescript-eslint/no-var-requires */
const assert = require('node:assert/strict')
const { chromium } = require('playwright')
;(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const page = await browser.newPage()
    const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
    await page.route('**/__video_storage_test', route => route.fulfill({ contentType: 'text/html', body: '<!doctype html><title>Video storage test</title>' }))
    await page.goto(`${base}/__video_storage_test`)
    const results = await page.evaluate(async () => {
      const { createVideoStudioDB } = await import('/src/features/video-studio/storage.ts')
      const first = createVideoStudioDB(910001), peer = createVideoStudioDB(910001), second = createVideoStudioDB(910002)
      const video = new Blob(['fixture'], { type: 'video/mp4' })
      const record = { id: 'r', generationId: 'g', sessionId: 's', createdAt: 1, prompt: 'local only', model: { id: 'm' }, parameters: {}, references: [], status: 'pending' }
      try {
        await first.put('records', record)
        const isolated = (await second.list('records')).length === 0
        await peer.updateRecord('r', 'g', { video, status: 'success' })
        const saved = (await first.list('records'))[0]
        const blobRoundtrip = saved.video instanceof Blob && await saved.video.text() === 'fixture'
        const staleGenerationRejected = !await first.updateRecord('r', 'old', { prompt: 'stale' })
        await peer.deleteSession('s')
        const deletedNotResurrected = !await first.updateRecord('r', 'g', { video }) && (await peer.list('records')).length === 0
        await first.put('drafts', { id: 'current', prompt: 'draft', references: [{ blob: video }] })
        await first.put('prompts', { id: 'p', title: 'saved', prompt: 'private prompt' })
        await first.put('assets', { id: 'a', kind: 'video', blob: video })
        const libraryIsolated = (await second.list('drafts')).length === 0 && (await second.list('assets')).length === 0 && (await second.list('prompts')).length === 0
        return { isolated, blobRoundtrip, staleGenerationRejected, deletedNotResurrected, libraryIsolated }
      } finally { first.close(); peer.close(); second.close() }
    })
    for (const [name, passed] of Object.entries(results)) assert.equal(passed, true, name)
    console.log(JSON.stringify({ passed: true, ...results }))
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode = 1 })
