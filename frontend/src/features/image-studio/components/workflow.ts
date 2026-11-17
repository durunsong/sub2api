export function extractVariables(template: string): string[] {
  return [...new Set([...template.matchAll(/\{\{\s*([^{}]+?)\s*\}\}/g)].map((match) => match[1].trim()))]
}

export function renderTemplate(template: string, values: Record<string, string>): string {
  return template.replace(/\{\{\s*([^{}]+?)\s*\}\}/g, (_match, key: string) => values[key.trim()] || '')
}

export function splitSeries(text: string): string[] {
  return text.split(/^\s*---\s*$/m).map((item) => item.trim()).filter(Boolean)
}
