import React, { useCallback, useState } from 'react'

interface Props extends React.HTMLAttributes<unknown> {
  startHidden?: boolean
}
function HidableSpan(props: Props) {
  const { startHidden, ...spanProps } = props
  const [hidden, setHidden] = useState(props.startHidden ?? false)

  const onToggle = useCallback(() => {
    setHidden((h) => !h)
  }, [setHidden])

  return (
    <span {...spanProps} onClick={onToggle}>
      {hidden ? '************' : props.children}
    </span>
  )
}

export default HidableSpan
