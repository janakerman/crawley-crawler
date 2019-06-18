import React from 'react';
import './App.css';
import 'bootstrap/dist/css/bootstrap.css';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Form from 'react-bootstrap/Form';
import Badge from 'react-bootstrap/Badge';
import Button from 'react-bootstrap/Button';

const badPracticeURL = 'https://higkb65cx1.execute-api.eu-west-2.amazonaws.com/dev'

function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    var r = Math.random() * 16 | 0, v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

class AppContent extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      inputText: '',
      isLoading: false,
      sites: []
    }
    this.doCrawl = this.doCrawl.bind(this)
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

  doCrawl() {

    this.setState({
      ...this.state,
      isLoading: true
    })

    const uuid = uuidv4()

    const data = {
      RootURLs: this.state.sites, // TODO: Update lambda to handle multiple sites.
      CrawlID: uuid,
    }

    console.log(`Posting crawl: ${JSON.stringify(data)}`)

    fetch(`${badPracticeURL}/crawl`, {
      method: 'POST',
      body: JSON.stringify(data),
      mode: 'no-cors',
      headers:{
        'Content-Type': 'application/json'
      }
    })
    .then(rest => console.log(`Response: ${rest.status}`))
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
              {this.state.sites.map(site => <h1 key={site}><Badge variant="secondary">{site}</Badge></h1>)}
            </Form.Group>
          </Form>
        </Col>
      </Row>
      <Row>
        <Col>
          <Button
            variant="primary"
            disabled={isLoading || this.state.sites.length === 0}
            onClick={!isLoading ? this.doCrawl : null}
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
