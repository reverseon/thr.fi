import { useLocation } from "preact-iso";

export function Set() {
    const location = useLocation();
    const at_token = location.query.at
    if (at_token) {
        localStorage.setItem('user_token', at_token)
    }
    location.route("/")
    return (
        <div>
            Please wait while we process your request...
        </div>
    )
}