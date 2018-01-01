import React from 'react';
import {render} from 'react-dom';
import axios from 'axios';
import ConfigFormComponent from './configFormComponent.jsx';

class AllocationListComponent extends React.Component {
  constructor(props) {
    super(props);


    this.refreshData = this.refreshData.bind(this);
    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);

    this.state = {
      allocations: [],
      alloc: {
        name: "",
        balance: "",
      }
    };
  }

  refreshData() {
    var confObj = {};
    axios.get(`/allocations`)
      .then(res => {
        var state = this.state;
        state.allocations = res.data;
        state.alloc = {
          name: "",
          balance: "",
        }
        this.setState(state);
      });
  }

  componentDidMount() {
    this.refreshData();
  }

  loadAllocation(alloc){
    var state = this.state;
    state.alloc = alloc;
    this.setState(state);
  }

  handleInputChange(event) {
    const target = event.target;
    const value = target.type === 'checkbox' ? target.checked : target.value;
    const name = target.name;

    var state = this.state;
    state.alloc[name] = value;
    this.setState(state);
  }

  handleSubmit(event) {
    // post request to `/config/{this.state.key}`
    event.preventDefault();
    var bodyFormData = new FormData();
    bodyFormData.set("balance", this.state.alloc.balance)
    axios({
      method: 'post',
      url: "/allocations/" + this.state.alloc.name,
      data: bodyFormData,
      config: { headers: {'Content-Type': 'multipart/form-data' }}
    }).then(res => {
        this.refreshData();
      });
  }

  allocationForm() {
    return (
      <form onSubmit={this.handleSubmit}>
        <input name="name" placeholder="Allocation Name" className="input" type="text" value={this.state.alloc.name} onChange={this.handleInputChange} />
        <input name="balance" placeholder="Balance" className="input" type="text" value={this.state.alloc.balance} onChange={this.handleInputChange} />
        <input type="submit" />
      </form>
    )
  }

  loadAllocation(alloc) {
    var state = this.state;
    state.alloc = {
      name:    alloc.name,
      balance: alloc.balance,
    };
    this.setState(state);
  }

  render() {
    return (
      <div>
        <table className="table table-striped">
          <thead>
            <tr>
              <th>Name</th>
              <th>Balance</th>
            </tr>
          </thead>
          <tbody>
            {this.state.allocations.map(alloc => {
              return <tr key={alloc._id}>
                <td>
                  <a href="#" onClick={() => this.loadAllocation(alloc)}> {alloc.name} </a>
                </td>
                <td>${alloc.balance}</td>
              </tr>
            })}
          </tbody>
        </table>
        <div>
          {this.allocationForm()}
        </div>
      </div>
    );
  }
}

export default AllocationListComponent;
