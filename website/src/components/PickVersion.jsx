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
  const version = usePickVersion()
  const rendered = children.replace('version', version)

  return (
    <div style={{ marginBottom: '1.25rem' }}>
      <CodeBlock className={className}>{rendered}</CodeBlock>
    </div>
  )
}

export default PickVersion
