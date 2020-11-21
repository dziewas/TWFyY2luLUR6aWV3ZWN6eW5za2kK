import React from "react";
import classes from "./task.module.css"

function Task(props) {
    return (
        <div>
            <div>
                <button type="button"
                    className={classes.linkButton}
                    onClick={() => props.onTaskClicked(props.task)}>
                    {props.task.url}
                </button>
                <button type="button"
                    className={classes.deleteButton}
                    onClick={() => props.onTaskDeleteClicked(props.task)}>
                    remove
                </button>
            </div>
            <h6>interval [s]: {props.task.interval}</h6>
        </div>
    )
}

export default Task