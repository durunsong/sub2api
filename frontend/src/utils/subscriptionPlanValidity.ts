import type { ComposerTranslation } from 'vue-i18n'

const UNIT_LABEL_KEY: Record<string, string> = {
  day: 'payment.admin.days',
  days: 'payment.admin.days',
  week: 'payment.admin.weeks',
  weeks: 'payment.admin.weeks',
  month: 'payment.admin.months',
  months: 'payment.admin.months',
  year: 'payment.years',
  years: 'payment.years',
}

export function formatPlanValiditySuffix(
  validityDays: number,
  validityUnit: string | undefined | null,
  t: ComposerTranslation,
): string {
  const unit = (validityUnit || 'day').toLowerCase()
  const labelKey = UNIT_LABEL_KEY[unit] ?? UNIT_LABEL_KEY.day
  return `${validityDays}${t(labelKey)}`
}
