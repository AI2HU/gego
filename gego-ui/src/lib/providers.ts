const providerStyles: Record<string, { badge: string; gradient: string; initial: string }> = {
  openai: {
    badge: 'bg-emerald-100 text-emerald-800 border-emerald-200',
    gradient: 'from-emerald-500 to-teal-600',
    initial: 'O',
  },
  anthropic: {
    badge: 'bg-orange-100 text-orange-800 border-orange-200',
    gradient: 'from-orange-500 to-amber-600',
    initial: 'A',
  },
  ollama: {
    badge: 'bg-violet-100 text-violet-800 border-violet-200',
    gradient: 'from-violet-500 to-purple-600',
    initial: 'L',
  },
  google: {
    badge: 'bg-blue-100 text-blue-800 border-blue-200',
    gradient: 'from-blue-500 to-indigo-600',
    initial: 'G',
  },
  perplexity: {
    badge: 'bg-cyan-100 text-cyan-800 border-cyan-200',
    gradient: 'from-cyan-500 to-sky-600',
    initial: 'P',
  },
}

const defaultStyle = {
  badge: 'bg-slate-100 text-slate-800 border-slate-200',
  gradient: 'from-slate-500 to-slate-600',
  initial: '?',
}

export function getProviderStyle(provider: string) {
  return providerStyles[provider] ?? defaultStyle
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
