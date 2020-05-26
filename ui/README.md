<p align="center">
  <img src="public/logo.svg" width="256" alt="Chaos Mesh Logo" />
</p>
<h1 align="center">Chaos Mesh Dashboard</h1>

> Note: Currently WIP

A Web UI for Chaos Mesh. Powered by ⚛️ [Create React App](https://github.com/facebook/create-react-app).

You can learn more in the [Create React App documentation](https://facebook.github.io/create-react-app/docs/getting-started).

To learn React, check out the [React documentation](https://reactjs.org/).

## How to develop

Into the `ui` folder, run:

```sh
yarn && yarn start
```

Your browser will open <localhost:3000> automatically.

## The rules we followed

For better collaboration and review, we have developed a few rules to help us develop better.

- [Ts or JS](#ts-or-js)
- [Styles](#styles)
- [Be Compact](#be-compact)
- [Necessary Test](#necessary-test)

**Before you contribute, you must read the following carefully.**

### TS or JS

First, we use [husky](https://github.com/typicode/husky) and [lint-staged](https://github.com/okonet/lint-staged) to make [prettier](https://prettier.io/) format our code automatically before commit.

And also, because some of us use vscode to develop the dashboard, we recommend to use [sort-imports](https://marketplace.visualstudio.com/items?itemName=amatiasq.sort-imports) to format all imports. (This is optional, we will not force you to use)

### Styles

Currently, we use `@material-ui/styles` to isolate each component styles.

Maybe we will use `sass` in the future.

But no matter what to use, we hope you can follow this order **(Don't care about their value)** to organize all styles:

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

**Don't include no used deps.**

**Don't let your code be too long-winded, there will be a lot of elegant writing.**

### Necessary Test

**Every new features must have a unit test.**

## Authors

Originally designed by PingCAP FE.

## License

Same as chaos mesh
