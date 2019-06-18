import React from 'react';
import logo from './logo.svg';
import './App.css';
import 'bootstrap/dist/css/bootstrap.css';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Form from 'react-bootstrap/Form';
import Badge from 'react-bootstrap/Badge';
import Button from 'react-bootstrap/Button';

class AppContent extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      inputText: '',
      isLoading: false,
      sites: []
    }
  }

  inputUpdated(event) {
    this.setState({
      ...this.state,
      inputText: event.target.value
    })
  }

  inputEntered(event) {
    console.log(event)
    if (event.key === "Enter") {
      event.preventDefault()
      this.setState((state) => ({
        ...state,
        inputText: '',
        sites: [...state.sites, state.inputText]
      }))
    }
  }

  render() {
    const { isLoading } = this.state;
    const siteLimitReached = this.state.sites.length >= 2

    return (<Container>
      <h1>Crawler</h1>
      <Row>
        <Col>
          <Form>
            <Form.Group controlId="exampleForm.ControlInput1">
              <Form.Control 
                placeholder="e.g https://janakerman.co.uk"
                value={this.state.inputText}
                onChange={e => this.inputUpdated(e)}
                onKeyPress={e => this.inputEntered(e)}  
                disabled={siteLimitReached}
              />
              {this.state.sites.map(site => <h1><Badge variant="secondary">{site}</Badge></h1>)}
            </Form.Group>
          </Form>
        </Col>
      </Row>
      <Row>
        <Col>
          <Button
            variant="primary"
            enabled
            disabled={isLoading || this.state.sites.length === 0}
            onClick={!isLoading ? this.handleClick : null}
          >
            {isLoading ? 'Loading…' : 'Go'}
          </Button>
        </Col>
      </Row>
    </Container>)
  }
}

function App() {
  return (
    <div className="App">
      <AppContent/>
    </div>
  );
}

export default App;
