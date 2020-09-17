<!-- markdownlint-disable-file MD033 -->
<!-- markdownlint-disable-file MD041 -->

<p align="center">
  <img src="../static/logo.svg" width="256" alt="Chaos Mesh Logo" />
</p>
<h1 align="center">Website</h1>
<p align="center">
  Built using <a href="https://v2.docusaurus.io/" target="_blank">Docusaurus 2</a>, a modern static website generator.
</p>

## How to develop

```sh
yarn # install deps
yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

## Build

```sh
yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## New version

```sh
yarn docusaurus docs:version x.x.x
```

The versions of the all docs split into two parts, one is the **latest (in docs/)** and the others are **versioned (in versioned_docs/)**. When a version has been released, the current latest docs will be copied into versiond_docs (by running the command above).

## How to contribute

Mostly you only need to modify the content in the `docs/`, but if you want some versions to take effect as the latest, please also update the same files in the `versioned_docs/` dir.

## License

Same as Chaos Mesh
