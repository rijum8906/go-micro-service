export function safeInternalRedirect(
  redirect: string | undefined,
  fallback = '/',
): string {
  const target = redirect?.trim()
  if (!target) return fallback

  if (
    !target.startsWith('/') ||
    target.startsWith('//') ||
    target.includes('\\')
  ) {
    return fallback
  }

  return target
}
