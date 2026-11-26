import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { expect, it, vi } from 'vitest'
import VideoComposer from './VideoComposer.vue'
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
it('closes settings on composer clicks, Escape and tab deactivation, but not inside settings', async () => {
  const wrapper = mount(VideoComposer, { attachTo: document.body, props: { prompt:'', modelID:'', models:[], parameters:{resolution:'720',size:'auto',seconds:6,mode:'frames'}, references:[], canGenerate:false, running:false, inline:true, active:true }, global:{stubs:{Select:true,Icon:true,VideoSettings:{template:'<button data-setting>control</button>'}}} })
  const trigger = wrapper.get('.video-parameter-trigger')
  await trigger.trigger('click')
  const control = document.querySelector('[data-setting]')!
  control.dispatchEvent(new Event('pointerdown',{bubbles:true}))
  await nextTick()
  expect(document.querySelector('.video-settings-popover')).not.toBeNull()
  await wrapper.get('textarea').trigger('pointerdown')
  expect(document.querySelector('.video-settings-popover')).toBeNull()
  await trigger.trigger('click')
  await wrapper.setProps({active:false})
  expect(document.querySelector('.video-settings-popover')).toBeNull()
  await wrapper.setProps({active:true})
  await trigger.trigger('click')
  document.querySelector('[data-setting]')!.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',bubbles:true}))
  await nextTick()
  expect(document.querySelector('.video-settings-popover')).toBeNull()
  expect(document.activeElement).toBe(trigger.element)
  wrapper.unmount()
})
