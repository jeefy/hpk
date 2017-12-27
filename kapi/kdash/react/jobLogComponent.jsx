import React from 'react';
import {render} from 'react-dom';
import axios from 'axios';

class JobLogComponent extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      log: {}
    };
  }
  componentDidMount() {
    axios.get(`/jobs/${this.props.job.name}/logs`)
      .then(res => {
        const joblog = res.data.data.children.map(obj => obj.data);
        this.setState({ "log": joblog });
      });
  }
}

export default JobLogComponent;
