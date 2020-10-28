import { ReactComponent as ClockIcon } from '../images/time.svg'
import { ReactComponent as FileSystemIOIcon } from '../images/io.svg'
import { ReactComponent as LinuxKernelIcon } from '../images/kernel.svg'
import { ReactComponent as NetworkIcon } from '../images/network.svg'
import { ReactComponent as PodLifecycleIcon } from '../images/pod.svg'
import React from 'react'
import { ReactComponent as StressIcon } from '../images/stress.svg'
import { SvgIcon } from '@material-ui/core'
import T from 'components/T'

export type Kind = 'PodChaos' | 'NetworkChaos' | 'IoChaos' | 'KernelChaos' | 'TimeChaos' | 'StressChaos'
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
export interface Category {
  name: string
  key: string
  spec: Spec
}
interface Target {
  name: JSX.Element
  icon: JSX.Element
  categories?: Category[]
  spec?: Spec
}

const networkCommon: Spec = {
  direction: {
    field: 'select',
    items: ['', 'from', 'to', 'both'],
    label: 'Direction',
    value: '',
    helperText: 'Specifies the network direction',
  },
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
    label: 'Volume Path',
    value: '',
    helperText: 'The mount path of injected volume',
  },
  path: {
    field: 'text',
    label: 'Path',
    value: '',
    helperText: "Optional. The path of files for injecting. If it's empty, the action will inject into all files.",
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

const data: Record<Kind, Target> = {
  // Pod LifeCycle
  PodChaos: {
    name: T('newE.target.pod.title'),
    icon: (
      <SvgIcon fontSize="large">
        <PodLifecycleIcon />
      </SvgIcon>
    ),
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
        },
      },
      {
        name: 'Container Kill',
        key: 'container-kill',
        spec: {
          action: 'container-kill' as any,
          container_name: {
            field: 'text',
            label: 'Container Name',
            value: '',
            helperText: 'Fill the container name',
          },
        },
      },
    ],
  },
  // Network
  NetworkChaos: {
    name: T('newE.target.network.title'),
    icon: (
      <SvgIcon fontSize="large">
        <NetworkIcon />
      </SvgIcon>
    ),
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
            value: '',
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
  // File System IO
  IoChaos: {
    name: T('newE.target.io.title'),
    icon: (
      <SvgIcon fontSize="large">
        <FileSystemIOIcon />
      </SvgIcon>
    ),
    categories: [
      {
        name: 'Delay',
        key: 'delay',
        spec: {
          action: 'delay' as any,
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
  // Linux Kernel
  KernelChaos: {
    name: T('newE.target.kernel.title'),
    icon: (
      <SvgIcon fontSize="large">
        <LinuxKernelIcon />
      </SvgIcon>
    ),
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
  // Clock
  TimeChaos: {
    name: T('newE.target.time.title'),
    icon: (
      <SvgIcon fontSize="large">
        <ClockIcon />
      </SvgIcon>
    ),
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
  // Stress CPU/Memory
  StressChaos: {
    name: T('newE.target.stress.title'),
    icon: (
      <SvgIcon fontSize="large">
        <StressIcon />
      </SvgIcon>
    ),
    spec: {
      stressors: {
        cpu: {
          workers: 1,
          load: 0,
          options: [],
        },
        memory: {
          workers: 1,
          options: [],
        },
      },
      stressng_stressors: '',
      container_name: '',
    } as any,
  },
}

export type dataType = typeof data

export default data
