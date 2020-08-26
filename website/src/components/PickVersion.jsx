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
    activeVersion = latestVersion
  }

  return activeVersion || latestVersion
}

const PickVersion = ({ children, className }) => {
  return (
    <BrowserOnly fallback={<CodeBlock className={className}>{children}</CodeBlock>}>
      {() => {
        const version = usePickVersion()
        const rendered = children.replace('latest', version)

        return (
          <div style={{ marginBottom: '1.25rem' }}>
            <CodeBlock className={className}>{rendered}</CodeBlock>
          </div>
        )
      }}
    </BrowserOnly>
  )
}

export default PickVersion
