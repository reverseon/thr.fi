import { useLocation } from "preact-iso";

export function Logout() {
    const location = useLocation();
    const to = location.query.to
    localStorage.removeItem('user_token')
    if (to) {
        location.route(to)
    } else {
        location.route("/")
    }
    return (
        <div>
            Please wait while we process your request...
        </div>
    )
}