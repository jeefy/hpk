[![wercker status](https://app.wercker.com/status/bff5b2ff478d8902c046532fb8724988/s/master "wercker status")](https://app.wercker.com/project/byKey/bff5b2ff478d8902c046532fb8724988)

# go-react


## How I created the app

Mostly based on the following sources https://www.codementor.io/reactjs/tutorial/beginner-guide-setup-reactjs-environment-npm-babel-6-webpack
https://github.com/cloudnativego/react-zombieoutbreak



```
$ npm init
$ npm i webpack -S
$ npm i react react-dom -S
$ npm i babel-loader babel-core babel-preset-es2015 babel-preset-react -S


```




Create **.babelrc**

```
{
  "presets" : ["es2015", "react"]
}
```



Create **webpack.config.js**

```
var webpack = require('webpack');
var path = require('path');

var BUILD_DIR = path.resolve(__dirname, 'static/js');
var APP_DIR = path.resolve(__dirname, 'react');

var config = {
  entry: APP_DIR + '/index.jsx',
  output: {
    path: BUILD_DIR,
    filename: 'bundle.js'
  },
  module : {
    loaders : [
      {
        test : /\.jsx?/,
        include : APP_DIR,
        loader : 'babel'
      }
    ]
  }
};

module.exports = config;

```

update **packages.json** as a tool runner

```
{
  // ...
  ,     // be sure to add the comma after the previous statement
  "scripts": {
    "dev": "webpack -d --watch",
    "build" : "webpack -p"
  }
  // ...
}
```

Create **react/index.jsx**

```
import React from 'react';
import {render} from 'react-dom';
// import AwesomeComponent from './AwesomeComponent.jsx';



class AwesomeComponent extends React.Component {

  constructor(props) {
    super(props);
    this.state = {likesCount : 0};
    this.onLike = this.onLike.bind(this);
  }

  onLike () {
    let newLikesCount = this.state.likesCount + 1;
    this.setState({likesCount: newLikesCount});
  }

  render() {
    return (
      <div>
        Likes : <span>{this.state.likesCount}</span>
        <div><button onClick={this.onLike}>Like Me</button></div>
      </div>
    );
  }

}

export default AwesomeComponent;




class App extends React.Component {
  render () {
    return (
      <div>
        <p> Hello React!</p>
        <AwesomeComponent />
      </div>
    );
  }
}

render(<App/>, document.getElementById('app'));


```
Create **static/index.html**

```
<html>
  <head>
    <meta charset="utf-8">
    <title>React.js using NPM, Babel6 and Webpack</title>
  </head>
  <body>
    <div id="app" />
    <script src="public/bundle.js" type="text/javascript"></script>
  </body>
</html>
```

Run webpack in dev mode (with watch capabilities)
```

$ npm run dev

```
Run webpack in prod mode (with minification)
```

$ npm run build

```

## Golang Code

Create **main.go**

```

package main

import (
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8100"
	}

	server := NewServer()
	server.Run(":" + port)
}

var webRoot string

// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {

	formatter := render.New(render.Options{
		IndentJSON: true,
	})

	n := negroni.Classic()
	mx := mux.NewRouter()

	initRoutes(mx, formatter)

	n.UseHandler(mx)
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render) {
	webRoot = os.Getenv("WEBROOT")
	if len(webRoot) == 0 {
		root, err := os.Getwd()
		if err != nil {
			panic("Could not retrieve working directory")
		} else {
			webRoot = root
		}
	}

	mx.PathPrefix("/").Handler(http.FileServer(http.Dir(webRoot + "/static/")))
}


```

Build and run the server

```
$ go get
$ go run main.go

```

Open ``http://localhost:8100/``
