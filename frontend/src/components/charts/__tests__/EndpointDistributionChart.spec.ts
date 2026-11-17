import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import EndpointDistributionChart from '../EndpointDistributionChart.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('vue-chartjs', () => ({ Doughnut: { template: '<div />' } }))
vi.mock('@/api/admin/dashboard', () => ({ getUserBreakdown: vi.fn().mockResolvedValue([]) }))

const endpointStats = [{ endpoint: '/v1/responses', requests: 2, total_tokens: 100, cost: 9, actual_cost: 3, account_cost: 1 }]

describe('EndpointDistributionChart cost visibility', () => {
  it('keeps actual charge, account cost and standard cost separate for admins', async () => {
    const wrapper = mount(EndpointDistributionChart, { props: { endpointStats } })
    expect(wrapper.findAll('thead th')).toHaveLength(6)
    const cells = wrapper.find('tbody tr').findAll('td')
    expect(cells.slice(3).map(cell => cell.text())).toEqual(['$3.00', '$1.00', '$9.00'])
    await wrapper.find('tbody tr').trigger('click')
    await flushPromises()
    expect(wrapper.find('td[colspan]').attributes('colspan')).toBe('6')
  })

  it('preserves the fork user cost label and hides internal cost columns', () => {
    const wrapper = mount(EndpointDistributionChart, {
      props: { endpointStats, showAccountCost: false, showStandardCost: false, costColumnLabelKey: 'usage.cost', enableBreakdown: false },
    })
    expect(wrapper.findAll('thead th')).toHaveLength(4)
    expect(wrapper.findAll('thead th')[3].text()).toBe('usage.cost')
    expect(wrapper.find('tbody tr').findAll('td').slice(3).map(cell => cell.text())).toEqual(['$3.00'])
    expect(wrapper.text()).not.toContain('admin.dashboard.accountCost')
    expect(wrapper.text()).not.toContain('admin.dashboard.standard')
  })
})
