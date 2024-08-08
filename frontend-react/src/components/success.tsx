import { useLocation } from 'preact-iso'
import '../css/general.css'
import { useContext } from 'preact/hooks'
import { SuccessShortenContext } from '../app'
export function Success() {
  const location = useLocation()
  const {
    shorten_result,
  } = useContext(SuccessShortenContext)
  if (!shorten_result) {
    location.route('/page/home')
    return null
  }

  return (
    <div className="container">
      <div className="card">
        <h2 className="pure-u-1">Shortened 🎉</h2>
        <div class="url-container">
            <div class="url">{
              `${import.meta.env.VITE_APP_URL}/${shorten_result}`
            }</div>
            <i className="bi bi-clipboard" onClick={
              (e) => {
                e.preventDefault()
                navigator.clipboard.writeText(`${import.meta.env.VITE_APP_URL}/${shorten_result}`)
              }
            }></i>
        </div>
        <img className="qr-img pure-img" src={`https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${
          encodeURIComponent(`${import.meta.env.VITE_APP_URL}/${shorten_result}`)
        }`} alt="QR Code"></img>
        <div style={{
          marginBottom: '1em',
          width: '100%',
          display: 'flex',
          justifyContent: 'space-between'
        }}>
          <div className="pure-button pure-button-primary" onClick={
            (e) => {
              e.preventDefault()
              location.route('/page/home')
            }
          }>Shorten Another</div>
          <div className="pure-button" onClick={
            (e) => {
              e.preventDefault()
              window.open(`${import.meta.env.VITE_APP_URL}/${shorten_result}`)
            }
          }>Go to Link</div>
        </div>
        <div className="info">
          Made by <a className="pure-anchor-style" href="https://www.naj.one">Thirafi Najwan</a> for personal use. Don't use it for important stuff as I don't guarantee the uptime of this service.
        </div>
      </div>
    </div>
  )

}