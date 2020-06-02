import { PayloadAction } from '@reduxjs/toolkit'

export interface NavigationBreadCrumbProps {
  name: string
  path?: string
}

export type NavigateAction = PayloadAction<string>
