import BrowserOnly from '@docusaurus/BrowserOnly'
import CodeBlock from '../theme/CodeBlock'
import React from 'react'
import { usePluginData } from '@docusaurus/useGlobalData'

export const usePickVersion = () => {
  const locationHref = window.location.href
  const contentDocsData = usePluginData('docusaurus-plugin-content-docs')
  const latestVersion = contentDocsData.latestVersionName
  let activeVersion = contentDocsData.versions.filter((d) => locationHref.includes(d.path)).map((d) => d.name)[0]

  if (activeVersion === 'next') {
    activeVersion = 'latest'
  }

  return activeVersion || latestVersion
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
