import { platformLabel } from '@/utils/platformColors'

export function subscriptionPlanBadgeLabel(
  platform: string,
  plan?: { group_name?: string; name?: string; description?: string },
): string {
  const isOpenAICompatibleGlm = platform === 'openai'
    && [plan?.group_name, plan?.name, plan?.description]
      .some(value => String(value || '').toLowerCase().includes('glm'))
  if (platform === 'kiro' || platform === 'anthropic') return 'Claude Max'
  if (isOpenAICompatibleGlm) return 'GLM'
  return platformLabel(platform)
}

export function subscriptionPlanFilterLabel(platform: string): string {
  if (platform === 'kiro' || platform === 'anthropic' || platform === 'claude') return 'Claude Max'
  if (platform === 'openai') return 'OpenAI(GPT Pro 20x)'
  return platformLabel(platform)
}
