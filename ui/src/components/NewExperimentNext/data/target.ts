import * as Yup from 'yup'

import { ExperimentKind } from 'components/NewExperiment/types'

export type Kind = ExperimentKind
type FieldType = 'text' | 'number' | 'select' | 'label' | 'autocomplete'
interface SpecField {
  field: FieldType
  items?: any[]
  isKV?: boolean
  label: string
  value: any
  helperText?: string
  inputProps?: Record<string, any>
}
export type Spec = Record<string, SpecField>
interface Category {
  name: string
  key: string
  spec: Spec
}
export interface Target {
  categories?: Category[]
  spec?: Spec
}

const networkCommon: Spec = {
  direction: {
    field: 'select',
    items: ['', 'from', 'to', 'both'],
    label: 'Direction',
    value: '',
    helperText: 'Specify the network direction',
  },
  external_targets: {
    field: 'label',
    label: 'External targets',
    value: [],
    helperText: 'Type string and end with a space to generate the network targets outside k8s',
  },
  target_scope: undefined as any,
}

const ioMethods = [
  '',
  'lookup',
  'forget',
  'getattr',
  'setattr',
  'readlink',
  'mknod',
  'mkdir',
  'unlink',
  'rmdir',
  'symlink',
  'rename',
  'link',
  'open',
  'read',
  'write',
  'flush',
  'release',
  'fsync',
  'opendir',
  'readdir',
  'releasedir',
  'fsyncdir',
  'statfs',
  'setxattr',
  'getxattr',
  'listxattr',
  'removexattr',
  'access',
  'create',
  'getlk',
  'setlk',
  'bmap',
]

const ioCommon: Spec = {
  volume_path: {
    field: 'text',
    label: 'Volume path',
    value: '',
    helperText: 'The mount path of injected volume',
  },
  path: {
    field: 'text',
    label: 'Path',
    value: '',
    helperText: "Optional. The path of files for injecting. If it's empty, the action will inject into all files.",
  },
  container_name: {
    field: 'text',
    label: 'Container name',
    value: '',
    helperText: 'Optional. The target container to inject in',
  },
  percent: {
    field: 'number',
    label: 'Percent',
    value: 100,
    helperText: 'The percentage of injection errors',
  },
  methods: {
    field: 'autocomplete',
    items: ioMethods,
    label: 'Methods',
    value: [],
    helperText: 'Optional. The IO methods for injecting IOChaos actions',
  },
}

const dnsCommon: Spec = {
  patterns: {
    field: 'label',
    label: 'Patterns',
    value: [],
    helperText: 'Specify the DNS patterns. For example, type google.com and then press space to add it.',
  },
  container_names: {
    field: 'label',
    label: 'Affected container names',
    value: [],
    helperText:
      "Optional. Type string and end with a space to generate the container names. If it's empty, all containers will be injected",
  },
}

const awsCommon: Spec = {
  secretName: {
    field: 'text',
    label: 'Secret name',
    value: '',
    helperText: 'Optional. The Kubernetes secret which includes AWS credentials',
  },
  awsRegion: {
    field: 'text',
    label: 'Region',
    value: '',
    helperText: 'The AWS region',
  },
  ec2Instance: {
    field: 'text',
    label: 'EC2 instance',
    value: '',
    helperText: 'The ID of a EC2 instance',
  },
}

const gcpCommon: Spec = {
  secretName: {
    field: 'text',
    label: 'Secret name',
    value: '',
    helperText: 'Optional. The Kubernetes secret which includes GCP credentials',
  },
  project: {
    field: 'text',
    label: 'Project',
    value: '',
    helperText: 'The name of a GCP project',
  },
  zone: {
    field: 'text',
    label: 'Zone',
    value: '',
    helperText: 'The zone of a GCP project',
  },
  instance: {
    field: 'text',
    label: 'Instance',
    value: '',
    helperText: 'The name of a VM instance',
  },
}

