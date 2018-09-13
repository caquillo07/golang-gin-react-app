class Joke extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            liked: ""
        };
        this.like = this.like.bind(this);
    }

    like() {
        // ... we'll add this block later
    }

    render() {
        return (
            <div className="col-xs-4">
                <div className="panel panel-default">
                    <div className="panel-heading">#{this.props.joke.id} <span className="pull-right">{this.state.liked}</span></div>
                    <div className="panel-body">
                        {this.props.joke.joke}
                    </div>
                    <div className="panel-footer">
                        {this.props.joke.likes} Likes &nbsp;
                        <a onClick={this.like} className="btn btn-default">
                            <span className="glyphicon glyphicon-thumbs-up"></span>
                        </a>
                    </div>
                </div>
            </div>
        )
    }
}