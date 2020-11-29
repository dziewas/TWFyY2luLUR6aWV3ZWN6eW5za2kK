import React from "react";
import Response from "./Response";

function ResponseList(props) {
    let id = 0
    const responses = props.responses.length >= 5 ?
        props.responses.slice(props.responses.length - 5) :
        props.responses;
    responses.reverse()

    return (responses.map((response) => {
        id += 1;
        return <Response key={id} response={response} />
    }));
}

export default ResponseList