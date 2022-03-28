import { ComponentMeta, ComponentStory } from '@storybook/react'
import React, { useState } from 'react'

import KVPairList from '../../esm/KVPairList'
import type { KVPairListProps } from '../../esm/KVPairList'

export default {
  title: 'Form/KVPairList',
  component: KVPairList,
  argTypes: {
    onChange: {
      action: 'onChange',
    },
  },
} as ComponentMeta<typeof KVPairList>

const Template: ComponentStory<typeof KVPairList> = ({ ...props }: KVPairListProps<string, string>) => {
  return <KVPairList {...props} />
}

export const Default = Template.bind({})
Default.args = {
  name: 'spec.abort',
  label: 'HTTP Headers',
  helperText: 'Set HTTP Headers',
  helperTextForKey: '',
  helperTextForValue: '',
  disabled: false,
  error: false,
  data: [{ key: 'key-1', value: 'value-1' }],
}
