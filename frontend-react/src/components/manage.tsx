import { useLocation } from "preact-iso"
import { useEffect, useState } from "preact/hooks"
import Swal from "sweetalert2"
import { useFetchWrapper } from "../utils/usefetchwrapper"

export function Manage() {
  const location = useLocation()

  const [current_page, setCurrentPage] = useState<number>(1)

  // Fetch Declaration
  const [user_info_loading, setUserInfoLoading] = useState<boolean>(true)
  const [user_info_failed, setUserInfoFailed] = useState<boolean>(false)
  const [user_info, setUserInfo] = useState<{
    email: string,
    id: string
  } | undefined>(undefined)

  const [urls_loading, setUrlsLoading] = useState<boolean>(true)
  const [urls_failed, setUrlsFailed] = useState<boolean>(false)
  const [urls, setUrls] = useState<{
    total: number,
    urls: {
      Backhalf: string,
      Password_protected: boolean
    }[]
  } | undefined>(undefined)

  // Logout URL Construction
  const logout_callback = `${import.meta.env.VITE_APP_URL}/page/logout?to=${encodeURIComponent("/")}`
  const logout_url = `${import.meta.env.VITE_CENTRAL_AUTH_URL}/?ras=${encodeURIComponent(logout_callback)}`

  // change page handler
  useEffect(() => {
    async function changePage() {
      const user_token = localStorage.getItem('user_token')
      useFetchWrapper<{
        total: number,
        urls: {
          Backhalf: string,
          Password_protected: boolean
        }[]
      }>(
        setUrlsLoading,
        setUrlsFailed,
        setUrls,
        import.meta.env.VITE_BACKEND_URL + `/url/user?per_page=5&page=${current_page}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${user_token}`
        }
      })
    }
    changePage()
  }, [current_page])

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
      total: number,
      urls: {
        Backhalf: string,
        Password_protected: boolean
      }[]
    }>(
      setUrlsLoading,
      setUrlsFailed,
      setUrls,
      import.meta.env.VITE_BACKEND_URL + `/url/user?per_page=5&page=${current_page}`, {
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
        <h2 className="pure-u-1">Manage URL</h2>
        {
        user_info_loading ?
        <p>Loading...</p> :
        user_info_failed ?
        <p>
          You don't have access. <span className="pure-anchor-style" onClick={
            (e) => {
              e.preventDefault()
              location.route('/page/home')
            }
          }>Back to home</span>
        </p> :
        <>
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
                  }>Logout</a> | <span type="button" className="pure-anchor-style" onClick={
                    (e) => {
                      e.preventDefault()
                      location.route('/page/home')
                    }
                  }>Back to home</span> 
                </p>
                </>
              }
            </legend>
          </fieldset>
        </form>
        { urls_loading ?
          <p>Loading...</p> :
          urls_failed ?
          <p>Failed to load URLs.</p> :
          urls?.total === 0 ?
          <p>No URL found.</p> :
          <>
          <div className="info" style={{
            marginBottom: '1em'
          }}>
            {
            `Displaying ${((current_page-1) * 5)+1} - ${Math.min(current_page*5, urls!.total)} of ${urls!.total} URLs`
            }
          </div>
          <div style={{
            marginBottom: '1em',
            width: '100%',
            overflowX: 'auto',
            fontSize: '0.9rem'
          }}>
            <table class="pure-table pure-table-horizontal last-fit" style={{
              width: '100%'
            }}>
              <thead>
                  <tr>
                      <th>Backhalf</th>
                      <th>Action</th>
                  </tr>
              </thead>
              <tbody>
                {urls === undefined ?
                  <tr>
                    <td colSpan={2}>Loading...</td>
                  </tr> :
                urls.urls.map((data) => {
                    return (
                      <tr>
                        <td>
                          {data.Backhalf}
                          {
                            data.Password_protected && 
                            <i class="bi bi-lock-fill" style={{
                              marginLeft: '0.5em'
                            }}></i>
                          }
                        </td>
                        <td>
                          <button type="button" className="pure-button pure-button-primary" style={{
                            marginRight: '1em'
                          }}
                          onClick={
                            (e) => {
                              e.preventDefault()
                              location.route(`/page/manage/${data.Backhalf}`)
                            }
                          }>Edit</button>
                          <button type="button" className="pure-button" onClick={
                            (e) => {
                              e.preventDefault()
                              Swal.fire({
                                title: 'Are you sure?',
                                text: 'You will not be able to recover this URL!',
                                icon: 'warning',
                                showCancelButton: true,
                                confirmButtonText: 'Yes, delete it!',
                                cancelButtonText: 'No, keep it'
                              }).then(async (result) => {
                                if (result.isConfirmed) {
                                  const user_token = localStorage.getItem('user_token')
                                  fetch(import.meta.env.VITE_BACKEND_URL + '/url/' + data.Backhalf, {
                                    method: 'DELETE',
                                    headers: {
                                      'Authorization': `Bearer ${user_token}`
                                    }
                                  }).then(async (response) => {
                                    if (response.ok) {
                                      Swal.fire('Deleted!', 'Your URL has been deleted.', 'success').then(() => {
                                          window.location.reload()
                                      })
                                    } else {
                                      Swal.fire('Failed!', 'Failed to delete URL.', 'error')
                                    }
                                  })
                                }
                              })
                            }
                          }>Delete</button>
                        </td>
                      </tr>
                    )
                  })
                }
              </tbody>
            </table>
          </div>
          <div style={{
            marginBottom: '1em',
            width: '100%',
            display: 'flex',
            justifyContent: 'space-between'
          }}>
            <div className="pure-button pure-button-primary"
            style={{
              visibility: current_page === 1 ? 'hidden' : 'visible'
            }}
            onClick={
              (e) => {
                e.preventDefault()
                setCurrentPage(current_page - 1)
              }
            }
            >Previous</div>
            <div className="pure-button pure-button-primary" style={{
              visibility: current_page * 5 >= urls!.total ? 'hidden' : 'visible'
            }}
            onClick={
              (e) => {
                e.preventDefault()
                setCurrentPage(current_page + 1)
              }
            }
            >Next</div>
          </div>
          </>
        }
        </>
        }
        <div className="info">
          Made by <a className="pure-anchor-style" href="https://www.naj.one">Thirafi Najwan</a> for personal use. Don't use it for important stuff as I don't guarantee the uptime of this service.
        </div>
      </div>
    </div>
  )
}