const data: Record<Kind, Target> = {
  // Pod Fault
  PodChaos: {
    categories: [
      {
        name: 'Pod Failure',
        key: 'pod-failure',
        spec: {
          action: 'pod-failure' as any,
        },
      },
      {
        name: 'Pod Kill',
        key: 'pod-kill',
        spec: {
          action: 'pod-kill' as any,
          grace_period: {
            field: 'number',
            label: 'Grace period',
            value: 0,
            helperText: 'Optional. Grace period represents the duration in seconds before the pod should be deleted',
          },
        },
      },
      {
        name: 'Container Kill',
        key: 'container-kill',
        spec: {
          action: 'container-kill' as any,
          container_names: {
            field: 'label',
            label: 'Container names',
            value: [],
            helperText: 'Type string and end with a space to generate the container names.',
          },
        },
      },
    ],
  },
  // Network Attack
  NetworkChaos: {
    categories: [
      {
        name: 'Partition',
        key: 'partition',
        spec: {
          action: 'partition' as any,
          ...networkCommon,
        },
      },
      {
        name: 'Loss',
        key: 'loss',
        spec: {
          action: 'loss' as any,
          loss: {
            field: 'text',
            label: 'Loss',
            value: '',
            helperText: 'The percentage of packet loss',
          },
          correlation: {
            field: 'text',
            label: 'Correlation',
            value: '0',
            helperText: 'The correlation of loss',
          },
          ...networkCommon,
        },
      },
      {
        name: 'Delay',
        key: 'delay',
        spec: {
          action: 'delay' as any,
          latency: {
            field: 'text',
            label: 'Latency',
            value: '',
            helperText: 'The latency of delay',
          },
          jitter: {
            field: 'text',
            label: 'Jitter',
            value: '',
            helperText: 'The jitter of delay',
          },
          correlation: {
            field: 'text',
            label: 'Correlation',
            value: '',
            helperText: 'The correlation of delay',
          },
          ...networkCommon,
        },
      },
      {
        name: 'Duplicate',
        key: 'duplicate',
        spec: {
          action: 'duplicate' as any,
          duplicate: {
            field: 'text',
            label: 'Duplicate',
            value: '',
            helperText: 'The percentage of packet duplication',
          },
          correlation: {
            field: 'text',
            label: 'Correlation',
            value: '',
            helperText: 'The correlation of duplicate',
          },
          ...networkCommon,
        },
      },
      {
        name: 'Corrupt',
        key: 'corrupt',
        spec: {
          action: 'corrupt' as any,
          corrupt: {
            field: 'text',
            label: 'Corrupt',
            value: '',
            helperText: 'The percentage of packet corruption',
          },
          correlation: {
            field: 'text',
            label: 'Correlation',
            value: '',
            helperText: 'The correlation of corrupt',
          },
          ...networkCommon,
        },
      },
      {
        name: 'Bandwidth',
        key: 'bandwidth',
        spec: {
          action: 'bandwidth' as any,
          rate: {
            field: 'text',
            label: 'Rate',
            value: '',
            helperText: 'The rate allows bps, kbps, mbps, gbps, tbps unit. For example, bps means bytes per second',
          },
          limit: {
            field: 'number',
            label: 'Limit',
            value: 0,
            helperText: 'The number of bytes that can be queued waiting for tokens to become available',
          },
          buffer: {
            field: 'number',
            label: 'Buffer',
            value: 0,
            helperText: 'The maximum amount of bytes that tokens can be available instantaneously',
          },
          minburst: {
            field: 'number',
            label: 'Min burst',
            value: 0,
            helperText: 'The size of the peakrate bucket',
          },
          peakrate: {
            field: 'number',
            label: 'Peak rate',
            value: 0,
            helperText: 'The maximum depletion rate of the bucket',
          },
          ...networkCommon,
        },
      },
    ],
  },
  // IO Injection
  IOChaos: {
    categories: [
      {
        name: 'Latency',
        key: 'latency',
        spec: {
          action: 'latency' as any,
          delay: {
            field: 'text',
            label: 'Delay',
            value: '',
            helperText:
              "The value of delay of I/O operations. If it's empty, the operator will generate a value for it randomly.",
            inputProps: { min: 0 },
          },
          ...ioCommon,
        },
      },
      {
        name: 'Fault',
        key: 'fault',
        spec: {
          action: 'fault' as any,
          errno: {
            field: 'number',
            label: 'Errno',
            value: 0,
            helperText: 'The error code returned by I/O operators. By default, it returns a random error code',
          },
          ...ioCommon,
        },
      },
      {
        name: 'AttrOverride',
        key: 'attrOverride',
        spec: {
          action: 'attrOverride' as any,
          attr: {
            field: 'label',
            isKV: true,
            label: 'Attr',
            value: [],
          },
          ...ioCommon,
        },
      },
    ],
  },
  // Stress Test
  StressChaos: {
    spec: {
      stressors: {
        cpu: {
          workers: 0,
          load: 0,
          options: [],
        },
        memory: {
          workers: 0,
          size: '',
          options: [],
        },
      },
      stressng_stressors: '',
      container_name: '',
    } as any,
  },
  // Kernel Fault
  KernelChaos: {
    spec: {
      fail_kern_request: {
        callchain: [],
        failtype: 0,
        headers: [],
        probability: 0,
        times: 0,
      },
    } as any,
  },
  // Clock Skew
  TimeChaos: {
    spec: {
      time_offset: {
        field: 'text',
        label: 'Offset',
        value: '',
        helperText: 'Fill the time offset',
      },
      clock_ids: {
        field: 'label',
        label: 'Clock ids',
        value: [],
        helperText:
          "Optional. Type string and end with a space to generate the clock ids. If it's empty, it will be set to ['CLOCK_REALTIME']",
      },
      container_names: {
        field: 'label',
        label: 'Affected container names',
        value: [],
        helperText:
          "Optional. Type string and end with a space to generate the container names. If it's empty, all containers will be injected",
      },
    },
  },
  // DNS Fault
  DNSChaos: {
    categories: [
      {
        name: 'Error',
        key: 'error',
        spec: {
          action: 'error' as any,
          ...dnsCommon,
        },
      },
      {
        name: 'Random',
        key: 'random',
        spec: {
          action: 'random' as any,
          ...dnsCommon,
        },
      },
    ],
  },
  // AWS
  AWSChaos: {
    categories: [
      {
        name: 'Stop EC2',
        key: 'ec2-stop',
        spec: {
          action: 'ec2-stop' as any,
          ...awsCommon,
        },
      },
      {
        name: 'Restart EC2',
        key: 'ec2-restart',
        spec: {
          action: 'ec2-restart' as any,
          ...awsCommon,
        },
      },
      {
        name: 'Detach Volumne',
        key: 'detach-volume',
        spec: {
          action: 'detach-volume' as any,
          ...awsCommon,
          deviceName: {
            field: 'text',
            label: 'Device name',
            value: '',
            helperText: 'The device name for the volume',
          },
          volumeID: {
            field: 'text',
            label: 'EBS volume',
            value: '',
            helperText: 'The ID of a EBS volume',
          },
        },
      },
    ],
  },
  // GCP
  GCPChaos: {
    categories: [
      {
        name: 'Stop node',
        key: 'node-stop',
        spec: {
          action: 'node-stop' as any,
          ...gcpCommon,
        },
      },
      {
        name: 'Reset node',
        key: 'node-reset',
        spec: {
          action: 'node-reset' as any,
          ...gcpCommon,
        },
      },
      {
        name: 'Loss disk',
        key: 'disk-loss',
        spec: {
          action: 'disk-loss' as any,
          ...gcpCommon,
          deviceNames: {
            field: 'label',
            label: 'Device names',
            value: [],
            helperText: 'Type string and end with a space to generate the device names',
          },
        },
      },
    ],
  },
}

