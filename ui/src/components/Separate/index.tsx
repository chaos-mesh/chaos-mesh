import React from 'react'

interface SeparateProps {
  separator: React.ReactNode
}

const Separate: React.FC<SeparateProps> = ({ children, separator }) => {
  return (
    <>
      {React.Children.map(children, (child, index) => {
        return (
          <>
            {index > 0 && separator}
            {child}
          </>
        )
      })}
    </>
  )
}

export default Separate
