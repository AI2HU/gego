const providerStyles: Record<string, { badge: string; gradient: string }> = {
  openai: {
    badge: 'bg-emerald-100 text-emerald-800 border-emerald-200',
    gradient: 'from-emerald-500 to-teal-600',
  },
  anthropic: {
    badge: 'bg-orange-100 text-orange-800 border-orange-200',
    gradient: 'from-orange-500 to-amber-600',
  },
  ollama: {
    badge: 'bg-violet-100 text-violet-800 border-violet-200',
    gradient: 'from-violet-500 to-purple-600',
  },
  google: {
    badge: 'bg-blue-100 text-blue-800 border-blue-200',
    gradient: 'from-blue-500 to-indigo-600',
  },
  perplexity: {
    badge: 'bg-cyan-100 text-cyan-800 border-cyan-200',
    gradient: 'from-cyan-500 to-sky-600',
  },
}

const providerLogos: Record<string, string> = {
  openai: '/providers/openai.svg',
  anthropic: '/providers/anthropic.svg',
  ollama: '/providers/ollama.svg',
  google: '/providers/google.svg',
  perplexity: '/providers/perplexity.svg',
}

const defaultStyle = {
  badge: 'bg-slate-100 text-slate-800 border-slate-200',
  gradient: 'from-slate-500 to-slate-600',
}

const defaultLogo = '/providers/default.svg'

export function getProviderStyle(provider: string) {
  return providerStyles[provider] ?? defaultStyle
}

export function getProviderLogo(provider: string): string {
  return providerLogos[provider] ?? defaultLogo
}

export function formatProviderName(provider: string): string {
  const names: Record<string, string> = {
    openai: 'OpenAI',
    anthropic: 'Anthropic',
    ollama: 'Ollama',
    google: 'Google',
    perplexity: 'Perplexity',
  }
  return names[provider] ?? provider
}
