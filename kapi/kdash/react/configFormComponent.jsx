import React from 'react';
import {render} from 'react-dom';
import axios from 'axios';

class ConfigFormComponent extends React.Component {
  constructor(props) {
    super(props);

    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);

    this.state = {
      config: this.props.state.config,
      key: "",
      val: ""
    };
  }

  componentWillReceiveProps(nextProps) {
    // You don't have to do this check first, but it can help prevent an unneeded render
    if(nextProps.configKey != this.state.key || nextProps.configVal != this.state.val) {
      this.setState({config: this.state.config, key: nextProps.configKey, val: nextProps.configVal });
    }
  }

  handleInputChange(event) {
    const target = event.target;
    const value = target.type === 'checkbox' ? target.checked : target.value;
    const name = target.name;

    this.setState({
      [name]: value
    });
  }

  handleSubmit(event) {
    // post request to `/config/{this.state.key}`
    event.preventDefault();
    var bodyFormData = new FormData();
    bodyFormData.set("val", this.state.val)
    axios({
      method: 'post',
      url: `/config/${this.state.key}`,
      data: bodyFormData,
      config: { headers: {'Content-Type': 'multipart/form-data' }}
    }).then(res => {
        this.props.refreshData();
        this.setState({key:"",val:""});
      });
  }

  render() {
    return (
      <div>
        <form onSubmit={this.handleSubmit}>
          <table>
            <tbody>
              <tr>
                <td>
                  <input name="key" type="text" value={this.state.key} onChange={this.handleInputChange} />
                </td>
                <td>
                  <input name="val" type="text" value={this.state.val} onChange={this.handleInputChange} />
                </td>
                <td>
                  <input type="submit" />
                </td>
              </tr>
            </tbody>
          </table>
        </form>
      </div>
    );
  }
}

export default ConfigFormComponent;
