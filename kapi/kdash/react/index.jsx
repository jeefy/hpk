import React from 'react';
import ReactDOM from 'react-dom';

import JobComponent from './jobComponent.jsx';
import JobListComponent from './jobListComponent.jsx';
import JobLogComponent from './jobLogComponent.jsx';
import ConfigListComponent from './configListComponent.jsx';
import AllocationListComponent from './allocationListComponent.jsx';

module.exports = {
  JobComponent,
  JobListComponent,
  JobLogComponent,
  ConfigListComponent,
  AllocationListComponent
}

// custom tags
function render (tag, Comp) {
  document.createElement(tag);

  const nodes = Array.from(document.getElementsByTagName(tag));
  nodes.map((node, i) => renderNode(tag, Comp, node, i));

  return Comp;
}

function renderNode (tag, Comp, node, i) {
  let attrs = Array.prototype.slice.call(node.attributes);
  let props = {
    key: `${ tag }-${ i }`,
  };

  attrs.map((attr) => props[attr.name] = attr.value);

  if (!!props.class) {
    props.className = props.class;
    delete props.class;
  }

  ReactDOM.render(
    <Comp { ...props }/>,
    node
  );
}

render("JobLogComponent", JobLogComponent);
render("JobComponent", JobComponent);
render("JobListComponent", JobListComponent);
render("ConfigListComponent", ConfigListComponent)
render("AllocationListComponent", AllocationListComponent)
