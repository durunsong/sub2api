/* eslint-disable @typescript-eslint/no-var-requires -- CommonJS resolves the host-provided Playwright bundle via NODE_PATH. */
// Run against the local Vite server; uses an existing Playwright installation.
// NODE_PATH may point at the host's bundled browser automation dependencies.
const assert = require('node:assert/strict')
const { chromium } = require('playwright')

;(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' })
  try {
    const context = await browser.newContext()
    const page = await context.newPage()
    const base = process.env.STUDIO_TEST_URL || 'http://127.0.0.1:5179'
    await page.route('**/__studio_storage_test', route => route.fulfill({ contentType: 'text/html', body: '<!doctype html><title>Image studio storage checks</title>' }))
    await page.goto(`${base}/__studio_storage_test`)
    const result = await page.evaluate(async () => {
      const { createImageStudioDB } = await import('/src/features/image-studio/storage.ts')
      const canvas = document.createElement('canvas'); canvas.width = 16; canvas.height = 16
      const ctx = canvas.getContext('2d'); ctx.fillStyle = '#4f46e5'; ctx.fillRect(0, 0, 16, 16)
      const image = await new Promise(resolve => canvas.toBlob(resolve, 'image/png'))
      const first = createImageStudioDB(900001), second = createImageStudioDB(900002)
      const snapshot = { groupID: 7, modelID: '7:gpt-image-2', model: 'gpt-image-2', prompt: 'test', size: '1024x1024', quality: 'auto', count: 1 }
      const history = { id: 'history-a', createdAt: Date.now(), status: 'success', snapshot, images: [image], references: [image], favorite: false, categoryID: 'category-a', categoryIDs: ['category-a','category-b'] }
      await first.saveCategory({ id: 'category-a', name: 'A' }); await first.saveCategory({ id: 'category-b', name: 'B' })
      await first.saveTask({ id: history.id, createdAt: history.createdAt, snapshot, references: [], status: 'running' })
      await first.saveHistory(history)
      const savedTaskRetired = (await first.listTasks()).length === 0
      // Metadata updates must read/write in one transaction, never restore a stale deleted Blob.
      const peer = createImageStudioDB(900001)
      await peer.deleteHistory(history.id)
      await first.toggleFavorite(history.id)
      await first.assignCategories([history.id], ['category-a'])
      const metadataCannotReviveDeleted = (await first.listHistory()).length === 0
      await first.saveHistory(history)
      await Promise.all([first.toggleFavorite(history.id), peer.toggleFavorite(history.id)])
      const concurrentFavoriteAtomic = (await first.listHistory())[0].favorite === false
      await peer.deleteCategory('category-b')
      await first.assignCategories([history.id], ['category-a', 'category-b'])
      const deletedCategoryNotRestored = (await first.listHistory())[0].categoryIDs.join(',') === 'category-a'
      await first.saveCategory({ id: 'category-b', name: 'B' })
      await first.assignCategories([history.id], ['category-a', 'category-b'])
      peer.close()
      await first.saveAsset({ id: 'asset-a', name: 'Reference', createdAt: Date.now(), image, categoryID: 'category-a' })
      await first.saveDraft({ id: 'current', snapshot, references: [image], layout: 'bottom', updatedAt: Date.now() })
      await first.savePrompt({ id: 'prompt-a', title: 'Product', prompt: '{{subject}} on white', seriesPrompts: ['front','side'], createdAt: Date.now() })
      await first.saveTask({ id: 'task-a', createdAt: Date.now(), snapshot, references: [], status: 'running' })
      const saved = await first.listHistory()
      const accountIsolated = (await second.listHistory()).length === 0
      const blobRoundTrip = saved[0].images[0] instanceof Blob && (await createImageBitmap(saved[0].images[0])).width === 16
      first.close()
      const reopened = createImageStudioDB(900001)
      const reloadWorks = (await reopened.listHistory()).length === 1 && (await reopened.getDraft()).layout === 'bottom'
      await reopened.deleteCategory('category-a')
      const classified = (await reopened.listHistory())[0]
      const categoryDeletePreservesImage = classified.categoryID === null && classified.categoryIDs.join(',') === 'category-b' && classified.images.length === 1
      const backup = await reopened.exportBackup()
      await second.importBackup(backup)
      const backupRestored = (await second.listHistory()).length === 1 && (await second.listAssets()).length === 1 && (await second.listPrompts()).length === 1 && (await second.listTasks())[0].status === 'interrupted'
      let invalidRejected = false
      try { await second.importBackup(new Blob(['{"version":1,"history":[{"id":"malicious","images":["https://evil.invalid/x.png"]}]}'], {type:'application/json'})) } catch { invalidRejected = true }
      const invalidImportAtomic = invalidRejected && (await second.listHistory()).length === 1
      // Simulate external manipulation: one missing Blob and one truncated image.
      const raw = await new Promise((resolve,reject) => { const request = indexedDB.open('sub2api-image-studio-900001',1); request.onsuccess=()=>resolve(request.result); request.onerror=()=>reject(request.error) })
      await new Promise((resolve,reject) => { const tx = raw.transaction('history','readwrite'); tx.objectStore('history').put({...history,id:'missing',images:[]}); tx.objectStore('history').put({...history,id:'truncated',images:[new Blob([new Uint8Array([137,80,78,71,13,10,26,10])],{type:'image/png'})]}); tx.oncomplete=resolve;tx.onabort=()=>reject(tx.error) })
      const corruptFiltered = (await reopened.listHistory()).length === 1
      await new Promise((resolve,reject) => { const tx = raw.transaction('history','readwrite');tx.objectStore('history').clear();tx.oncomplete=resolve;tx.onabort=()=>reject(tx.error) })
      const externalClearVisible = (await reopened.listHistory()).length === 0
      raw.close(); reopened.close(); second.close()
      return { metadataCannotReviveDeleted, concurrentFavoriteAtomic, deletedCategoryNotRestored, accountIsolated, blobRoundTrip, savedTaskRetired, reloadWorks, categoryDeletePreservesImage, backupRestored, invalidImportAtomic, corruptFiltered, externalClearVisible }
    })
    for (const [name, passed] of Object.entries(result)) assert.equal(passed, true, name)
    console.log(JSON.stringify(result, null, 2))
  } finally { await browser.close() }
})().catch(error => { console.error(error); process.exitCode = 1 })
