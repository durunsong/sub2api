import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import admin from './admin'
import misc from './misc'
import fork from './fork'

const isRecord = (value: unknown): value is Record<string, unknown> =>
  !!value && typeof value === 'object' && !Array.isArray(value)

const mergeMessages = <T extends Record<string, unknown>>(base: T, override: Record<string, unknown>): T => {
  for (const [key, value] of Object.entries(override)) {
    if (isRecord(value) && isRecord(base[key])) {
      mergeMessages(base[key] as Record<string, unknown>, value)
    } else {
      const target = base as Record<string, unknown>
      target[key] = value
    }
  }
  return base
}

export default mergeMessages({
  ...landing,
  ...common,
  ...dashboard,
  admin,
  ...misc,
}, fork)
