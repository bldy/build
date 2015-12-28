
var Graphs = React.createClass({
    getInitialState: function() {
        return {
            tab: 2
        };
    },
    depGraph: function() {
        this.setState({
            tab: 1
        });

    },
    depTree: function() {
        this.setState({
            tab: 2
        });
    },
    render: function() {
        return (
            <div>
                <a href="#" onClick={this.depGraph}>dep graph</a>{' '}
                    <a href="#" onClick={this.depTree}>build graph</a>
                    <div id="depgraph" style={{display: this.state.tab===1? "block" :"none"}}/>
                    <div id="buildgraph" style={{display: this.state.tab===2? "block" :"none"}}/>
                    <DepGraph/>
                    <BuildGraph/>
                </div>
            );
        }

    });
