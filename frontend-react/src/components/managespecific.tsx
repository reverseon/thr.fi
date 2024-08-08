import { useLocation, useRoute } from "preact-iso"
import { useEffect, useRef, useState } from "preact/hooks"
import { useFetchWrapper } from "../utils/usefetchwrapper"
import Swal from "sweetalert2"


export function ManageSpecific() {
    const location = useLocation()
    const backhalf = useRoute().params.backhalf

    const [show_password, setShowPassword] = useState<boolean>(false)

    const original_url_ref = useRef<HTMLInputElement>(null)
    const backhalfref = useRef<HTMLInputElement>(null)
    const passwordref = useRef<HTMLInputElement>(null)

    // Fetch Declaration
    const [user_info_loading, setUserInfoLoading] = useState<boolean>(true)
    const [user_info_failed, setUserInfoFailed] = useState<boolean>(false)
    const [user_info, setUserInfo] = useState<{
      email: string,
      id: string
    } | undefined>(undefined)

    const [data_loading, setDataLoading] = useState<boolean>(true)
    const [data_failed, setDataFailed] = useState<boolean>(false)
    const [data, setData] = useState<{
        backhalf: string,
        password_protected: boolean,
        original_url: string
    } | undefined>(undefined)
  
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
      useFetchWrapper<{
        backhalf: string,
        password_protected: boolean,
        original_url: string
      }>(
        setDataLoading,
        setDataFailed,
        setData,
        import.meta.env.VITE_BACKEND_URL + `/url/${backhalf}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${user_token}`
        }
      })
    }
    init()
    }, [])
    return (
      <div className="container">
        <div className="card">
          {
          (data_loading || user_info_loading) ?
          <p>Loading...</p> :
          (data_failed || user_info_failed) ?
          <p>
            You don't have access. <span className="pure-anchor-style" onClick={
              (e) => {
                e.preventDefault()
                location.route('/page/home')
              }
            }>Back to home</span>
          </p> :
          <>
          <h2 className="pure-u-1">
            {`Manage thr.fi/${data?.backhalf}`}
          </h2>
          <form className="pure-form pure-form-stacked">
            <fieldset>
              <legend>
                {user_info && 
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
                        setShowPassword(false)
                        location.route('/page/manage')
                      }
                    }>Back to URL Manager</span> 
                  </p>
                  </>
                }
              </legend>
              <input type="url" name="original_url" ref={original_url_ref} placeholder="Original URL" style={{
              width: '100%'
              }} value={data?.original_url}/>
              <span className="pure-form-message">Please include http(s) in your link</span>
              <input type="text" name="backhalf" ref={backhalfref} placeholder="Backhalf" style={{
                width: '100%'
              }} value={data?.backhalf}/>
              <span className="pure-form-message">Shorten as thr.fi/Backhalf</span>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                marginBottom: '1em'
              }}>
                { data?.password_protected === false &&
                <button type="button" className="pure-button" onClick={
                  (e) => {
                    e.preventDefault()
                    e.currentTarget.classList.toggle('button-error')
                    setShowPassword(!show_password)
                  }
                }>
                  {
                  show_password ?
                  "Cancel" :
                  "Enable Password"
                  }
                </button>
                }
                { data?.password_protected === true &&
                <>
                <button type="button" className="pure-button" onClick={
                  (e) => {
                    e.preventDefault()
                    Swal.fire({
                      title: 'Are you sure?',
                      text: 'You will disable the password protection for this URL.',
                      icon: 'warning',
                      showCancelButton: true,
                      confirmButtonText: 'Yes, disable it!',
                      showLoaderOnConfirm: true,
                    }).then((result) => {
                      if (result.isConfirmed) {
                        fetch(import.meta.env.VITE_BACKEND_URL + '/url/' + data.backhalf + '/disable_password', {
                          method: 'PUT',
                          headers: {
                            'Authorization': `Bearer ${localStorage.getItem('user_token')}`
                          }
                        }).then(async (response) => {
                          if (response.ok) {
                            Swal.fire('Disabled!', 'Password has been disabled.', 'success').then(() => {
                              window.location.reload()
                            })
                          } else {
                            Swal.fire('Failed!', 'Failed to disable password.', 'error')
                          }
                        })
                      }
                    })
                  }
                }>
                  Disable Password
                </button>
                <button type="button" className="pure-button" onClick={
                  (e) => {
                    e.preventDefault()
                    // toggle class
                    e.currentTarget.classList.toggle('button-error')
                    setShowPassword(!show_password)
                  }
                }>
                  {
                  show_password ?
                  "Cancel" :
                  "Change Password"
                  }
                </button>
                </>
                }
              </div>
              {show_password &&
              <>
                <input type="password" name="password" ref={passwordref} placeholder="Password" style={{
                  width: '100%'
                }}/>
                <span className="pure-form-message">8-255 characters</span>
              </>
              }
              <button type="button" className="pure-button pure-button-primary" onClick={
                (e) => {
                  e.preventDefault()
                  Swal.fire({
                    title: 'Are you sure?',
                    text: 'You will update this URL.',
                    icon: 'warning',
                    showCancelButton: true,
                    confirmButtonText: 'Yes, update it!',
                    showLoaderOnConfirm: true,
                  }).then((result) => {
                    if (result.isConfirmed) {
                      fetch(import.meta.env.VITE_BACKEND_URL + '/url/' + data!.backhalf, {
                        method: 'PUT',
                        headers: {
                          'Content-Type': 'application/json',
                          'Authorization': `Bearer ${localStorage.getItem('user_token')}`
                        },
                        body: JSON.stringify({
                          original_url: original_url_ref.current?.value == data!.original_url ? undefined : original_url_ref.current?.value,
                          backhalf: backhalfref.current?.value == data!.backhalf ? undefined : backhalfref.current?.value,
                          password: passwordref.current?.value,
                        })
                      }).then(async (response) => {
                        if (response.ok) {
                          Swal.fire('Updated!', 'Your URL has been updated.', 'success').then(() => {
                            window.location.reload()
                          })
                        } else {
                          Swal.fire('Failed!', 'Failed to update URL.', 'error')
                        }
                      })
                    }
                  })
                }
              }>Save</button>
            </fieldset>
          </form>
          </>
          }
          <div className="info">
            Made by <a className="pure-anchor-style" href="https://www.naj.one">Thirafi Najwan</a> for personal use. Don't use it for important stuff as I don't guarantee the uptime of this service.
          </div>
        </div>
      </div>
    )
}