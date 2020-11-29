import React from "react";
import NewTask from "../components/task/NewTask";
import TaskList from "../components/task/TaskList";
import ResponseList from "../components/response/ResponseList";
import classes from "./crawler.module.css";

const BASE_URL = "http://localhost:8080/api/fetcher"
const TASK_URL = "http://localhost:8080/api/fetcher/{id}"
const RESPONSES_URL = "http://localhost:8080/api/fetcher/{id}/history"

class Crawler extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      tasks: [],
      responses: []
    }
  }

  componentDidMount() {
    this.handleShouldRefreshTasks();
  }

  refreshTasks = (tasks) => {
    this.setState({ tasks: tasks })
  }

  refreshResponses = (responses) => {
    this.setState({ responses: responses })
  }

  handleNewTaskSubmitted = (task) => {
    this.createTask(task, this.handleShouldRefreshTasks);
  }

  handleTaskClicked = (task) => {
    this.fetchResponses(task, this.refreshResponses)
  }

  handleTaskDeleteClicked = (task) => {
    this.removeTask(task, this.handleShouldRefreshTasks)
  }

  handleShouldRefreshTasks = () => {
    this.fetchTasks(this.refreshTasks);
  }

  fetchResponses(task, callback = null) {
    const url = RESPONSES_URL.replace("{id}", task.id)

    fetch(url)
      .then((response) => response.json())
      .then((data) => {
        callback(data)
      })
      .catch(console.error)
  }

  fetchTasks(callback = null) {
    fetch(BASE_URL)
      .then((response) => response.json())
      .then((data) => {
        callback(data)
      })
      .catch(console.error)
  }

  createTask(task, callback = null) {
    const request = new Request(BASE_URL, { method: 'POST', body: JSON.stringify(task) });

    fetch(request)
      .then((response) => {
        response.status === 200 && callback()
      })
      .catch(console.error)
  }

  removeTask(task, callback = null) {
    const url = TASK_URL.replace("{id}", task.id)
    const request = new Request(url, { method: 'DELETE' });

    fetch(request)
      .then((response) => {
        response.status === 200 && callback()
      })
      .catch(console.error)
  }

  render() {
    return (
      <div className={classes.crawler}>
        <div className={classes.left}>
          <NewTask onTaskCreated={this.handleNewTaskSubmitted} />
          <h3>Tasks:</h3>
          <TaskList
            tasks={this.state.tasks}
            onTaskClicked={this.handleTaskClicked}
            onTaskDeleteClicked={this.handleTaskDeleteClicked}
          />
        </div>
        <div className={classes.right}>
          <h3>Responses:</h3>
          <ResponseList responses={this.state.responses} />
        </div>
      </div>
    )
  }
}

export default Crawler