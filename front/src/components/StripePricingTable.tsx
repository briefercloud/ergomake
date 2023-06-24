import React, { useEffect } from 'react'

interface Props {
  owner: string
}
const StripePricingTable = (props: Props) => {
  useEffect(() => {
    const script = document.createElement('script')
    script.src = 'https://js.stripe.com/v3/pricing-table.js'
    script.async = true

    document.body.appendChild(script)

    return () => {
      document.body.removeChild(script)
    }
  }, [])

  const pricingTableId = process.env.REACT_APP_STRIPE_PRICING_TABLE_ID
  const publishableKey = process.env.REACT_APP_STRIPE_PUBLISHABLE_KEY

  return React.createElement('stripe-pricing-table', {
    'pricing-table-id': pricingTableId,
    'publishable-key': publishableKey,
    'client-reference-id': props.owner,
  })
}

export default StripePricingTable
