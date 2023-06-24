import React from 'react'

interface Props {
  className?: string
}
function ExternalLinkIcon(props: Props) {
  return (
    <svg
      className={props.className}
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M18 13.0004V19.0004C18 19.5308 17.7893 20.0395 17.4142 20.4146C17.0391 20.7897 16.5304 21.0004 16 21.0004H5C4.46957 21.0004 3.96086 20.7897 3.58579 20.4146C3.21071 20.0395 3 19.5308 3 19.0004V8.00037C3 7.46993 3.21071 6.96123 3.58579 6.58615C3.96086 6.21108 4.46957 6.00037 5 6.00037H11"
        stroke="#3B9F74"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M15 3.00037H21V9.00037"
        stroke="#3B9F74"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M10 14.0004L21 3.00037"
        stroke="#3B9F74"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
  )
}

export default ExternalLinkIcon
