import BrowserOnly from '@docusaurus/BrowserOnly'
import CodeBlock from '../theme/CodeBlock'
import React from 'react'
import { usePluginData } from '@docusaurus/useGlobalData'

export const usePickVersion = () => {
  const locationHref = window.location.href
  const { versions } = usePluginData('docusaurus-plugin-content-docs')

  const latestStableVersion = versions.filter((d) => d.isLast)[0].name
  let activeVersion = versions.filter((d) => locationHref.includes(d.name)).map((d) => d.name)[0]

  if (locationHref.includes('/docs/next')) {
    activeVersion = 'latest'
  }

  return activeVersion || latestStableVersion
}

const PickVersion = ({ children, className }) => {
  const Result = ({ children }) => (
    <div style={{ marginBottom: '1.25rem' }}>
      <CodeBlock className={className}>{children}</CodeBlock>
    </div>
  )

  return (
    <BrowserOnly fallback={<Result>{children}</Result>}>
      {() => {
        const version = usePickVersion()
        const rendered = version === 'latest' ? children : children.replace('latest', 'v' + version)

        return <Result>{rendered}</Result>
      }}
    </BrowserOnly>
  )
}

export default PickVersion
