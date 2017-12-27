import React from 'react';
import {render} from 'react-dom';
import axios from 'axios';

class JobComponent extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      job: {}
    };
    if('_id' in this.props.job){
      this.state.job = this.props.job;
    }
  }
  componentDidMount() {
    if(!'_id' in this.props.job){
      axios.get(`/jobs/${this.props.job.name}`)
        .then(res => {
          const job = res.data.data.children.map(obj => obj.data);
          this.setState({ "job": job });
        });
    }
  }
}

export default JobComponent;
