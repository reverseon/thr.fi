import { ErrorBoundary, LocationProvider, Router } from "preact-iso";
import "./app.css";
import { Home } from "./components/home";
import { Logout } from "./components/logout";
import { NotFound } from "./components/notfound";
import { Set } from "./components/set";
import { LinkHandler } from "./components/linkhandler";
import { Manage } from "./components/manage";
import { ManageSpecific } from "./components/managespecific";
import { Success } from "./components/success";
import { Toaster } from "react-hot-toast";
import { createContext } from "preact";
import { useMemo, useState } from "preact/hooks";

// Context Declaration
export const SuccessShortenContext = createContext<{
  setShortenResult: (backhalf: string) => void
  shorten_result: string | undefined
}>({
  shorten_result: undefined,
  setShortenResult: () => {}
})

export function App() {
  // Context Initialization

  const [shorten_result, setShortenResult] = useState<string | undefined>(undefined)
  const shortencontext = useMemo(() => {
    return {
      shorten_result,
      setShortenResult
    }
  }, [shorten_result])
  return (
    <SuccessShortenContext.Provider value={shortencontext}>
      <LocationProvider>
        <Toaster />
        <ErrorBoundary>
          <Router>
            <Home path="/page/home" />
            <Success path="/page/result" />
            <Manage path="/page/manage" />
            <ManageSpecific path="/page/manage/:backhalf" />
            <Set path="/page/set" />
            <Logout path="/page/logout" />
            <LinkHandler path="/:backhalf?" />
            <NotFound default />
          </Router>
        </ErrorBoundary>
      </LocationProvider>
    </SuccessShortenContext.Provider>
  )
}
