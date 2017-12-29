import React from 'react';
import {render} from 'react-dom';
import axios from 'axios';
import Moment from 'react-moment';
// import AwesomeComponent from './AwesomeComponent.jsx';

class JobListComponent extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      jobs: [],
      costs: {},
    };
  }
  componentDidMount() {
    axios.get(`/jobs`)
      .then(jobs => {
        this.setState({ "jobs": jobs.data });
        jobs.data.map(job => {
          axios.get(`/jobs_use/${job.name}`)
            .then(res => {
              var use = res.data[0];
              var cost = use.parallelism * ((use.cpu_cost * (use.cpu_val/1000)) + (use.memory_cost * (use.memory_val/1000)));
              console.log("It cost " + cost);
              var costs = this.state.costs;
              costs[job.name] = cost;
              this.setState({"jobs": jobs.data, "costs": costs});
            });
        });
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
            <th>Cost</th>
            <th>Logs</th>
          </tr>
          {
            this.state.jobs.map(job => {
              return <tr key={job._id}>
                <td><a target="_blank" href={"/jobs/" + job.name}>{job.name}</a></td>
                <td>
                  <Moment format="YYYY-MM-DD HH:mm">
                    {job.changelog[0].metadata.creationTimestamp.replace('T', ' ')}
                  </Moment>
                </td>
                <td>
                  <Moment format="YYYY-MM-DD HH:mm">
                    {job.changelog[job.changelog.length-1].status.completionTime.replace('T', ' ')}
                  </Moment>
                </td>
                <td>{job.changelog[0].spec.template.spec.containers[0].command.join(' ')}</td>
                <td>{ this.state.costs[job.name] }</td>
                <td>
                  <a target="_blank" href={"/jobs/" + job.name + "/logs"}>Logs</a>
                  <a target="_blank" href={"/jobs_use/" + job.name + ""}>Use</a>
                </td>
              </tr>
            })
          }
        </tbody>
      </table>
    );
  }
}

export default JobListComponent;
