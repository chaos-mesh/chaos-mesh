import * as Yup from 'yup'

import { Kind } from './target'

const schema: Partial<Record<Kind, Record<string, Yup.ObjectSchema>>> = {
  PodChaos: {
    'container-kill': Yup.object({
      container_name: Yup.string().required('The container name is required'),
    }),
  },
  NetworkChaos: {
    partition: Yup.object({
      direction: Yup.string().required('The direction is required'),
    }),
    loss: Yup.object({
      loss: Yup.object({
        loss: Yup.string().required('The loss is required'),
      }),
    }),
    delay: Yup.object({
      delay: Yup.object({
        latency: Yup.string().required('The latency is required'),
      }),
    }),
    duplicate: Yup.object({
      duplicate: Yup.object({
        duplicate: Yup.string().required('The duplicate is required'),
      }),
    }),
    corrupt: Yup.object({
      corrupt: Yup.object({
        corrupt: Yup.string().required('The corrupt is required'),
      }),
    }),
    bandwidth: Yup.object({
      bandwidth: Yup.object({
        rate: Yup.string().required('The rate of bandwidth is required'),
      }),
    }),
  },
  IoChaos: {
    latency: Yup.object({
      delay: Yup.string().required('The delay is required'),
    }),
    fault: Yup.object({
      errno: Yup.number().min(0).required('The errno is required'),
    }),
    attrOverride: Yup.object({
      attr: Yup.array().of(Yup.string()).required('The attr is required'),
    }),
  },
  TimeChaos: {
    default: Yup.object({
      time_offset: Yup.string().required('The time offset is required'),
    }),
  },
}

export default schema
