export const getApiBase = () => {
  if (window.location.port === '5173') {
    return `http://${window.location.hostname}:8080`
  }
  return ''
}
