import React, { Component } from 'react';
import { List, Card, Avatar, Row, Col, Carousel, PageHeader } from 'antd';
import './App.css';

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      list: null,
    };
  }

  async componentDidMount() {
    const raw = await fetch("/services");
    const res = await raw.json();

    this.setState({
      list: res.map(name => {
          return {
              name,
              desc: ""
          }
      })
    });
  }

  render() {
    return (<div style={{
      padding: 24,
    }}>
      <PageHeader
        title='Chaos Dashboard'
        subTitle='chaos monitor for each namespace'
        style={{
          backgroundColor: '#FFFFFF',
        }}
      />
      <Row type="flex" justify="space-around" style={{marginTop: 12, backgroundColor: '#FFFFFF', minHeight: 800, paddingTop: 8}}>
        <Col span={18}>
          <List
            grid={{
              gutter: 16,
              xs: 2,
              sm: 2,
              md: 3,
              lg: 3,
              xl: 4,
              xxl: 4,
            }}
            dataSource={this.state.list || []}
            renderItem={item => (
              <List.Item>
                <a href={`/dashboard/${item.name}/d/H3wXWqfWz/chaos-monitor`} target="_blank" rel="noopener noreferrer">
                  <Card
                    style={{width: 450}}
                    cover={<Carousel autoplay>
                      <iframe src={`/dashboard/${item.name}/d-solo/H3wXWqfWz/chaos-monitor?&panelId=2`}
                              title='duration'
                              width="450" 
                              height="225" 
                              frameBorder="0">        
                      </iframe>
                      <iframe src={`/dashboard/${item.name}/d-solo/H3wXWqfWz/chaos-monitor?&panelId=3`}
                              title='QPS' 
                              width="450" 
                              height="225" 
                              frameBorder="0">        
                      </iframe>
                    </Carousel>
                    }
                  >
                    <Card.Meta
                      avatar={<Avatar style={{background: this.color(item.name)}}>{item.name[0].toUpperCase()}</Avatar>}
                      title={item.name}
                      description={item.desc}
                    />
                  </Card>
                </a>
              </List.Item>
            )}
            loading={!this.state.list}
          />
        </Col>
      </Row>
    </div>);
  }

  color(str) {
    const char = str[0];
    const colors = ["#f50", "#2db7f5", "#87d068", "#108ee9"];
    return colors[char.charCodeAt() % colors.length];
  }
}

export default App;
