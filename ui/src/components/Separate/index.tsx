import React from 'react'

interface Separate {
  separator: React.ReactNode
}

const Separate: React.FC<Separate> = ({ children, separator }) => {
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
