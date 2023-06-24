import React from 'react'

interface Props {
  className?: string
}
function CopyIcon(props: Props) {
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
        d="M20 9.00037H11C9.89543 9.00037 9 9.8958 9 11.0004V20.0004C9 21.1049 9.89543 22.0004 11 22.0004H20C21.1046 22.0004 22 21.1049 22 20.0004V11.0004C22 9.8958 21.1046 9.00037 20 9.00037Z"
        stroke="#3B9F74"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M5 15.0004H4C3.46957 15.0004 2.96086 14.7897 2.58579 14.4146C2.21071 14.0395 2 13.5308 2 13.0004V4.00037C2 3.46993 2.21071 2.96123 2.58579 2.58615C2.96086 2.21108 3.46957 2.00037 4 2.00037H13C13.5304 2.00037 14.0391 2.21108 14.4142 2.58615C14.7893 2.96123 15 3.46993 15 4.00037V5.00037"
        stroke="#3B9F74"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
  )
}

export default CopyIcon
