import React from 'react';
import './App.css';
import 'bootstrap/dist/css/bootstrap.css';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Form from 'react-bootstrap/Form';
import Badge from 'react-bootstrap/Badge';
import Button from 'react-bootstrap/Button';

import Graph from './Graph'

const badPracticeURL = 'https://higkb65cx1.execute-api.eu-west-2.amazonaws.com/dev'
const badPracticeWebsocketURL = 'wss://4epq4ctp9a.execute-api.eu-west-2.amazonaws.com/dev'

const testEvents = [
  {"CrawlID":"c8bdff31-562c-4930-a2aa-74a568263b21","ParentURL":"https://janakerman.co.uk","ChildURLs":["https://janakerman.co.uk/","https://janakerman.co.uk/cloudformation-dynamodb-data-ingest/","https://janakerman.co.uk/relational-data-in-dynamodb/","https://janakerman.co.uk/serverless-acceptance-test-environments-jest/","https://janakerman.co.uk/docker-git-clone/"]},
  {"CrawlID":"c8bdff31-562c-4930-a2aa-74a568263b21","ParentURL":"https://janakerman.co.uk/cloudformation-dynamodb-data-ingest/","ChildURLs":["https://janakerman.co.uk/","https://janakerman.co.uk/relational-data-in-dynamodb/","https://janakerman.co.uk/static/custom-resource-09600b5a1e0210ce19dffc4fa671e311-ed5d9.jpeg","https://janakerman.co.uk/static/stack-diagram-2c0a88c8645d5b0946c678ee2878ac69-0cef2.jpeg","https://janakerman.co.uk/relational-data-in-dynamodb/"]},
  {"CrawlID":"c8bdff31-562c-4930-a2aa-74a568263b21","ParentURL":"https://janakerman.co.uk/serverless-acceptance-test-environments-jest/","ChildURLs":["https://janakerman.co.uk/","https://janakerman.co.uk/docker-git-clone/","https://janakerman.co.uk/relational-data-in-dynamodb/"]},
  {"CrawlID":"c8bdff31-562c-4930-a2aa-74a568263b21","ParentURL":"https://janakerman.co.uk/docker-git-clone/","ChildURLs":["https://janakerman.co.uk/","https://janakerman.co.uk/serverless-acceptance-test-environments-jest/"]},
  {"CrawlID":"c8bdff31-562c-4930-a2aa-74a568263b21","ParentURL":"https://janakerman.co.uk/relational-data-in-dynamodb/","ChildURLs":["https://janakerman.co.uk/","https://janakerman.co.uk/serverless-acceptance-test-environments-jest/","https://janakerman.co.uk/cloudformation-dynamodb-data-ingest/"]},
  {"CrawlID":"c8bdff31-562c-4930-a2aa-74a568263b21","ParentURL":"https://janakerman.co.uk/","ChildURLs":["https://janakerman.co.uk/","https://janakerman.co.uk/cloudformation-dynamodb-data-ingest/","https://janakerman.co.uk/relational-data-in-dynamodb/","https://janakerman.co.uk/serverless-acceptance-test-environments-jest/","https://janakerman.co.uk/docker-git-clone/"]}
]


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
    const uuid = uuidv4()
    const data = {
      RootURLs: this.state.sites,
      CrawlID: uuid,
    }

    this.setState({
      ...this.state,
      isLoading: true
    })
    
    this.subscribe(uuid).then(() => {
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
    })
  }

  subscribe(crawlID) {
    return new Promise((resolve) => {
      console.log('Opening websocket connection')
      const webSocket = new WebSocket(badPracticeWebsocketURL)

      webSocket.onmessage = event => {
        var msg = JSON.parse(event.data);
        console.log(`Received message: ${JSON.stringify(msg)}`)
      }

      webSocket.onopen = event => {
        console.log('Websocket connected')

        const subscribtion = { action:"subscribe", CrawlID: crawlID }
        console.log(`Sending message: ${JSON.stringify(subscribtion)}`)
        webSocket.send(JSON.stringify(subscribtion));

        resolve()
      };
    })
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
      <Row>
        <Graph/>
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
