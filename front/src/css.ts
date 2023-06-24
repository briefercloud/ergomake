export function classNames(...names: string[]): string {
  return names.filter(Boolean).join(' ')
}
