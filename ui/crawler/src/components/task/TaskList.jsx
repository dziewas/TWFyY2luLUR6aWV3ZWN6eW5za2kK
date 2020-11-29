import React from "react";
import Task from "./Task";

function TaskList(props) {
    return (props.tasks.map((task) => {
        return <Task
            key={task.id}
            onTaskClicked={props.onTaskClicked}
            onTaskDeleteClicked={props.onTaskDeleteClicked}
            task={task}
        />
    }));
}

export default TaskList