import Layout from '@theme/Layout'
import PickVersion from '../components/PickVersion'
import React from 'react'
import clsx from 'clsx'
import styles from './styles.module.css'
import useBaseUrl from '@docusaurus/useBaseUrl'
import useDocusaurusContext from '@docusaurus/useDocusaurusContext'

const features = [
  {
    title: <>Easy to Use</>,
    imageUrl: 'img/undraw_server_down_s4lk.svg',
    description: (
      <>
        <p>
          No special dependencies, Chaos Mesh can be easily deployed directly on Kubernetes clusters, including{' '}
          <a href="https://github.com/kubernetes/minikube">Minikube</a> and{' '}
          <a href="https://kind.sigs.k8s.io/docs/user/quick-start/">Kind</a>.
        </p>
        <ul>
          <li>Require no modification to the deployment logic of the system under test (SUT)</li>
          <li>Easily orchestrate fault injection behaviors in chaos experiments</li>
          <li>Hide underlying implementation details so that users can focus on orchestrating the chaos experiments</li>
        </ul>
      </>
    ),
  },
  {
    title: <>Design for Kubernetes</>,
    imageUrl: 'img/logos/kubernetes.svg',
    description: (
      <>
        <p>
          Chaos Mesh uses{' '}
          <a
            href="https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/"
            target="_blank"
          >
            CustomResourceDefinitions
          </a>{' '}
          (CRD) to define chaos objects.
        </p>
        <p>
          In the Kubernetes realm, CRD is a mature solution for implementing custom resources, with abundant
          implementation cases and toolsets available. Using CRD makes Chaos Mesh naturally integrate with the
          Kubernetes ecosystem.
        </p>
      </>
    ),
    reverse: true,
  },
]

function Feature({ imageUrl, title, description, reverse }) {
  return (
    <div className={clsx('row', styles.feature, reverse ? styles.featureReverse : '')}>
      <div className="col col--6">
        <div className="text--center">
          {imageUrl && <img className={styles.featureImage} src={useBaseUrl(imageUrl)} alt={title} />}
        </div>
      </div>
      <div className={clsx('col col--6', styles.featureDesc)}>
        <div>
          <h3>{title}</h3>
          <div>{description}</div>
        </div>
      </div>
    </div>
  )
}

function Home() {
  const context = useDocusaurusContext()
  const { siteConfig } = context

  return (
    <Layout title={siteConfig.tagline} description={siteConfig.tagline}>
      <header className={clsx('hero', styles.hero)}>
        <div className="container text--center">
          <div className={styles.heroLogoWrapper}>
            <img className={styles.heroLogo} src={useBaseUrl('img/logos/logo-mini.svg')} alt="Chaos Mesh Logo" />
          </div>
          <h1 className={clsx('hero__title', styles.heroTitle)}>{siteConfig.title}</h1>
          <p className="hero__subtitle">{siteConfig.tagline}</p>
        </div>
      </header>

      <div className={clsx('text--center', styles.install)}>
        <h2>Start By One Line</h2>
        <div className={styles.installTextWrapper}>
          <PickVersion>curl -sSL https://mirrors.chaos-mesh.org/latest/install.sh | bash</PickVersion>
        </div>
      </div>

      <main className={styles.main}>
        {features && features.length > 0 && (
          <section className={styles.features}>
            <div className="container">
              {features.map((f, idx) => (
                <Feature key={idx} {...f} />
              ))}
            </div>
          </section>
        )}

        <div className="hero">
          <div className="container text--center">
            <h2 className="hero__subtitle">
              Chaos Mesh® is a <a href="https://cncf.io/">Cloud Native Computing Foundation</a> sandbox project
            </h2>
            <div className={clsx('cncf-logo', styles.cncfLogo)} />
          </div>
        </div>
      </main>
    </Layout>
  )
}

export default Home
