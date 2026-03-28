export const getApiBase = () => {
  if (window.location.port === '5173' || window.location.port === '5174') {
    return `http://${window.location.hostname}:8080`
  }
  return ''
}
