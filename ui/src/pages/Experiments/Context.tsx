import React, { FC, useContext, useReducer } from 'react'
import { StepperAction, StepperState, StepperContextProps } from './types'

const reducer = (state: StepperState, action: StepperAction): StepperState => {
  switch (action.type) {
    case 'next':
      return { ...state, activeStep: state.activeStep + 1 }
    case 'back':
      return { ...state, activeStep: state.activeStep - 1 }
    case 'jump':
      return { ...state, activeStep: action.step }
    case 'reset':
      return { ...state, activeStep: 0 }
    default:
      return state
  }
}

const initialState = { activeStep: 0 }
const StepperContext = React.createContext<StepperContextProps | undefined>({ state: initialState, dispatch: () => {} })

const StepperProvider: FC = ({ children }) => {
  const [state, dispatch] = useReducer(reducer, initialState)

  return <StepperContext.Provider value={{ state, dispatch }}>{children}</StepperContext.Provider>
}

const useStepperContext = () => {
  const context = useContext(StepperContext)
  if (context === undefined) {
    throw new Error('useStepperContext must be used within a StepperProvider')
  }
  return context
}

export { StepperContext, StepperProvider, useStepperContext }
