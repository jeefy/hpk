import React from 'react';
import {render} from 'react-dom';
import axios from 'axios';
import ConfigFormComponent from './configFormComponent.jsx';

class ConfigListComponent extends React.Component {
  constructor(props) {
    super(props);


    this.refreshData = this.refreshData.bind(this);
    this.editConfig = this.editConfig.bind(this);
    this.deleteConfig = this.deleteConfig.bind(this);

    this.state = {
      config: {},
      key: "",
      val: "",
    };
  }

  refreshData() {
    var confObj = {};
    axios.get(`/config`)
      .then(res => {
        this.setState({
          config: res.data,
          key: "",
          val: "",
        });
      });
  }

  editConfig(entry) {
    this.setState({
      config: this.state.config,
      key: entry[0],
      val: entry[1],
    });
    console.log(entry);
    console.log(this.state);
    this.forceUpdate();
  }

  deleteConfig(entry) {
    axios.delete(`/config/${entry[0]}`)
      .then(res => {
        this.refreshData();
      });
  }

  componentDidMount() {
    this.refreshData();
  }

  render() {
    return (
      <div>
        <table className="table table-striped">
          <thead>
            <tr>
              <th>Key</th>
              <th>Value</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {
              Object.entries(this.state.config).map(function(config,index){
                return <tr key={index}>
                  <td>{config[0]}</td>
                  <td>{config[1]}</td>
                  <td>
                    <button type="button" className="btn btn-sm btn-warning" onClick={ () => this.editConfig(config) }>
                      <span className="glyphicon glyphicon-pencil" aria-hidden="true"></span> Edit
                    </button>
                    <button type="button" className="btn btn-sm btn-danger" onClick={ () => this.deleteConfig(config) }>
                      <span className="glyphicon glyphicon-trash" aria-hidden="true"></span> Delete
                    </button>
                  </td>
                </tr>
              }, this)
            }
          </tbody>
        </table>
        <ConfigFormComponent configKey = {this.state.key} configVal = {this.state.val} state = { this.state } refreshData = { this.refreshData } />
      </div>
    );
  }
}

export default ConfigListComponent;
