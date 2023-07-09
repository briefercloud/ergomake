import React from 'react'
import { Navigate, RouterProvider, createBrowserRouter } from 'react-router-dom'

import { RequireAuth, RequireNoAuth } from './components/RequireAuth'
import Environment from './pages/Environment'
import Environments from './pages/Environments'
import NoInstallation from './pages/NoInstallation'
import Projects from './pages/Projects'
import PublicEnvironment from './pages/PublicEnvironment'
import Purchase from './pages/Purchase'
import Registries from './pages/Registries'
import EnvironmentV2 from './v2/pages/Environment'
import EnvironmentsV2 from './v2/pages/Environments'
import Login from './v2/pages/Login'
import NoInstallationV2 from './v2/pages/NoInstallation'
import ProjectsV2 from './v2/pages/Projects'
import PurchaseV2 from './v2/pages/Purchase'

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
    path: '/v2/gh',
    element: (
      <RequireAuth>
        {(profile) => <NoInstallationV2 profile={profile} />}
      </RequireAuth>
    ),
  },
  {
    path: '/v2/gh/:owner',
    element: (
      <RequireAuth>{(profile) => <ProjectsV2 profile={profile} />}</RequireAuth>
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
    path: '/v2/gh/:owner/repos/:repo',
    element: (
      <RequireAuth>
        {(profile) => <EnvironmentsV2 profile={profile} />}
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
    path: '/v2/gh/:owner/repos/:repo/envs/:env',
    element: (
      <RequireAuth>
        {(profile) => <EnvironmentV2 profile={profile} />}
      </RequireAuth>
    ),
  },
  {
    path: '/gh/:owner/registries',
    element: (
      <RequireAuth>{(profile) => <Registries profile={profile} />}</RequireAuth>
    ),
  },
  {
    path: '/v2/gh/:owner/purchase',
    element: (
      <RequireAuth>{(profile) => <PurchaseV2 profile={profile} />}</RequireAuth>
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
  return <RouterProvider router={router} />
}

export default App
