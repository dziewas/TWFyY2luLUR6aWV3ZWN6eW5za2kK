import React from "react";
import classes from "./response.module.css"

function Response(props) {
    return (
        <div className={classes.note}>
            <h5 className={classes.response}>{props.response.response}</h5>
            <h6>created at: {props.response.created_at}, duration: {props.response.duration}</h6>
        </div>
    )
}

export default Response