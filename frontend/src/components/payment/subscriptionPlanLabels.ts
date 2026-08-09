import { platformLabel } from '@/utils/platformColors'

export function subscriptionPlanBadgeLabel(
  platform: string,
  plan?: { group_name?: string; name?: string; description?: string },
): string {
  const isOpenAICompatibleGlm = platform === 'openai'
    && [plan?.group_name, plan?.name, plan?.description]
      .some(value => String(value || '').toLowerCase().includes('glm'))
  if (platform === 'anthropic' || isOpenAICompatibleGlm) return 'GLM'
  return platformLabel(platform)
}

export function subscriptionPlanFilterLabel(platform: string): string {
  if (platform === 'kiro') return 'Claude(Max 5x)'
  if (platform === 'anthropic') return 'Claude(GLM coding Max)'
  if (platform === 'openai') return 'OpenAI(GPT Pro 20x)'
  return platformLabel(platform)
}
