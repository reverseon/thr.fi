import { useLocation } from "preact-iso";
import { useContext, useEffect, useRef, useState } from "preact/hooks";
import '../css/general.css';
import { useFetchWrapper } from "../utils/usefetchwrapper";
import toast from "react-hot-toast";
import { SuccessShortenContext } from "../app";

export function Home() {
  const location = useLocation();
  const [show_password, setShowPassword] = useState<boolean>(false)

  // Context
  const shortenResultContext = useContext(SuccessShortenContext)

  // Form Inputs
  const original_url = useRef<HTMLInputElement>(null)
  const backhalf = useRef<HTMLInputElement>(null)
  const password = useRef<HTMLInputElement>(null)

  // Fetch Components
  const [user_info_loading, setUserInfoLoading] = useState<boolean>(true)
  const [user_info_failed, setUserInfoFailed] = useState<boolean>(false)
  const [user_info, setUserInfo] = useState<{
    email: string,
    id: string
  } | undefined>(undefined)

  const [shorten_loading, setShortenLoading] = useState<boolean>(false)
  const [shorten_result, setShortenResult] = useState<{
    "backhalf": string,
    "original_url": string,
    "password_protected": boolean,
  } | undefined>(undefined)

  // Login URL Construction
  const access_token_handler = `${import.meta.env.VITE_APP_URL}/page/set`;
  const redirect_ep = `/api/redirect`
  const constructed_redirectToPath = `${redirect_ep}?c=t&to=${encodeURIComponent(access_token_handler)}`
  const login_url = `${import.meta.env.VITE_CENTRAL_AUTH_URL}/auth?showtab=signin&redirectToPath=${encodeURIComponent(constructed_redirectToPath)}`;

  // Logout URL Construction
  const logout_callback = `${import.meta.env.VITE_APP_URL}/page/logout?to=${encodeURIComponent("/")}`
  const logout_url = `${import.meta.env.VITE_CENTRAL_AUTH_URL}/?ras=${encodeURIComponent(logout_callback)}`

  // init
  useEffect(() => {
  async function init() {
    const user_token = localStorage.getItem('user_token')
    useFetchWrapper<{
      email: string,
      id: string
    }>(
      setUserInfoLoading,
      setUserInfoFailed,
      setUserInfo,
      import.meta.env.VITE_BACKEND_URL + '/user/info', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${user_token}`
      }
    })
  }
  init()
  }, [])

  async function doShorten() {
    fetch(import.meta.env.VITE_BACKEND_URL + '/url/create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('user_token')}`
      },
      body: JSON.stringify({
        original_url: original_url.current?.value,
        backhalf: backhalf.current?.value,
        password_protected: show_password,
        password: password.current?.value
      })
    }).then(async (response) => {
      const data = await response.json()
      if (response.ok) {
        setShortenResult(data)
        shortenResultContext.setShortenResult(data.backhalf)
        location.route('/page/result')
      } else if (response.status === 400) {
        toast.error('Failed to shorten due to invalid input. Please check your URL, backhalf, and password.')
      } else if (response.status === 500) {
        toast.error('Failed to shorten due to server error')
      } else {
        toast.error(`Failed to shorten: ${data.error}`)
      }
    }).catch(() => {
      toast.error('Failed to shorten due to network error')
    })
  }

  return (
    <div className="container">
      <div className="card">
        <h2 className="pure-u-1">Shorten an URL</h2>
        <form className="pure-form pure-form-stacked" onSubmit={
          (e) => {
            e.preventDefault()
          }
        } autocomplete="off">
          <fieldset>
            <legend>
              {
              user_info_loading ?
                <p className="info">Loading...</p>
              : user_info ?
                <>
                <p className="info">Authenticated as {user_info.email}.</p>
                <p style={{
                  textAlign: 'right'
                }}>
                  <a className="pure-anchor-style" href={
                    logout_url
                  }>Logout</a> | <span className="pure-anchor-style" onClick={
                    (e) => {
                      e.preventDefault()
                      location.route('/page/manage')
                    }
                  }>Manage URL</span>
                </p>
                </>
              :
                <>
                  <p>
                    <a href={login_url} className="pure-anchor-style">Login</a> to manage your shortened URLs
                  </p>
                </>
              }
            </legend>
            <input type="url" ref={original_url} name="original_url" placeholder="Original URL" style={{
              width: '100%'
            }}/>
            <span className="pure-form-message">Please include http(s) in your link</span>
            <input type="text" ref={backhalf} placeholder="Backhalf" name="backhalf" style={{
              width: '100%'
            }}/>
            <span className="pure-form-message">Shorten as thr.fi/Backhalf, 1 to 20 alphanumeric characters.</span>
            <label for="enable-password" className="pure-checkbox" style={{
              display: 'block',
              marginBottom: '1em'
            }}>
                <input onChange={
                  (e) => {
                    setShowPassword(e.currentTarget.checked)
                  }
                } type="checkbox" id="enable-password"
                checked={show_password}
                /> Password Protected?
            </label>
            {show_password &&
            <>
              <input type="password" ref={password} name="password" placeholder="Password" style={{
                width: '100%'
              }}/>
              <span className="pure-form-message">8-32 characters</span>
            </>
            }
            <button type="button" className="pure-button pure-button-primary" onClick={
              (e) => {
                e.preventDefault()
                doShorten()
              }
            }>Shorten</button>
          </fieldset>
        </form>
        <div className="info">
          Made by <a className="pure-anchor-style" href="https://www.naj.one">Thirafi Najwan</a> for personal use. Don't use it for important stuff as I don't guarantee the uptime of this service.
        </div>
      </div>
    </div>
  )
}