import React from 'react'

interface Props {
  className?: string
}
function LockIcon(props: Props) {
  return (
    <svg
      className={props.className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M12.6667 7.33331H3.33333C2.59695 7.33331 2 7.93027 2 8.66665V13.3333C2 14.0697 2.59695 14.6666 3.33333 14.6666H12.6667C13.403 14.6666 14 14.0697 14 13.3333V8.66665C14 7.93027 13.403 7.33331 12.6667 7.33331Z"
        stroke="#3B9F74"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M4.66675 7.33331V4.66665C4.66675 3.78259 5.01794 2.93475 5.64306 2.30962C6.26818 1.6845 7.11603 1.33331 8.00008 1.33331C8.88414 1.33331 9.73198 1.6845 10.3571 2.30962C10.9822 2.93475 11.3334 3.78259 11.3334 4.66665V7.33331"
        stroke="#3B9F74"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
  )
}

export default LockIcon
