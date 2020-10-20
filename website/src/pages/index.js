import Layout from '@theme/Layout'
import PickVersion from '../components/PickVersion'
import React from 'react'
import clsx from 'clsx'
import features from '../data/features'
import styles from './index.module.css'
import useBaseUrl from '@docusaurus/useBaseUrl'
import useDocusaurusContext from '@docusaurus/useDocusaurusContext'
import whoIsUsing from '../data/whoIsUsing'

function Feature({ imgUrl, title, description, reverse }) {
  return (
    <div className={clsx('row', styles.feature, reverse ? styles.featureReverse : '')}>
      <div className="col col--6">
        <div className="text--center">
          {imgUrl && <img className={styles.featureImage} src={useBaseUrl(imgUrl)} alt={title} />}
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

      <div className={clsx('hero', styles.hero)}>
        <div className="container text--center">
          <h2 className="hero__subtitle">Start By One Line</h2>
          <div className={styles.installTextWrapper}>
            <PickVersion>curl -sSL https://mirrors.chaos-mesh.org/latest/install.sh | bash</PickVersion>
          </div>
        </div>
      </div>

      <main className={clsx('hero', styles.hero)}>
        <div className="container">
          <section className={styles.features}>
            <div className="container">
              {features.map((f, idx) => (
                <Feature key={idx} {...f} />
              ))}
            </div>
          </section>
        </div>
      </main>

      <div className={clsx('hero', styles.hero)}>
        <div className="container text--center">
          <h2 className="hero__subtitle">Who Are Using Chaos Mesh?</h2>
          <div className={styles.whiteboard}>
            <div className="row">
              {whoIsUsing.map((w) => (
                <div key={w.name} className="col col--3">
                  <div className={styles.logoWrapper}>
                    <img src={useBaseUrl(w.img)} alt={w.name} />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className={clsx('hero', styles.hero)}>
        <div className="container text--center">
          <h2 className="hero__subtitle">
            Chaos Mesh® is a <a href="https://cncf.io/">Cloud Native Computing Foundation</a> sandbox project
          </h2>
          <div className={clsx('cncf-logo', styles.cncfLogo)} />
        </div>
      </div>
    </Layout>
  )
}

export default Home
