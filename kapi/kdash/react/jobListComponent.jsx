import React from 'react';
import {render} from 'react-dom';
import axios from 'axios';
import Moment from 'react-moment';
// import AwesomeComponent from './AwesomeComponent.jsx';

class JobListComponent extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      jobs: []
    };
  }
  componentDidMount() {
    axios.get(`/jobs`)
      .then(res => {
        this.setState({ "jobs": res.data });
      });
  }

  /* {this.state.jobs.map(job =>
    <JobComponent job={job} />
  )}*/
  render() {
    return (
      <table className="table table-striped">
        <tbody>
          <tr>
            <th>Name</th>
            <th>Start</th>
            <th>End</th>
            <th>Command</th>
            <th>Concurrency</th>
            <th>Logs</th>
          </tr>
          {
            this.state.jobs.map(function(job) {
              return <tr key={job._id}>
                <td><a target="_blank" href={"/jobs/" + job.name}>{job.name}</a></td>
                <td>
                  <Moment format="YYYY-MM-DD HH:mm">
                    {job.changelog[0].objectmeta.creationtimestamp.time.replace('T', ' ')}
                  </Moment>
                </td>
                <td>
                  <Moment format="YYYY-MM-DD HH:mm">
                    {job.changelog[job.changelog.length-1].status.completiontime.time.replace('T', ' ')}
                  </Moment>
                </td>
                <td>{job.changelog[0].spec.template.spec.containers[0].command.join(' ')}</td>
                <td>{job.changelog[0].spec.parallelism}</td>
                <td><a target="_blank" href={"/jobs/" + job.name + "/logs"}>Logs</a></td>
              </tr>
            })
          }
        </tbody>
      </table>
    );
  }
}

export default JobListComponent;