const targetScopeSchema = Yup.object({
  namespaces: Yup.array().min(1, 'The namespace selectors is required'),
})

const patternsSchema = Yup.array().of(Yup.string()).required('The patterns is required')

const AwsChaosCommonSchema = Yup.object({
  awsRegion: Yup.string().required('The region is required'),
  ec2Instance: Yup.string().required('The ID of the EC2 instance is required'),
})

const GcpChaosCommonSchema = Yup.object({
  project: Yup.string().required('The project is required'),
  zone: Yup.string().required('The zone is required'),
  instance: Yup.string().required('The instance is required'),
})

export const schema: Partial<Record<Kind, Record<string, Yup.ObjectSchema>>> = {
  PodChaos: {
    'pod-kill': Yup.object({
      grace_period: Yup.number().min(0, 'Grace period must be non-negative integer'),
    }),
    'container-kill': Yup.object({
      container_names: Yup.array().of(Yup.string()).required('The container name is required'),
    }),
  },
  NetworkChaos: {
    partition: Yup.object({
      direction: Yup.string().required('The direction is required'),
      target_scope: targetScopeSchema,
    }),
    loss: Yup.object({
      loss: Yup.object({
        loss: Yup.string().required('The loss is required'),
      }),
      target_scope: targetScopeSchema,
    }),
    delay: Yup.object({
      delay: Yup.object({
        latency: Yup.string().required('The latency is required'),
      }),
      target_scope: targetScopeSchema,
    }),
    duplicate: Yup.object({
      duplicate: Yup.object({
        duplicate: Yup.string().required('The duplicate is required'),
      }),
      target_scope: targetScopeSchema,
    }),
    corrupt: Yup.object({
      corrupt: Yup.object({
        corrupt: Yup.string().required('The corrupt is required'),
      }),
      target_scope: targetScopeSchema,
    }),
    bandwidth: Yup.object({
      bandwidth: Yup.object({
        rate: Yup.string().required('The rate of bandwidth is required'),
      }),
      target_scope: targetScopeSchema,
    }),
  },
  IOChaos: {
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
  DNSChaos: {
    error: Yup.object({
      patterns: patternsSchema,
    }),
    random: Yup.object({
      patterns: patternsSchema,
    }),
  },
  AWSChaos: {
    'ec2-stop': AwsChaosCommonSchema,
    'ec2-restart': AwsChaosCommonSchema,
    'detach-volume': AwsChaosCommonSchema.shape({
      deviceName: Yup.string().required('The device name is required'),
      volumeID: Yup.string().required('The ID of the EBS volume is required'),
    }),
  },
  GCPChaos: {
    'node-stop': GcpChaosCommonSchema,
    'node-reset': GcpChaosCommonSchema,
    'disk-loss': GcpChaosCommonSchema.shape({
      deviceNames: Yup.array().of(Yup.string()).required('At least one device name is required'),
    }),
  },
}

export type dataType = typeof data

export default data
