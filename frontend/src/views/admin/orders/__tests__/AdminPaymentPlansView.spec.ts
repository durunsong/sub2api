import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminPaymentPlansView from '../AdminPaymentPlansView.vue'

const { getPlans, getConfig, getGroups } = vi.hoisted(() => ({
  getPlans: vi.fn(),
  getConfig: vi.fn(),
  getGroups: vi.fn(),
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getPlans,
    getConfig,
  },
}))

vi.mock('@/api/admin', () => ({
  default: {
    groups: {
      getAll: getGroups,
    },
  },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const DataTableStub = {
  props: ['data'],
  template: `
    <div data-testid="plans-table">
      <div v-for="row in data" :key="row.id" :data-plan-id="row.id">
        <slot name="cell-price" :value="row.price" :row="row" />
        <span>{{ row.name }}</span>
      </div>
    </div>
  `,
}

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue'],
  template: `
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target).value)"
    >
      <option v-for="opt in options" :key="String(opt.value)" :value="opt.value">{{ opt.label }}</option>
    </select>
  `,
}

function mountView() {
  return mount(AdminPaymentPlansView, {
    global: {
      plugins: [createPinia()],
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        DataTable: DataTableStub,
        ConfirmDialog: true,
        GroupBadge: true,
        Icon: true,
        PlanEditDialog: true,
        Select: SelectStub,
      },
    },
  })
}

describe('AdminPaymentPlansView', () => {
  beforeEach(() => {
    getGroups.mockResolvedValue([
      { id: 1, name: 'Claude Group', platform: 'kiro' },
      { id: 2, name: 'Grok Group', platform: 'grok' },
    ])
    getConfig.mockResolvedValue({ data: {} })
    getPlans.mockResolvedValue({
      data: [
        {
          id: 1,
          name: 'CNY plan',
          group_id: 1,
          price: 499,
          original_price: 599,
          currency: 'CNY',
          validity_days: 30,
          validity_unit: 'day',
          sort_order: 0,
          for_sale: true,
          features: [],
        },
        {
          id: 2,
          name: 'Legacy plan',
          group_id: 2,
          price: 10,
          original_price: 0,
          currency: '',
          validity_days: 30,
          validity_unit: 'day',
          sort_order: 0,
          for_sale: false,
          features: [],
        },
      ],
    })
  })

  it('uses the configured currency symbol and keeps legacy prices in USD', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('¥499.00CNY')
    expect(wrapper.text()).toContain('¥599.00')
    expect(wrapper.text()).toContain('$10.00')
  })

  it('filters plans by search keyword and sale status', async () => {
    const wrapper = mountView()
    await flushPromises()

    const table = wrapper.get('[data-testid="plans-table"]')
    expect(table.findAll('[data-plan-id]')).toHaveLength(2)

    await wrapper.get('input').setValue('CNY')
    expect(table.findAll('[data-plan-id]')).toHaveLength(1)
    expect(table.text()).toContain('CNY plan')
    expect(table.text()).not.toContain('Legacy plan')

    await wrapper.get('input').setValue('Grok')
    expect(table.findAll('[data-plan-id]')).toHaveLength(1)
    expect(table.text()).toContain('Legacy plan')

    await wrapper.get('input').setValue('')
    await wrapper.get('select').setValue('true')
    expect(table.findAll('[data-plan-id]')).toHaveLength(1)
    expect(table.text()).toContain('CNY plan')
  })
})
