import { platformLabel } from '@/utils/platformColors'

export function subscriptionPlanBadgeLabel(platform: string): string {
  if (platform === 'anthropic') return 'GLM'
  return platformLabel(platform)
}

export function subscriptionPlanFilterLabel(platform: string): string {
  if (platform === 'kiro') return 'Claude(Max 5x)'
  if (platform === 'anthropic') return 'Claude(GLM coding Max)'
  if (platform === 'openai') return 'OpenAI(GPT Pro 20x)'
  return platformLabel(platform)
}
