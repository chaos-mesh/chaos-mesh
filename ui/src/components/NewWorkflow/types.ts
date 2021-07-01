import * as Yup from 'yup'

export const schemaBasic = Yup.object({
  name: Yup.string().trim().required('The task name is required'),
  deadline: Yup.string().trim().required('The deadline is required'),
})

export const schema = schemaBasic.shape({
  container: Yup.object({
    name: Yup.string().trim().required('The container name is required'),
    image: Yup.string().trim().required('The image is required'),
    command: Yup.array().of(Yup.string()),
  }),
  conditionalBranches: Yup.array()
    .of(
      Yup.object({
        target: Yup.string().trim().required('The target is required'),
        expression: Yup.string().trim().required('The expression is required'),
      })
    )
    .min(1)
    .required('The conditional branches should be defined'),
})
