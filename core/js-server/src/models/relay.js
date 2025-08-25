import mongoose from "mongoose";

const relaySchema = new mongoose.Schema({
    address: { type: String, required: true }
});

const Relay = mongoose.model("Relay", relaySchema);

export default Relay;