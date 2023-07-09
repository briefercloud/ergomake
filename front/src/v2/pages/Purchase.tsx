import { useMemo } from 'react'
import { useParams } from 'react-router-dom'

import StripePricingTable from '../../components/StripePricingTable'
import { orElse } from '../../hooks/useHTTPRequest'
import { useOwners } from '../../hooks/useOwners'
import { Profile } from '../../hooks/useProfile'
import Layout from '../components/Layout'

interface Props {
  profile: Profile
}

function Purchase({ profile }: Props) {
  const params = useParams<{ owner: string }>()
  const ownersRes = useOwners()
  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])
  const currentOwner = owners.find((o) => o.login === params.owner)

  const pages = [
    {
      name: 'Purchase',
      href: `/v2/gh/purchase`,
      label: 'Repositories',
    },
  ]

  return (
    <Layout profile={profile} pages={pages}>
      <div className="flex flex-col pt-8 px-8 space-y-12 rounded">
        <h1 className="text-gray-500 text-center font-bold text-4xl">
          Upgrade your plan for {currentOwner?.login}
        </h1>
        <StripePricingTable owner={currentOwner?.login ?? ''} />
      </div>
    </Layout>
  )
}

export default Purchase
