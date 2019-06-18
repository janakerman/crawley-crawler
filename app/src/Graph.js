import React from 'react';
import {
  forceSimulation,
  forceLink,
  forceManyBody,
  forceCenter,
  forceCollide,
  forceX,
  forceY,
} from 'd3-force';

import {
  selectAll,
  select,
} from 'd3-selection';


const crawls = [
  { "CrawlID": "c8bdff31-562c-4930-a2aa-74a568263b21", "ParentURL": "https://janakerman.co.uk", "ChildURLs": ["https://janakerman.co.uk/", "https://janakerman.co.uk/cloudformation-dynamodb-data-ingest/", "https://janakerman.co.uk/relational-data-in-dynamodb/", "https://janakerman.co.uk/serverless-acceptance-test-environments-jest/", "https://janakerman.co.uk/docker-git-clone/"] },
  { "CrawlID": "c8bdff31-562c-4930-a2aa-74a568263b21", "ParentURL": "https://janakerman.co.uk/cloudformation-dynamodb-data-ingest/", "ChildURLs": ["https://janakerman.co.uk/", "https://janakerman.co.uk/relational-data-in-dynamodb/", "https://janakerman.co.uk/static/custom-resource-09600b5a1e0210ce19dffc4fa671e311-ed5d9.jpeg", "https://janakerman.co.uk/static/stack-diagram-2c0a88c8645d5b0946c678ee2878ac69-0cef2.jpeg", "https://janakerman.co.uk/relational-data-in-dynamodb/"] },
  { "CrawlID": "c8bdff31-562c-4930-a2aa-74a568263b21", "ParentURL": "https://janakerman.co.uk/serverless-acceptance-test-environments-jest/", "ChildURLs": ["https://janakerman.co.uk/", "https://janakerman.co.uk/docker-git-clone/", "https://janakerman.co.uk/relational-data-in-dynamodb/"] },
  { "CrawlID": "c8bdff31-562c-4930-a2aa-74a568263b21", "ParentURL": "https://janakerman.co.uk/docker-git-clone/", "ChildURLs": ["https://janakerman.co.uk/", "https://janakerman.co.uk/serverless-acceptance-test-environments-jest/"] },
  { "CrawlID": "c8bdff31-562c-4930-a2aa-74a568263b21", "ParentURL": "https://janakerman.co.uk/relational-data-in-dynamodb/", "ChildURLs": ["https://janakerman.co.uk/", "https://janakerman.co.uk/serverless-acceptance-test-environments-jest/", "https://janakerman.co.uk/cloudformation-dynamodb-data-ingest/"] },
  { "CrawlID": "c8bdff31-562c-4930-a2aa-74a568263b21", "ParentURL": "https://janakerman.co.uk/", "ChildURLs": ["https://janakerman.co.uk/", "https://janakerman.co.uk/cloudformation-dynamodb-data-ingest/", "https://janakerman.co.uk/relational-data-in-dynamodb/", "https://janakerman.co.uk/serverless-acceptance-test-environments-jest/", "https://janakerman.co.uk/docker-git-clone/"] }
]
function toNodes(messages) {
  const nodes = new Set()
  const links = []

  messages.forEach(m => {
    // Add nodes
    nodes.add(m.ParentURL)
    m.ChildURLs.forEach(childURL => {
      nodes.add(childURL)
      links.push({ source: m.ParentURL, target: childURL })
    })
  })

  return {
    nodes: Array.from(nodes).map(n => ({ site: n })),
    links: links,
  }
}

export class Graph extends React.Component {
  constructor(props) {
    super(props)

    this.createGraph = this.createGraph.bind(this)

    this.state = {
      ...toNodes(crawls)
    }
  }

  componentDidMount() {
    this.createGraph()
  }

  componentDidUpdate() {
    this.createGraph()
  }

  createGraph() {
    const svg = this.svg
    const height = 500
    const width = 500
    const { nodes, links } = this.state

    console.log(nodes)
    console.log(links)

    //set up the simulation 
    //nodes only for now 
    var simulation = forceSimulation()
      //add nodes
      .nodes(nodes);

    //add forces
    //we're going to add a charge to each node 
    //also going to add a centering force
    simulation
      .force("charge_force", forceManyBody())
      .force("center_force", forceCenter(width / 2, height / 2));

    //draw circles for the nodes 
    var node = select(svg).append("g")
      .attr("class", "nodes")
      .selectAll("circle")
      .data(nodes)
      .enter()
      .append("circle")
      .attr("r", 5)
      .attr("fill", "red");

    //add tick instructions: 
    simulation.on("tick", tickActions);




    var link_force = forceLink(links)
      .id(function (d) { return d.site; })

    simulation.force("links", link_force)

    //draw lines for the links 
    var link = select(svg).append("g")
      .attr("class", "links")
      .selectAll("line")
      .data(links)
      .enter().append("line")
      .attr("stroke-width", 2)
      .attr("stroke", 'black')


    function tickActions() {
      //update circle positions each tick of the simulation 
      node
        .attr("cx", function (d) { return d.x; })
        .attr("cy", function (d) { return d.y; });

      //update link positions 
      //simply tells one end of the line to follow one node around
      //and the other end of the line to follow the other node around
      link
        .attr("x1", function (d) { return d.source.x; })
        .attr("y1", function (d) { return d.source.y; })
        .attr("x2", function (d) { return d.target.x; })
        .attr("y2", function (d) { return d.target.y; });

    }


  }

  render() {
    return (
      <svg
        ref={el => this.svg = el}
        width={500} height={500}>
      </svg>
    )
  }
}

export default Graph
