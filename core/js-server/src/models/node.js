import mongoose from "mongoose";

const node_schema = new mongoose.Schema({
    peer_id: { type: String, required: true },
    node_id: { type: String, required: true, unique: true },
});

const Nodes = mongoose.model("node", node_schema);

export { Nodes };