import React from 'react';
import './App.scss';
import { Term } from './Term';
import axios, { AxiosInstance } from 'axios';
import { LoginForm, LoginFormProps } from './LoginForm'
import { Form, Spin, Icon, Button, message } from 'antd'

interface LoginData {
  token: string
}

interface AppState {
  state: State
}

enum State {
  Loading = "Loading",
  Login = "Login",
  Terminal = "Terminal",
  Closed = "Closed"
}

export class App extends React.Component<{}, AppState> {
  private io: AxiosInstance;
  public state = {state: State.Loading};

  constructor(props: {}) {
    super(props);
    this.io = axios.create();
  }

  componentDidMount() {
    this.io.get("/api/v1/refresh").then(resp => {
      this.setState({ state: State.Terminal })
    }).catch(error => {
      this.setState({ state: State.Login })
    })
  }
    
  login = (username: string, password: string) => {
    this.io.post<LoginData>("/api/v1/login", {
      username: username,
      password: password
    }).then(resp => {
      this.setState({ state: State.Terminal })
    }).catch(error => {
      message.error('Sorry, invalid username or password.');
    })
  }

  render() {
    const WrappedLoginForm = Form.create<LoginFormProps>({ name: 'normal_login' })(LoginForm);
    var content;
    switch (this.state.state) {
      case State.Loading: {
        const antIcon = <Icon type="loading" style={{ fontSize: 24 }} spin />;
        content = <div className="AppContent"><Spin indicator={antIcon} /></div>;
        break;
      }
      case State.Login: {
        content = <div className="AppContent"><WrappedLoginForm Login={ this.login } /></div>;
        break;
      }
      case State.Terminal: {
        content = <Term onClose={ () => this.setState({state: State.Closed}) }/>;
        break;
      }
      case State.Closed: {
        content = (
        <div className="AppContent">
          <Button type="primary" onClick={ () => { window.location.reload(); } }>Refresh</Button>
        </div>);
        break;
      }
    }
    return (
      <div className='App'>
        {content}
      </div>
    );
  }
}
