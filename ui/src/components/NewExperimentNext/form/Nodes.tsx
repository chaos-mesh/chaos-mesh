/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { getIn, useFormikContext } from 'formik'

import { LabelField } from 'components/FormField'
import Space from 'components-mui/Space'
import T from 'components/T'

const Nodes = () => {
  const { errors, touched } = useFormikContext()

  return (
    <Space>
      <LabelField
        name={'spec.address'}
        label={T('physic.address')}
        helperText={
          getIn(touched, 'spec.address') && getIn(errors, 'spec.address')
            ? getIn(errors, 'spec.address')
            : T('physic.addressHelper')
        }
        error={getIn(errors, 'spec.address') && getIn(touched, 'spec.address') ? true : false}
      />
    </Space>
  )
}

export default Nodes
