import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import SubscriptionsView from '../SubscriptionsView.vue'

const { listSubscriptions, getAllGroups } = vi.hoisted(() => ({
  listSubscriptions: vi.fn(),
  getAllGroups: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    subscriptions: {
      list: listSubscriptions
    },
    groups: {
      getAll: getAllGroups
    },
    usage: {
      searchUsers: vi.fn()
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: `
    <div>
      <button
        v-for="option in options"
        :key="String(option.value)"
        :data-option="option.value"
        @click="$emit('update:modelValue', option.value); $emit('change', option.value, option)"
      >
        {{ option.label }}
      </button>
    </div>
  `
}

const mountSubscriptionsView = () => mount(SubscriptionsView, {
  global: {
    stubs: {
      AppLayout: { template: '<div><slot /></div>' },
      TablePageLayout: {
        template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
      },
      DataTable: { template: '<div />' },
      Pagination: true,
      BaseDialog: true,
      ConfirmDialog: true,
      EmptyState: true,
      Select: SelectStub,
      GroupBadge: true,
      GroupOptionItem: true,
      Icon: true,
      RouterLink: true,
      Teleport: true
    }
  }
})

describe('admin SubscriptionsView filters', () => {
  beforeEach(() => {
    localStorage.clear()
    listSubscriptions.mockReset()
    getAllGroups.mockReset()
    listSubscriptions.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    })
    getAllGroups.mockResolvedValue([])
  })

  it('defaults to active subscriptions with quota remaining', async () => {
    mountSubscriptionsView()
    await flushPromises()

    const filters = listSubscriptions.mock.calls.at(-1)?.[2]
    expect(filters).toMatchObject({ status: 'active_available' })
  })

  it('requests active subscriptions with quota remaining from the combined status option', async () => {
    const wrapper = mountSubscriptionsView()
    await flushPromises()

    const option = wrapper.find('[data-option="active_available"]')
    expect(option.exists()).toBe(true)
    expect(option.text()).toBe('admin.subscriptions.status.activeAvailable')

    await option.trigger('click')
    await flushPromises()

    const filters = listSubscriptions.mock.calls.at(-1)?.[2]
    expect(filters).toMatchObject({ status: 'active_available' })
  })
})
