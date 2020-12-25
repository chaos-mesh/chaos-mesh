<!-- markdownlint-disable-file MD033 -->
<!-- markdownlint-disable-file MD041 -->

<img src="../static/logo.svg" width="450" alt="Chaos Mesh Logo" />

# Dashboard

A Web UI for Chaos Mesh. Powered by ⚛️ [Create React App](https://github.com/facebook/create-react-app).

## How to develop

### Main technologies

<div style="display: flex; align-items: center;">
<a href="https://www.typescriptlang.org/">
  <img src="https://upload.wikimedia.org/wikipedia/commons/4/4c/Typescript_logo_2020.svg" height="45" alt="TypeScript" />
</a>
<a href="https://reactjs.org/">
  <img src="https://upload.wikimedia.org/wikipedia/commons/a/a7/React-icon.svg" height="45" alt="React" />
</a>
<a href="https://redux.js.org/">
  <img src="https://redux.js.org/img/redux.svg" height="45" alt="Redux" />
</a>
<a href="https://material-ui.com/" style="margin-left: 15px;">
  <img src="https://material-ui.com/static/logo_raw.svg" height="45" alt="Material UI" />
</a>
</div>

### Bootstrap

#### Global env

If you haven't installed the nodejs environment, checkout [https://nodejs.org/en/download/](https://nodejs.org/en/download/).

And also, we use [Yarn 1](https://classic.yarnpkg.com/en/) as the dependency management. Maybe we will migrate to Yarn 2 in the future, but not now.

#### Install deps

If you just cloned a fresh Chaos Mesh repo, into the `ui` folder, run:

```sh
yarn
```

This command will install all deps the dashboard needed.

Then, you need to provide an API server as a proxy, it will pass into an env var which named: `REACT_APP_API_URL`. There are three ways to get it:

- From other deployed Chaos Mesh Dashboard

  If you have Chaos Mesh deployed in the cluster, you can use the dashboard service url as the proxy.

  A simple way:

  ```sh
  kubectl port-forward -n chaos-testing svc/chaos-dashboard 2333:2333
  ```

- From local deployed Chaos Mesh Dashboard

  Same as above. Only the cluster is local. (E.g., [kind](https://kind.sigs.k8s.io/) or [minikube](https://minikube.sigs.k8s.io/))

- From local server

  There have two ways to run chaos-dashboard server in your terminal:

  - `go run cmd/chaos-dashboard/main.go`
  - `SWAGGER=1 make chaos-dashboard && ./bin/chaos-dashboard`

  One is real-time, the other needs to be compiled before use. The compiled bundle the extra Swagger API HTML into the binary file.

  For the other usage of `make xxx` that please refer the [Makefile](https://github.com/chaos-mesh/chaos-mesh/blob/master/Makefile).

### Structure

```sh
src
├── @types
├── api
├── components
├── components-mui
├── i18n
├── lib
├── pages
├── reducers
├── routes.tsx
├── slices
├── store.ts
└── theme.ts
```

The above tree structure explained as follow:

- `@types` place global types, which use for Typescript.
- `api` place all requests.
- `components` place all packaged components.
- `components-mui` nearly the same as `components`, but inherit from Material UI.
- `i18n` place all translations.
- `lib` place some independent functions and common utils.
- `reducers` (Redux reducers)
- `routes.tsx` (Client routes)
- `slices` [Redux Tookit slices](https://redux-toolkit.js.org/api/createSlice)
- `store.ts` (Redux store)
- `theme` place global theme definitions.

### The rules we followed

For better collaboration and review, we have developed a few rules to help us develop better.

- [TS or JS](#ts-or-js)
- [Styles](#styles)
- [Be Compact](#be-compact)

**Before you contribute, you must read the following carefully.**

### TS or JS

First, we use [husky](https://github.com/typicode/husky) and [lint-staged](https://github.com/okonet/lint-staged) to make [prettier](https://prettier.io/) format our code automatically before commit.

~~And also, because some of us use vscode to develop the dashboard, we recommend to use [sort-imports](https://marketplace.visualstudio.com/items?itemName=amatiasq.sort-imports) to format all imports.~~ (Sort automatically for now)

### Styles

Currently, we use `@material-ui/styles` to isolate each component styles.

We hope you can follow this order **(Don't care about their value)** to organize all styles:

```scss
// Position first
position: relative;
top: 0;
bottom: 0;
left: 0;
right: 0;
// Then display
display: flex;
flex-direction: column;
justify-content: center;
align-items: center;
// Layout
width: 0;
height: 0;
margin: 0;
padding: 0;
// Colors
background: #fff;
color: #000;
// Outside
border: 0;
box-shadow: none;
// Finally, not often used values can be in any order
```

### Be Compact

- **Don't include unused deps.**
- **Don't let your code be too long-winded, there will be a lot of elegant writing.**

## License

Same as Chaos Mesh.
