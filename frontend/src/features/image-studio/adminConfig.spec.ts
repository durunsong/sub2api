import { describe, expect, it } from 'vitest'
import { mergeModelSettings, updateModelSetting } from './adminConfig'
import type { StudioModel, WorkflowModel } from './api'
const model = { group_id: 3, model: 'image1', label: 'Image One', protocol: 'images' } as StudioModel

describe('studio presentation overrides', () => {
  it('retains text models without assigning an image protocol', () => {
    const text = { id: 'text', group_id: 3, group_name: 'Text', model: 'gpt-4', label: 'GPT', supports_references: true } as WorkflowModel
    expect(mergeModelSettings([model, text], []).find(item => item.model === 'gpt-4')).toEqual({ group_id: 3, model: 'gpt-4', label: 'gpt-4', enabled: true })
  })
  it('shows automatic models and retains hidden settings when a candidate disappears', () => {
    expect(mergeModelSettings([model], [{ group_id: 4, model: 'old', enabled: false, label: 'Old' }])).toEqual([
      { group_id: 3, model: 'image1', label: 'image1', protocol: 'images', enabled: true },
      { group_id: 4, model: 'old', enabled: false, label: 'Old' }
    ])
  })
  it('restores the model name after removing an override from an already labeled catalog', () => {
    const catalog = [{ ...model, label: 'Custom display name' }]
    const overrides = [{ group_id: 3, model: 'image1', enabled: false, label: 'Custom display name' }]
    expect(mergeModelSettings(catalog, overrides)[0].label).toBe('Custom display name')
    expect(mergeModelSettings(catalog, [])[0]).toEqual({
      group_id: 3, model: 'image1', label: 'image1', protocol: 'images', enabled: true
    })
  })
  it('hides a specific group/model without hiding an identically named model in another group', () => {
    const updated = updateModelSetting([{ group_id: 4, model: 'image1', enabled: true, label: 'Other' }], { group_id: 3, model: 'image1', enabled: false, label: 'Image One' })
    expect(updated).toHaveLength(2)
    expect(mergeModelSettings([model], updated).find(item => item.group_id === 3)?.enabled).toBe(false)
    expect(updated.find(item => item.group_id === 4)?.enabled).toBe(true)
  })
})
