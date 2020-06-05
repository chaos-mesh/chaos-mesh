import React, { createContext, useContext, useReducer } from 'react'
import { StepperAction, StepperContextProps, StepperState, StepperType } from './types'

import { createAction } from '@reduxjs/toolkit'

const initialState = { activeStep: 0 }

export const next = createAction(StepperType.NEXT)
export const back = createAction(StepperType.BACK)
export const jump = createAction<number, StepperType>(StepperType.JUMP)
export const reset = createAction(StepperType.RESET)

const reducer = (state: StepperState, action: StepperAction): StepperState => {
  switch (action.type) {
    case StepperType.NEXT:
      return { ...state, activeStep: state.activeStep + 1 }
    case StepperType.BACK:
      return { ...state, activeStep: state.activeStep - 1 }
    case StepperType.JUMP:
      return { ...state, activeStep: action.payload! }
    case StepperType.RESET:
      return { ...state, activeStep: 0 }
    default:
      return state
  }
}

const StepperContext = createContext<StepperContextProps | undefined>({ state: initialState, dispatch: () => {} })

const StepperProvider: React.FC = ({ children }) => {
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

export { StepperProvider, useStepperContext }
