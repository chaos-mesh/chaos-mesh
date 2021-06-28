import * as Yup from 'yup'

import { Form, Formik } from 'formik'
import { IconButton, Typography } from '@material-ui/core'
import { LabelField, Submit, TextField } from 'components/FormField'
import { Template, TemplateCustom } from 'slices/workflows'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import RemoveCircleIcon from '@material-ui/icons/RemoveCircle'
import Space from 'components-mui/Space'
import T from 'components/T'

const schema = Yup.object({
  name: Yup.string().trim().required('The task name is bequired'),
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

export interface CustomValues extends TemplateCustom {
  name: string
}

interface CustomProps {
  initialValues?: CustomValues
  submit: (template: Template) => void
}

const Custom: React.FC<CustomProps> = ({ initialValues, submit }) => {
  const onSubmit = ({ name: _name, container, conditionalBranches }: CustomValues) => {
    const { name, ...rest } = schema.cast({ name: _name, container, conditionalBranches }) as any

    submit({
      type: 'custom',
      name,
      custom: rest,
    })
  }

  return (
    <Paper>
      <Space>
        <PaperTop title={T('newW.customTitle')} />
        <Formik
          initialValues={
            initialValues || {
              name: '',
              container: {
                name: '',
                image: '',
                command: [],
              },
              conditionalBranches: [
                {
                  target: '',
                  expression: '',
                },
              ],
            }
          }
          validationSchema={schema}
          onSubmit={onSubmit}
        >
          {({ values: { conditionalBranches }, setFieldValue, errors, touched }) => {
            const addBranch = () =>
              setFieldValue(
                'conditionalBranches',
                conditionalBranches.concat([
                  {
                    target: '',
                    expression: '',
                  },
                ])
              )

            const removeBranch = (index: number) => () => {
              setFieldValue(
                'conditionalBranches',
                conditionalBranches.filter((_: any, i: number) => index !== i)
              )
            }

            return (
              <Form>
                <Space>
                  <TextField
                    fast
                    name="name"
                    label={T('common.name')}
                    helperText={errors.name && touched.name ? errors.name : T('newW.node.nameHelper')}
                    error={errors.name && touched.name ? true : false}
                  />
                  <Typography variant="body2">{T('newW.node.container.title')}</Typography>
                  <TextField
                    fast
                    name="container.name"
                    label={T('common.name')}
                    helperText={
                      errors.container?.name && touched.container?.name
                        ? errors.container.name
                        : T('newW.node.container.nameHelper')
                    }
                    error={errors.container?.name && touched.container?.name ? true : false}
                  />
                  <TextField
                    fast
                    name="container.image"
                    label={T('newW.node.container.image')}
                    helperText={
                      errors.container?.image && touched.container?.image
                        ? errors.container.image
                        : T('newW.node.container.imageHelper')
                    }
                    error={errors.container?.image && touched.container?.image ? true : false}
                  />
                  <LabelField
                    name="container.command"
                    label={T('newW.node.container.command')}
                    helperText={T('newW.node.container.commandHelper')}
                  />
                  <Typography variant="body2">{T('newW.node.conditionalBranches.title')}</Typography>
                  {conditionalBranches.length > 0 &&
                    conditionalBranches.map((_, i) => (
                      <Space key={i} direction="row" alignItems="center">
                        <Typography component="div" variant="button">
                          if
                        </Typography>
                        <TextField
                          name={`conditionalBranches[${i}].expression`}
                          label={T('newW.node.conditionalBranches.expression')}
                        />
                        <Typography component="div" variant="button">
                          then
                        </Typography>
                        <TextField
                          name={`conditionalBranches[${i}].target`}
                          label={T('newW.node.conditionalBranches.target')}
                        />
                        {i !== conditionalBranches.length - 1 && (
                          <IconButton color="secondary" size="small" onClick={removeBranch(i)}>
                            <RemoveCircleIcon />
                          </IconButton>
                        )}
                        {i === conditionalBranches.length - 1 && (
                          <IconButton color="primary" size="small" onClick={addBranch}>
                            <AddCircleIcon />
                          </IconButton>
                        )}
                      </Space>
                    ))}
                </Space>
                <Submit />
              </Form>
            )
          }}
        </Formik>
      </Space>
    </Paper>
  )
}

export default Custom
