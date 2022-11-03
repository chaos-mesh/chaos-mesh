/**
 * This file was auto-generated by @ui/openapi.
 * Do not make direct changes to the file.
 */

export const actions = ['netem', 'delay', 'loss', 'duplicate', 'corrupt', 'partition', 'bandwidth'],
  data = [
    {
      field: 'ref',
      label: 'bandwidth',
      children: [
        {
          field: 'number',
          label: 'buffer',
          value: 0,
          helperText: 'Buffer is the maximum amount of bytes that tokens can be available for instantaneously.',
        },
        {
          field: 'number',
          label: 'limit',
          value: 0,
          helperText: 'Limit is the number of bytes that can be queued waiting for tokens to become available.',
        },
        {
          field: 'number',
          label: 'minburst',
          value: 0,
          helperText:
            'Optional. Minburst specifies the size of the peakrate bucket. For perfect accuracy, should be set to the MTU of the interface.  If a peakrate is needed, but some burstiness is acceptable, this size can be raised. A 3000 byte minburst allows around 3mbit/s of peakrate, given 1000 byte packets.',
        },
        {
          field: 'number',
          label: 'peakrate',
          value: 0,
          helperText:
            'Optional. Peakrate is the maximum depletion rate of the bucket. The peakrate does not need to be set, it is only necessary if perfect millisecond timescale shaping is required.',
        },
        {
          field: 'text',
          label: 'rate',
          value: '',
          helperText: 'Rate is the speed knob. Allows bps, kbps, mbps, gbps, tbps unit. bps means bytes per second.',
        },
      ],
      when: "action=='bandwidth'",
    },
    {
      field: 'ref',
      label: 'corrupt',
      children: [
        {
          field: 'text',
          label: 'correlation',
          value: '',
          helperText: 'Optional.',
        },
        {
          field: 'text',
          label: 'corrupt',
          value: '',
          helperText: '',
        },
      ],
      when: "action=='corrupt'",
    },
    {
      field: 'ref',
      label: 'delay',
      children: [
        {
          field: 'text',
          label: 'correlation',
          value: '',
          helperText: 'Optional.',
        },
        {
          field: 'text',
          label: 'jitter',
          value: '',
          helperText: 'Optional.',
        },
        {
          field: 'text',
          label: 'latency',
          value: '',
          helperText: '',
        },
        {
          field: 'ref',
          label: 'reorder',
          children: [
            {
              field: 'text',
              label: 'correlation',
              value: '',
              helperText: 'Optional.',
            },
            {
              field: 'number',
              label: 'gap',
              value: 0,
              helperText: '',
            },
            {
              field: 'text',
              label: 'reorder',
              value: '',
              helperText: '',
            },
          ],
        },
      ],
      when: "action=='delay'",
    },
    {
      field: 'text',
      label: 'device',
      value: '',
      helperText: 'Optional. Device represents the network device to be affected.',
    },
    {
      field: 'text',
      label: 'direction',
      value: '',
      helperText: 'Optional. Direction represents the direction, this applies on netem and network partition action',
    },
    {
      field: 'ref',
      label: 'duplicate',
      children: [
        {
          field: 'text',
          label: 'correlation',
          value: '',
          helperText: 'Optional.',
        },
        {
          field: 'text',
          label: 'duplicate',
          value: '',
          helperText: '',
        },
      ],
      when: "action=='duplicate'",
    },
    {
      field: 'label',
      label: 'externalTargets',
      value: [],
      helperText: 'Optional. ExternalTargets represents network targets outside k8s',
    },
    {
      field: 'ref',
      label: 'loss',
      children: [
        {
          field: 'text',
          label: 'correlation',
          value: '',
          helperText: 'Optional.',
        },
        {
          field: 'text',
          label: 'loss',
          value: '',
          helperText: '',
        },
      ],
      when: "action=='loss'",
    },
    {
      field: 'ref',
      label: 'target',
      children: [
        {
          field: 'text',
          label: 'mode',
          value: '',
          helperText:
            'Mode defines the mode to run chaos action. Supported mode: one / all / fixed / fixed-percent / random-max-percent',
        },
        {
          field: 'ref',
          label: 'selector',
          children: [
            {
              field: 'text-text',
              label: 'annotationSelectors',
              value: {},
              helperText:
                'Optional. Map of string keys and values that can be used to select objects. A selector based on annotations.',
            },
            {
              field: 'text-text',
              label: 'fieldSelectors',
              value: {},
              helperText:
                'Optional. Map of string keys and values that can be used to select objects. A selector based on fields.',
            },
            {
              field: 'text-text',
              label: 'labelSelectors',
              value: {},
              helperText:
                'Optional. Map of string keys and values that can be used to select objects. A selector based on labels.',
            },
            {
              field: 'label',
              label: 'namespaces',
              value: [],
              helperText: 'Optional. Namespaces is a set of namespace to which objects belong.',
            },
            {
              field: 'text-text',
              label: 'nodeSelectors',
              value: {},
              helperText:
                "Optional. Map of string keys and values that can be used to select nodes. Selector which must match a node's labels, and objects must belong to these selected nodes.",
            },
            {
              field: 'label',
              label: 'nodes',
              value: [],
              helperText: 'Optional. Nodes is a set of node name and objects must belong to these nodes.',
            },
            {
              field: 'label',
              label: 'podPhaseSelectors',
              value: [],
              helperText:
                'Optional. PodPhaseSelectors is a set of condition of a pod at the current time. supported value: Pending / Running / Succeeded / Failed / Unknown',
            },
            {
              field: 'text-label',
              label: 'pods',
              value: {},
              helperText:
                'Optional. Pods is a map of string keys and a set values that used to select pods. The key defines the namespace which pods belong, and the each values is a set of pod names.',
            },
          ],
        },
        {
          field: 'text',
          label: 'value',
          value: '',
          helperText:
            'Optional. Value is required when the mode is set to `FixedMode` / `FixedPercentMode` / `RandomMaxPercentMode`. If `FixedMode`, provide an integer of pods to do chaos action. If `FixedPercentMode`, provide a number from 0-100 to specify the percent of pods the server can do chaos action. IF `RandomMaxPercentMode`,  provide a number from 0-100 to specify the max percent of pods to do chaos action',
        },
      ],
    },
    {
      field: 'text',
      label: 'targetDevice',
      value: '',
      helperText: 'Optional. TargetDevice represents the network device to be affected in target scope.',
    },
  ]
