export const getApiBase = () => {
  if (window.location.port === '5173' || window.location.port === '5174') {
    const protocol = window.location.protocol === 'https:' ? 'https:' : 'http:'
    const port = window.location.protocol === 'https:' ? '8443' : '8080'
    return `${protocol}//${window.location.hostname}:${port}`
  }
  return ''
}
