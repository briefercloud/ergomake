import React from 'react'

import Background from '../components/Background'
import Navbar from '../components/Navbar'
import StripePricingTable from '../components/StripePricingTable'
import { Profile } from '../hooks/useProfile'

interface Props {
  profile: Profile
  owner: string
}
function Purchase(props: Props) {
  return (
    <Background>
      <Navbar profile={props.profile} currentOwner={props.owner} />
      <div className="flex flex-col p-20 space-y-12 rounded">
        <h1 className="text-white text-center font-bold text-4xl">
          Upgrade your plan for {props.owner}
        </h1>
        <StripePricingTable owner={props.owner} />
      </div>
    </Background>
  )
}

export default Purchase
