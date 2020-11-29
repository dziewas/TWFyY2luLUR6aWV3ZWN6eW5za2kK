import React, { createRef } from "react";

class NewTask extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            url: "",
            interval: null
        }

        this.formRef = createRef()
    }

    submitHandler = (event) => {
        event.preventDefault()

        let task = {
            url: event.target.url.value,
            interval: Number(event.target.interval.value)
        }

        if (task.url === "" || !Number(task.interval)) {
            console.log("invalid data submitted")
            return
        }

        this.props.onTaskCreated(task)
    }

    render() {
        return (<div>
            <form onSubmit={this.submitHandler} ref={this.formRef}>
                <p>url:</p>
                <input type="text" name="url" required />
                <p>interval:</p>
                <input type="text" name="interval" required />
                <input type="submit" value="Submit" />
            </form>
        </div>);
    }
}

export default NewTask
