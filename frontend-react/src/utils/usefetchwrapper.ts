export function useFetchWrapper<T>(
    setLoading: (arg0: boolean) => void,
    setFailed: (arg0: boolean) => void,
    setData: (arg0: T | undefined) => void,
    input: RequestInfo | URL, init?: RequestInit) {
    setLoading(true)
    setFailed(false)
    fetch(input, init).then(async (response) => {
        if (response.ok) {
            const data = await response.json()
            setData(data)
        } else {
            setFailed(true)
        }
        setLoading(false)
    }).catch((e) => {
        setFailed(true)
        setLoading(false)
    })
}