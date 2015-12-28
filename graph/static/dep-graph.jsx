var selected = "";

var StringArray = React.createClass({
    render: function() {
        var x = _.map(this.props.elems, function(n) {return <code>{n}{'\n'}</code>})
            return (
                <div>{x}</div>
            );
        }

    });
    var DepGraph = React.createClass({
        componentDidMount: function() {
            var self = this;
            d3.json("/graph", function(error, graph) {
                var targs = {};
                var links = [];
                var nodes = {};
                console.log(graph)
                var f = true;
                var root =graph.Target.Name;
                var i = 0;
                function getRandomArbitrary(min, max) {
                    return Math.random() * (max - min) + min;
                }
                function addToGraph(g) {
                    for (var key in g.Edges) {
                        targs[g.Edges[key].Target.Name] = g.Edges[key].Target;




                        var link  = {};
                        var p = key.split(":")
                        link.source = p[1];
                        link.target =p[0];

                        link.type = p[0] == root ? "root": "import";


                        links.push(link);
                        addToGraph(g.Edges[key]);
                    }
                }

                addToGraph(graph)
                // Compute the distinct nodes from the links.
                links.forEach(function(link) {
                    link.source = nodes[link.source] || (nodes[link.source] = {name: link.source});
                    link.target = nodes[link.target] || (nodes[link.target] = {name: link.target});
                });


                function HSVtoRGB(h, s, v) {
                    var r, g, b, i, f, p, q, t;
                    if (arguments.length === 1) {
                        s = h.s, v = h.v, h = h.h;
                    }
                    i = Math.floor(h * 6);
                    f = h * 6 - i;
                    p = v * (1 - s);
                    q = v * (1 - f * s);
                    t = v * (1 - (1 - f) * s);
                    switch (i % 6) {
                        case 0: r = v, g = t, b = p; break;
                        case 1: r = q, g = v, b = p; break;
                        case 2: r = p, g = v, b = t; break;
                        case 3: r = p, g = q, b = v; break;
                        case 4: r = t, g = p, b = v; break;
                        case 5: r = v, g = p, b = q; break;
                    }
                    return {
                        r: Math.round(r * 255),
                        g: Math.round(g * 255),
                        b: Math.round(b * 255)
                    };
                }

                function componentToHex(c) {
                    var hex = c.toString(16);
                    return hex.length == 1 ? "0" + hex : hex;
                }

                function rgbToHex(r, g, b) {
                    return "#" + componentToHex(r) + componentToHex(g) + componentToHex(b);
                }

                function hashColor(d) {
                    var golden_ratio_conjugate = 0.618033988749895;
                    var shaObj = new jsSHA("SHA-1","TEXT");
                    shaObj.update(d.name);
                    var hash = shaObj.getHash("HEX");
                    var x =parseInt(hash.substring(0,2), 16);
                    x *= golden_ratio_conjugate;
                    x += golden_ratio_conjugate;
                    x  %= 1;

                    var t = HSVtoRGB(x,  0.5+(1* (d.weight/100)), 0.99)
                    var r ="#"+hash.substr(22,6);

                    return rgbToHex(t.r, t.g,t.b);
                }

                var width = window.innerWidth
                || document.documentElement.clientWidth
                || document.body.clientWidth;

                var height = window.innerHeight
                || document.documentElement.clientHeight
                || document.body.clientHeight;

                var force = d3.layout.force()
                .nodes(d3.values(nodes))
                .links(links)
                .size([width, height])
                .linkDistance(25)
                .charge(-200)
                .on("tick", tick)
                .start();

                var svg = d3.select("#depgraph").append("svg")
                .attr("width", width)
                .attr("height", height);


                svg.append("svg:defs").selectAll("marker")
                .data(["end"])
                .enter().append("svg:marker")
                .attr("id", String)
                .attr("viewBox", "0 -5 10 10")
                .attr("refX", 15)
                .attr("refY", -1.5)
                .attr("markerWidth", 6)
                .attr("markerHeight", 6)
                .attr("orient", "auto")
                .append("svg:path")
                .attr("d", "M0,-5L10,0L0,5");


                var path = svg.append("svg:g").selectAll("path")
                .data(force.links())
                .enter().append("svg:path")
                .attr("class", "link")
                .attr("id", function(d) {return d.source.name})
                .attr("stroke", function(d) {return hashColor(d.source)})
                .attr("marker-end", "url(#end)");


                var node = svg.selectAll(".node")
                .data(force.nodes())
                .enter().append("g")
                .attr("class", "node")
                .on({
                    "click": function(d) {

                        self.setState({target: targs[d.name]});
                    },
                })
                .call(force.drag);



                node.append("circle")
                .attr("fill", function(d) {return hashColor(d)})
                .attr("r", function(d) { return d.name == root? 10:8;});


                node.append("text")
                .attr("fill", function(d) {return hashColor(d); })
                //    .attr("fill", "#F3F315")
                .attr("x", 14)
                .attr("dy", ".35em")

                .text(function(d) { return d.name; });

                function tick() {
                    path.attr("d", function(d) {
                        var dx = d.target.x - d.source.x,
                        dy = d.target.y - d.source.y,
                        dr = Math.sqrt(dx * dx + dy * dy);
                        return "M" +
                        d.source.x + "," +
                        d.source.y + "A" +
                        dr + "," + dr + " 0 0,1 " +
                        d.target.x + "," +
                        d.target.y;
                    });

                    node
                    .attr("transform", function(d) {
                        return "translate(" + d.x + "," + d.y + ")"; });
                    }

                });
            },
            setInitialState: function() {
                return {};
            },
            render: function() {
                self = this;
                var size = function(obj) {
                    var size = 0, key;
                    for (key in obj) {
                        if (obj.hasOwnProperty(key)) size++;
                    }
                    return size;
                };
                if (this.state !== null){
                    console.log(this.state.target);
                    var x =[];
                    for (var key in this.state.target) {
                        if (typeof this.state.target[key] !== "string") {
                            if (size(this.state.target[key]) >0) {
                                x.push(<span id={key}><b>{key}:</b><StringArray elems={this.state.target[key]}></StringArray></span>);
                            }
                        }
                    }
                    return (
                        <div >

                            <div className="info-pane">
                                <pre>
                                    {this.state.target.Name}{'\n'}
                                    {x}
                                </pre>
                            </div>

                        </div>

                    );
                } else {
                    return <span />;
                }
            }
        });
