import { Navigate, RouterProvider, createBrowserRouter } from 'react-router-dom'

import { RequireAuth, RequireNoAuth } from './components/RequireAuth'
import { AuthErrorProvider } from './hooks/useAuthError'
import { ThemeProvider } from './hooks/useTheme'
import Environment from './pages/Environment'
import Environments from './pages/Environments'
import Login from './pages/Login'
import NoInstallation from './pages/NoInstallation'
import Projects from './pages/Projects'
import PublicEnvironment from './pages/PublicEnvironment'
import Purchase from './pages/Purchase'

const router = createBrowserRouter([
  {
    path: '/login',
    element: (
      <RequireNoAuth>
        <Login />
      </RequireNoAuth>
    ),
  },
  {
    path: '/environments/:env',
    element: <PublicEnvironment />,
  },
  {
    path: '/gh',
    element: (
      <RequireAuth>
        {(profile) => <NoInstallation profile={profile} />}
      </RequireAuth>
    ),
  },
  {
    path: '/gh/:owner',
    element: (
      <RequireAuth>{(profile) => <Projects profile={profile} />}</RequireAuth>
    ),
  },
  {
    path: '/gh/:owner/repos/:repo',
    element: (
      <RequireAuth>
        {(profile) => <Environments profile={profile} />}
      </RequireAuth>
    ),
  },
  {
    path: '/gh/:owner/repos/:repo/envs/:env',
    element: (
      <RequireAuth>
        {(profile) => <Environment profile={profile} />}
      </RequireAuth>
    ),
  },
  {
    path: '/gh/:owner/purchase',
    element: (
      <RequireAuth>{(profile) => <Purchase profile={profile} />}</RequireAuth>
    ),
  },
  {
    path: '*',
    element: <Navigate to="/login" />,
  },
])

function App() {
  return (
    <ThemeProvider>
      <AuthErrorProvider>
        <RouterProvider router={router} />
      </AuthErrorProvider>
    </ThemeProvider>
  )
}

export default App
