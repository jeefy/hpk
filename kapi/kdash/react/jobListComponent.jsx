import React from 'react';
import {render} from 'react-dom';
import axios from 'axios';
import Moment from 'react-moment';
import moment from 'moment';
// import AwesomeComponent from './AwesomeComponent.jsx';

class JobListComponent extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      jobs:     [],
      totals:   {},
      rates:    {},
      runtimes: {},
    };
  }
  componentDidMount() {
    axios.get(`/jobs`)
      .then(jobs => {
        this.setState({
          "jobs":     jobs.data,
          "totals":   {},
          "rates":    {},
          "runtimes": {},
        });
        jobs.data.map(job => {
          axios.get(`/jobs_use/${job.name}`)
            .then(res => {
              var totals   = this.state.totals;
              var rates    = this.state.rates;
              var runtimes = this.state.runtimes;

              if(res.data && "total_cost" in res.data[0]) {
                totals[job.name]   = parseFloat(res.data[0].total_cost).toFixed(2);
                rates[job.name]    = res.data[0].memory_cost + res.data[0].cpu_cost;
                runtimes[job.name] = res.data[0].duration;
              } else {
                totals[job.name]   = "?"
                rates[job.name]    = "?"
                runtimes[job.name] = "?"
              }
              this.setState({"jobs": jobs.data, "totals": totals, "rates":rates, "runtimes":runtimes});
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
                  { 'completionTime' in job.changelog[job.changelog.length-1].status ? (
                    <Moment format="YYYY-MM-DD HH:mm">
                      {job.changelog[job.changelog.length-1].status.completionTime.replace('T', ' ')}
                    </Moment>
                  ) : (
                    <b>Still running</b>
                  )}
                </td>
                <td>{job.changelog[0].spec.template.spec.containers[0].command.join(' ')}</td>
                <td>
                  { job.name in this.state.totals ? (
                    <i>${ this.state.totals[job.name] }</i>
                  ) : (
                    <b>?</b>
                  )}
              </td>
                <td>
                  <a className="btn btn-sm btn-info" target="_blank" href={"/jobs/" + job.name + "/logs"}>Logs</a>
                  <a className="btn btn-sm btn-success" target="_blank" href={"/jobs_use/" + job.name + ""}>Use</a>
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
