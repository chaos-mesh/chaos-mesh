import React, { useEffect } from 'react'

import Layout from '@theme/Layout'

const useScript = (url) => {
  useEffect(() => {
    const script = document.createElement('script')

    script.src = url
    script.async = true

    document.body.appendChild(script)

    return () => {
      document.body.removeChild(script)
    }
  }, [url])
}

const InteractiveTutorial = () => {
  useScript('//katacoda.com/embed.js')

  return (
    <Layout
      title="Interactive Tutorial"
      description="This interactive tutorial will use Katacoda to cover an introduction to Chaos Mesh and guides the learner through two experiments that coincide with the Chaos Mesh documentation"
    >
      <div
        id="katacoda-scenario-1"
        data-katacoda-id="javajon/courses/kubernetes-chaos/chaos-mesh"
        data-katacoda-color="172d72"
        style={{ height: 'calc(100vh - 60px)' }}
      ></div>
    </Layout>
  )
}

export default InteractiveTutorial
