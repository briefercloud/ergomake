import { useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { orElse } from '../../hooks/useHTTPRequest'
import { useOwners } from '../../hooks/useOwners'
import { Profile } from '../../hooks/useProfile'
import Layout from '../components/Layout'

interface Props {
  profile: Profile
}

const Projects = ({ profile }: Props) => {
  const ownersRes = useOwners()
  const params = useParams<{ owner: string }>()

  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])

  const owner =
    owners.find((o) => o.login === params.owner)?.login ?? profile.username

  const pages = [
    { name: 'Repositories', href: `/v2/gh/${owner}`, label: 'Projects' },
  ]

  return (
    <Layout profile={profile} pages={pages}>
      <h1>Hello world!</h1>
    </Layout>
  )
}

export default Projects
