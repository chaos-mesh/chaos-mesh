import type { AutoFormProps } from 'components/AutoForm'

export enum ElementTypes {
  Kubernetes = 'Kubernetes',
  PhysicalNodes = 'PhysicalNodes',
}

export type ElementDragData = Omit<AutoFormProps, 'formikProps'>
