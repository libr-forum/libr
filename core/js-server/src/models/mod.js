import mongoose from "mongoose";

const mod_schema = new mongoose.Schema({
    peer_id: { type: String, required: true },
    public_key: { type: String, required: true, unique: true },
});

const Mods = mongoose.model("mod", mod_schema); // if multiple databases then use db1.model and so on

export { Mods };