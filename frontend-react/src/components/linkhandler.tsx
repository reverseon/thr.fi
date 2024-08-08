import { useLocation, useRoute } from "preact-iso";
import { useEffect, useRef, useState } from "preact/hooks";
import { NotFound } from "./notfound";
import toast from "react-hot-toast";

export function LinkHandler() {
    const backhalf = useRoute().params.backhalf
    const location = useLocation()
    const [status, setStatus] = useState<0 | 1 | 2 | 3 | 4>(0)
    const passwordref = useRef<HTMLInputElement>(null)
    const [show_password, setShowPassword] = useState<boolean>(true)
    // Status 0: Loading
    // Status 1: Success
    // Status 2: Need Password
    // Status 3: Not Found
    // Status 4: Error
    useEffect(() => {
        if (backhalf == undefined || backhalf == "" || backhalf == null) {
            location.route('/page/home')
        } else {
            fetch(`${import.meta.env.VITE_BACKEND_URL}/url/${backhalf}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                }
            }).then((res) => {
                if (res.ok) {
                    setStatus(1)
                    res.json().then((data) => {
                        window.location.href = data.original_url
                    })
                } else if (res.status == 403) {
                    setStatus(2)
                } else if (res.status == 404) {
                    setStatus(3)
                } else {
                    setStatus(4)
                }
            })
        }
    }, [])
    return (
        status == 0 || status == 1 ? 
        <div>
            Redirecting you...
        </div> :
        status == 2 ?
        <div className="container">
            <div className="card">
                <h2 className="pure-u-1">
                    This link is password protected
                </h2>
                <form className="pure-form pure-form-stacked" onSubmit={
                    (e) => {
                        e.preventDefault()
                    }
                }>
                    <fieldset>
                        { show_password &&
                        <input type="password" name="password" ref={passwordref} placeholder="Password" style={{
                        width: '100%'
                        }}/>
                        }
                    </fieldset>
                </form>
                { show_password ? <button type="button" className="pure-button pure-button-primary" onClick={
                () => {
                    fetch(`${import.meta.env.VITE_BACKEND_URL}/url/${backhalf}/unlock`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            password: passwordref.current?.value
                        })
                    }).then((res) => {
                        if (res.ok) {
                            toast.success('Link unlocked. Redirecting you...')
                            setShowPassword(false)
                            // timeout 20ms to show the toast
                            setTimeout(() => {
                                res.json().then((data) => {
                                    window.location.href = data.original_url
                                })
                            }, 20)
                        } else if (res.status == 500) {
                            toast.error('Failed to unlock the link due to server error.')
                        } else {
                            toast.error('Failed to unlock the link. Please check your password.')
                        }
                    }).catch(() => {
                        toast.error('Failed to unlock the link due to network error.')
                    })
                }}
                style={{
                    marginBottom: '1em'
                }}
                >Unlock</button>
                : <p>
                    Redirecting you...
                </p> }
                <div className="info">
                    Made by <a className="pure-anchor-style" href="https://www.naj.one">Thirafi Najwan</a> for personal use. Don't use it for important stuff as I don't guarantee the uptime of this service.
                </div>
            </div>
        </div> :
        status == 3 ?
        <NotFound /> : 
        <div>
            <h1>Something wen't wrong</h1>
        </div>
    )
}