import type { ObjectDirective } from 'vue'

let sequence = 0
let closeActive: (() => void) | undefined
const cleanups = new WeakMap<HTMLElement, () => void>()

// Read the accessible label at open time so translated and toggled actions stay in sync.
export const vTooltip: ObjectDirective<HTMLElement> = {
  mounted(el, binding) {
    if (binding.value === false) return
    let bubble: HTMLDivElement | undefined
    let hovered = false
    let focused = false
    let bubbleHovered = false
    let closeTimer: ReturnType<typeof setTimeout> | undefined
    function cancelClose() { clearTimeout(closeTimer); closeTimer = undefined }
    function scheduleClose() {
      cancelClose()
      closeTimer = setTimeout(() => { if (!focused && !hovered && !bubbleHovered) close() }, 150)
    }
    const id = `icon-tooltip-${++sequence}`
    function close() {
      cancelClose()
      if (closeActive === close) closeActive = undefined
      bubble?.remove()
      bubble = undefined
      bubbleHovered = false
      const descriptions = (el.getAttribute('aria-describedby') || '').split(/\s+/).filter(value => value && value !== id)
      if (descriptions.length) el.setAttribute('aria-describedby', descriptions.join(' '))
      else el.removeAttribute('aria-describedby')
    }
    function position() {
      if (!bubble) return
      const rect = el.getBoundingClientRect()
      const box = bubble.getBoundingClientRect()
      const left = Math.max(8, Math.min(innerWidth - box.width - 8, rect.left + (rect.width - box.width) / 2))
      const top = rect.top >= box.height + 16 ? Math.min(innerHeight - box.height - 8, rect.top - box.height - 8) : Math.min(innerHeight - box.height - 8, rect.bottom + 8)
      bubble.style.left = `${left}px`
      bubble.style.top = `${Math.max(8, top)}px`
    }
    function open() {
      const label = el.getAttribute('aria-label')
      if (!label) return
      cancelClose()
      if (closeActive !== close) closeActive?.()
      closeActive = close
      if (!bubble) {
        bubble = document.createElement('div')
        bubble.id = id
        bubble.role = 'tooltip'
        bubble.className = 'fixed z-[99999] max-w-[calc(100vw-16px)] break-words rounded-lg bg-gray-900 px-3 py-2 text-xs leading-relaxed text-white shadow-xl ring-1 ring-white/10 dark:bg-gray-800'
        bubble.addEventListener('mouseenter', () => { bubbleHovered = true; cancelClose() })
        bubble.addEventListener('mouseleave', () => { bubbleHovered = false; scheduleClose() })
        document.body.append(bubble)
        el.setAttribute('aria-describedby', [el.getAttribute('aria-describedby'), id].filter(Boolean).join(' '))
      }
      bubble.textContent = label
      position()
    }
    const enter = () => { hovered = true; open() }
    const leave = () => { hovered = false; if (!focused) scheduleClose() }
    const focus = () => { focused = true; open() }
    const blur = () => { focused = false; if (!hovered) scheduleClose() }
    const escape = (event: KeyboardEvent) => {
      if (event.key !== 'Escape' || !bubble) return
      close()
    }
    el.addEventListener('mouseenter', enter)
    el.addEventListener('mouseleave', leave)
    el.addEventListener('focus', focus)
    el.addEventListener('blur', blur)
    el.addEventListener('click', close)
    document.addEventListener('keydown', escape, true)
    window.addEventListener('scroll', position, true)
    window.addEventListener('resize', position)
    cleanups.set(el, () => {
      close()
      el.removeEventListener('mouseenter', enter)
      el.removeEventListener('mouseleave', leave)
      el.removeEventListener('focus', focus)
      el.removeEventListener('blur', blur)
      el.removeEventListener('click', close)
      document.removeEventListener('keydown', escape, true)
      window.removeEventListener('scroll', position, true)
      window.removeEventListener('resize', position)
    })
  },
  beforeUnmount(el) { cleanups.get(el)?.(); cleanups.delete(el) },
}
