import React from 'react'

interface Props {
  className?: string
}
function TrashIcon(props: Props) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={props.className}
    >
      <path
        d="M3 6.00006H5H21"
        stroke="#F2F2F2"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M19 6.00006V20.0001C19 20.5305 18.7893 21.0392 18.4142 21.4143C18.0391 21.7893 17.5304 22.0001 17 22.0001H7C6.46957 22.0001 5.96086 21.7893 5.58579 21.4143C5.21071 21.0392 5 20.5305 5 20.0001V6.00006M8 6.00006V4.00006C8 3.46963 8.21071 2.96092 8.58579 2.58585C8.96086 2.21077 9.46957 2.00006 10 2.00006H14C14.5304 2.00006 15.0391 2.21077 15.4142 2.58585C15.7893 2.96092 16 3.46963 16 4.00006V6.00006"
        stroke="#F2F2F2"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M10 11.0001V17.0001"
        stroke="#F2F2F2"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M14 11.0001V17.0001"
        stroke="#F2F2F2"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
  )
}

export default TrashIcon
